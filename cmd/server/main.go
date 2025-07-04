package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/google/gousb"
)

const systemdService = `[Unit]
Description=ESC/POS USB Printer Server
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s --port 8080 --vendor 0x04b8 --product 0x0e15
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

func installService() {
	// Check if running as root
	if os.Geteuid() != 0 {
		log.Fatal("install-service must be run as root (use sudo)")
	}

	// Get the user who ran sudo
	currentUser := os.Getenv("SUDO_USER")
	if currentUser == "" {
		log.Fatal("Could not determine the original user. Make sure to run with sudo.")
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	execPath, _ = filepath.EvalSymlinks(execPath)
	workingDir := filepath.Dir(execPath)

	// Blacklist usblp module
	blacklistContent := "# Blacklist usblp module to allow direct USB printer access\nblacklist usblp\n"
	if err := os.WriteFile("/etc/modprobe.d/blacklist-usblp.conf", []byte(blacklistContent), 0644); err != nil {
		log.Fatalf("Failed to write blacklist file: %v", err)
	}
	fmt.Println("Blacklisted usblp module")

	// Remove usblp if currently loaded
	exec.Command("rmmod", "usblp").Run() // Ignore errors

	serviceContent := fmt.Sprintf(systemdService, currentUser, workingDir, execPath)
	servicePath := "/etc/systemd/system/escpos-server.service"

	// Write the service file
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		log.Fatalf("Failed to write service file: %v", err)
	}
	fmt.Println("Systemd service file written to", servicePath)

	// ... rest of the systemctl commands
}

type PrinterServer struct {
	ctx      *gousb.Context
	dev      *gousb.Device
	intf     *gousb.Interface
	done     func()
	endpoint *gousb.OutEndpoint
	mu       sync.Mutex

	vendorID  uint16
	productID uint16
}

func NewPrinterServer(vendorID, productID uint16) (*PrinterServer, error) {
	ctx := gousb.NewContext()

	dev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(vendorID), gousb.ID(productID))
	if err != nil {
		ctx.Close()
		return nil, fmt.Errorf("failed to open device: %w", err)
	}

	intf, done, err := dev.DefaultInterface()
	if err != nil {
		dev.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to claim interface: %w", err)
	}

	ep, err := intf.OutEndpoint(1)
	if err != nil {
		done()
		dev.Close()
		ctx.Close()
		return nil, fmt.Errorf("failed to open endpoint: %w", err)
	}

	return &PrinterServer{
		ctx:       ctx,
		dev:       dev,
		intf:      intf,
		done:      done,
		endpoint:  ep,
		vendorID:  vendorID,
		productID: productID,
	}, nil
}

func (ps *PrinterServer) Close() {
	if ps.done != nil {
		ps.done()
	}
	if ps.dev != nil {
		ps.dev.Close()
	}
	if ps.ctx != nil {
		ps.ctx.Close()
	}
}

func (ps *PrinterServer) handlePrint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the raw data from the request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Lock to ensure thread-safe access to the USB device
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Write the data to the printer
	_, err = ps.endpoint.Write(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to printer: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully sent %d bytes to printer\n", len(data))
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "install-service" {
		installService()
	}
	var (
		port      = flag.String("port", "8080", "HTTP server port")
		vendorID  = flag.Uint("vendor", 0x04b8, "USB vendor ID")
		productID = flag.Uint("product", 0x0e15, "USB product ID")
	)
	flag.Parse()

	// Initialize the printer server
	ps, err := NewPrinterServer(uint16(*vendorID), uint16(*productID))
	if err != nil {
		log.Fatalf("Failed to initialize printer: %v", err)
	}
	defer ps.Close()

	// Set up HTTP routes
	http.HandleFunc("/print", ps.handlePrint)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	log.Printf("Starting server on port %s", *port)
	log.Printf("Connected to printer (VID: 0x%04x, PID: 0x%04x)", *vendorID, *productID)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

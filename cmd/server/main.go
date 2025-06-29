package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/google/gousb"
)

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

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/google/gousb"
	"github.com/petertjmills/escpos-server/escpos"
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

func (ps *PrinterServer) Write(b []byte) (n int, err error) {
	// Lock to ensure thread-safe access to the USB device
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Write the data to the printer
	n, err = ps.endpoint.Write(b)
	if err != nil {
		return 0, fmt.Errorf("failed to write to printer: %w", err)
	}

	return n, nil
}

func main() {
	var (
		vendorID  = flag.Uint("vendor", 0x04b8, "USB vendor ID")
		productID = flag.Uint("product", 0x0e15, "USB product ID")
		// markdown  = flag.String("markdown", "", "Print receipt from markdown")
		debug = flag.Bool("debug", false, "Debug mode - print raw commands instead of sending to server")
	)
	flag.Parse()

	// 2. Determine the input source
	var input []byte
	var err error

	// If there are remaining arguments after flags, use the last one as input
	args := flag.Args()
	if len(args) > 0 {
		input = []byte(args[0])
	} else {
		// Otherwise, read from Stdin (piped or redirected)
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize the printer server
	ps, err := NewPrinterServer(uint16(*vendorID), uint16(*productID))
	if err != nil {
		log.Fatalf("Failed to initialize printer: %v", err)
	}
	defer ps.Close()

	var writer io.Writer
	var debugWriter *DebugWriter

	if *debug {
		// Use debug writer
		debugWriter = &DebugWriter{}
		writer = debugWriter
	} else {
		writer = ps
	}

	// Create ESC/POS printer instance
	p := escpos.New(writer)
	p.SetConfig(escpos.ConfigEpsonTMT20II)

	p.WriteMarkdown(input)
	p.LineFeed()
	if _, err := p.Cut(); err != nil {
		log.Fatalf("Failed to cut: %v", err)
	}
	if err := p.Print(); err != nil {
		log.Fatalf("Failed to print: %v", err)
	}

	if debugWriter != nil {
		// Print debug output
		fmt.Println("=== DEBUG OUTPUT ===")
		fmt.Println(debugWriter.PrettyPrint())
		fmt.Printf("\nTotal bytes: %d\n", len(debugWriter.Bytes()))
	}

}

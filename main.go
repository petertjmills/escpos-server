package main

import (
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	"log"
	"os"

	"github.com/google/gousb"
	"github.com/petertjmills/escpos-server/escpos"
)

func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func main() {
	ctx := gousb.NewContext()
	defer ctx.Close()

	dev, err := ctx.OpenDeviceWithVIDPID(0x04b8, 0x0e15)
	if err != nil {
		fmt.Println("Error opening device:", err)
		return
	}
	defer dev.Close()

	// Claim the default interface using a convenience function.
	// The default interface is always #0 alt #0 in the currently active
	// config.
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()

	// Open an OUT endpoint.
	ep, err := intf.OutEndpoint(1)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(1): %v", intf, err)
	}

	p := escpos.New(ep)
	p.SetConfig(escpos.ConfigEpsonTMT20II)

	for _, i := range []uint8{0, 1, 2, 3, 4, 5} {
		p.Size(i, i).Write("I LOVE YOU!!")
		p.LineFeed()
	}

	p.PrintAndCut()
}

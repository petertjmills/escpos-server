package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/petertjmills/escpos-server/escpos"
)

func printTextStyles(p *escpos.Escpos) error {
	// Text style examples
	p.Justify(escpos.JustifyCenter)
	p.Size(2, 2).Write("Text Style Examples\n")
	p.Size(1, 1)
	p.LineFeed()

	p.Justify(escpos.JustifyLeft)

	// Bold
	p.Bold(true)
	p.Write("This is bold text\n")
	p.Bold(false)

	// Underline
	p.Underline(1)
	p.Write("This is underlined text\n")
	p.Underline(0)
	p.Underline(2)
	p.Write("This is thick underlined text\n")
	p.Underline(0)

	// Reverse
	p.Reverse(true)
	p.Write("This is reverse text\n")
	p.Reverse(false)

	// Different sizes
	for i := uint8(1); i <= 5; i++ {
		p.Size(i, i).Write(fmt.Sprintf("Size %d x %d\n", i, i))
	}
	p.Size(1, 1)

	// Rotation
	p.Rotate(true)
	p.Write("This is rotated text\n")
	p.Rotate(false)

	// Upside down
	p.UpsideDown(true)
	p.Write("This is upside down\n")
	p.UpsideDown(false)

	p.LineFeed()
	return nil
}

func printBarcodes(p *escpos.Escpos) error {
	p.Justify(escpos.JustifyCenter)
	p.Size(2, 2).Write("Barcode Examples\n")
	p.Size(1, 1)
	p.LineFeed()

	// UPC-A
	p.Write("UPC-A:\n")
	p.HRIPosition(2) // Below barcode
	p.UPCA("12345678901")
	p.LineFeed()

	// EAN-13
	p.Write("EAN-13:\n")
	p.EAN13("1234567890123")
	p.LineFeed()

	// EAN-8
	p.Write("EAN-8:\n")
	p.EAN8("12345678")
	p.LineFeed()

	// QR Codes
	p.Write("QR Code (Small):\n")
	p.QRCode("Hello, World!", false, 4, escpos.QRCodeErrorCorrectionLevelL)
	p.LineFeed()

	p.Write("QR Code (Large):\n")
	p.QRCode("https://github.com", true, 8, escpos.QRCodeErrorCorrectionLevelH)
	p.LineFeed()

	return nil
}

func printImage(p *escpos.Escpos, imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	p.Justify(escpos.JustifyCenter)
	p.Size(2, 2).Write("Image Example\n")
	p.Size(1, 1)
	p.LineFeed()

	if _, err := p.PrintImage(img); err != nil {
		return fmt.Errorf("failed to print image: %w", err)
	}

	p.LineFeed()
	return nil
}

func printTable(p *escpos.Escpos) error {
	p.Justify(escpos.JustifyCenter)
	p.Size(2, 2).Write("Table Example\n")
	p.Size(1, 1)
	p.LineFeed()

	p.Justify(escpos.JustifyLeft)

	// Header
	p.Bold(true)
	p.Write("Item              Qty   Price\n")
	p.Write("------------------------------\n")
	p.Bold(false)

	// Items
	items := []struct {
		name  string
		qty   int
		price float64
	}{
		{"Coffee", 2, 3.50},
		{"Sandwich", 1, 8.00},
		{"Cookie", 3, 2.50},
		{"Juice", 1, 4.00},
	}

	total := 0.0
	for _, item := range items {
		line := fmt.Sprintf("%-16s %3d   $%5.2f\n", item.name, item.qty, item.price*float64(item.qty))
		p.Write(line)
		total += item.price * float64(item.qty)
	}

	p.Write("------------------------------\n")
	p.Bold(true)
	p.Write(fmt.Sprintf("%-20s  $%5.2f\n", "Total:", total))
	p.Bold(false)

	return nil
}

func printInternational(p *escpos.Escpos) error {
	p.Justify(escpos.JustifyCenter)
	p.Size(2, 2).Write("International Text\n")
	p.Size(1, 1)
	p.LineFeed()

	p.Justify(escpos.JustifyLeft)

	// Example with different encodings
	p.Write("ASCII: Hello World!\n")
	p.WriteGBK("Chinese: 你好世界\n")
	p.WriteWEU("Spanish: ¡Hola Mundo! ñ á é í ó ú\n")

	p.LineFeed()
	return nil
}

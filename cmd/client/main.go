package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	dailyFns "github.com/petertjmills/escpos-server/daily"
	"github.com/petertjmills/escpos-server/escpos"
)

// HTTPWriter collects data and sends it to the HTTP server
type HTTPWriter struct {
	serverURL string
	buffer    bytes.Buffer
}

func NewHTTPWriter(serverURL string) *HTTPWriter {
	return &HTTPWriter{
		serverURL: serverURL,
	}
}

func (hw *HTTPWriter) Write(p []byte) (n int, err error) {
	return hw.buffer.Write(p)
}

func (hw *HTTPWriter) Flush() error {
	if hw.buffer.Len() == 0 {
		return nil
	}

	resp, err := http.Post(hw.serverURL+"/print", "application/octet-stream", &hw.buffer)
	if err != nil {
		return fmt.Errorf("failed to send data to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s - %s", resp.Status, string(body))
	}

	// Clear the buffer after successful send
	hw.buffer.Reset()
	return nil
}

// DebugWriter captures all data written to it
type DebugWriter struct {
	buffer bytes.Buffer
}

func (dw *DebugWriter) Write(p []byte) (n int, err error) {
	return dw.buffer.Write(p)
}

func (dw *DebugWriter) String() string {
	return dw.buffer.String()
}

func (dw *DebugWriter) Bytes() []byte {
	return dw.buffer.Bytes()
}

func (dw *DebugWriter) HexDump() string {
	return hex.Dump(dw.buffer.Bytes())
}

func (dw *DebugWriter) HexString() string {
	return hex.EncodeToString(dw.buffer.Bytes())
}

// PrettyPrint returns a formatted view showing both hex and ASCII
func (dw *DebugWriter) PrettyPrint() string {
	data := dw.buffer.Bytes()
	var result strings.Builder

	result.WriteString("Raw bytes (hex): ")
	result.WriteString(dw.HexString())
	result.WriteString("\n\nHex dump:\n")
	result.WriteString(dw.HexDump())
	result.WriteString("\nPrintable ASCII:\n")

	for _, b := range data {
		if b >= 32 && b <= 126 { // Printable ASCII range
			result.WriteByte(b)
		} else if b == '\n' {
			result.WriteString("\\n\n")
		} else if b == '\r' {
			result.WriteString("\\r")
		} else if b == '\t' {
			result.WriteString("\\t")
		} else {
			result.WriteString(fmt.Sprintf("\\x%02x", b))
		}
	}

	return result.String()
}

func main() {
	var (
		serverURL = flag.String("server", "http://localhost:8080", "Server URL")
		text      = flag.String("text", "", "Text to print")
		markdown  = flag.String("markdown", "", "Print receipt from markdown")
		daily     = flag.Bool("daily", false, "Print daily receipt")
		debug     = flag.Bool("debug", false, "Debug mode - print raw commands instead of sending to server")
	)
	flag.Parse()

	var writer io.Writer
	var debugWriter *DebugWriter

	if *debug {
		// Use debug writer
		debugWriter = &DebugWriter{}
		writer = debugWriter
	} else {
		// Use HTTP writer
		writer = NewHTTPWriter(*serverURL)
	}

	// Create ESC/POS printer instance
	p := escpos.New(writer)
	p.SetConfig(escpos.ConfigEpsonTMT20II)

	if *text != "" {
		p.Write(*text)
		p.LineFeed()
		if _, err := p.Cut(); err != nil {
			log.Fatalf("Failed to cut: %v", err)
		}
		if err := p.Print(); err != nil {
			log.Fatalf("Failed to print: %v", err)
		}

	} else if *markdown != "" {
		// load markdown file
		data, err := os.ReadFile(*markdown)
		if err != nil {
			log.Fatalf("Failed to read markdown file: %v", err)
		}
		p.WriteMarkdown(data)
		p.LineFeed()
		if _, err := p.Cut(); err != nil {
			log.Fatalf("Failed to cut: %v", err)
		}
		if err := p.Print(); err != nil {
			log.Fatalf("Failed to print: %v", err)
		}
	} else if *daily {
		// open .env file if it's there and set os env vars
		if err := godotenv.Load(); err != nil {
			fmt.Println("Error loading .env file:", err)
		}
		// print current day like monday 25th July 2025
		p.WriteMarkdown(fmt.Appendf([]byte("### "), time.Now().Format("Monday, 2 January 2006")))

		dailyWotd, _ := dailyFns.GetWordOfTheDay()
		p.WriteMarkdown([]byte(dailyWotd))
		dailyWeather := dailyFns.GetWeatherMD()
		p.WriteMarkdown([]byte(dailyWeather))
		dailyNews, _ := dailyFns.GetNews()
		p.WriteMarkdown([]byte(dailyNews))
		pollen, _ := dailyFns.GetPollenCount()
		p.WriteMarkdown([]byte(pollen))
		hn, _ := dailyFns.GetHackerNewsFront()
		p.WriteMarkdown([]byte(hn))

		p.LineFeed()
		if _, err := p.Cut(); err != nil {
			log.Fatalf("Failed to cut: %v", err)
		}
		if err := p.Print(); err != nil {
			log.Fatalf("Failed to print: %v", err)
		}
	} else {
		log.Fatal("Please specify either -demo, -text, or -markdown")
	}

	if *debug {
		// Print debug output
		fmt.Println("=== DEBUG OUTPUT ===")
		fmt.Println(debugWriter.PrettyPrint())
		fmt.Printf("\nTotal bytes: %d\n", len(debugWriter.Bytes()))
	} else {
		// Send data to server
		if httpWriter, ok := writer.(*HTTPWriter); ok {
			if err := httpWriter.Flush(); err != nil {
				log.Fatalf("Failed to send data to server: %v", err)
			}
			fmt.Println("Successfully sent print job to server")
		}
	}
}

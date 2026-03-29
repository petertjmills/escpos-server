package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

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

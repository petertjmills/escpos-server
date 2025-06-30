# ESC/POS Server

A client-server architecture for ESC/POS thermal printers. The server handles USB communication while clients generate ESC/POS commands.

# important commands

```sh
peter@raspberrypi:~ $ sudo rmmod usblp
peter@raspberrypi:~ $ sudo ./go/bin/server
```

## Architecture

- **Server**: HTTP server that receives raw ESC/POS commands and forwards them to the USB printer
- **Client**: Generates ESC/POS commands using the escpos package and sends them to the server

## Installation

```bash
go build -o escpos-server ./cmd/server
go build -o escpos-client ./cmd/client
```

## Usage

### Starting the Server

```bash
# Start server with default settings (port 8080, Epson TM-T20II)
./escpos-server

# Custom port and USB device
./escpos-server -port 9090 -vendor 0x04b8 -product 0x0e15
```

### Using the Client

```bash
# Print simple text
./escpos-client -server http://localhost:8080 -text "Hello, World!"

# Print demo receipt
./escpos-client -server http://localhost:8080 -demo

# Print to remote server
./escpos-client -server http://printer.local:8080 -text "Remote printing!"
```

## API

The server exposes the following endpoints:

- `POST /print` - Send raw ESC/POS commands
  - Body: Binary data (application/octet-stream)
  - Response: 200 OK on success

- `GET /health` - Health check endpoint
  - Response: 200 OK

## Client Examples

The client supports various ESC/POS features:

- Text formatting (bold, underline, size, rotation)
- Barcodes (UPC-A, EAN-13, EAN-8)
- QR codes
- Images
- International character sets (GBK, Western European)

## Custom Client

You can create custom clients in any language. Simply send raw ESC/POS commands to the server:

```python
import requests

# Example in Python
commands = b'\x1b@'  # Initialize printer
commands += b'Hello from Python!\n'
commands += b'\x1dV\x41\x00'  # Cut paper

response = requests.post('http://localhost:8080/print',
                        data=commands,
                        headers={'Content-Type': 'application/octet-stream'})
```

## Benefits

1. **Network Printing**: Print from any device on the network
2. **Language Agnostic**: Write clients in any language
3. **Centralized USB Management**: Server handles all USB communication
4. **Multiple Clients**: Support multiple clients sending to the same printer
5. **Remote Printing**: Print from anywhere with internet access


## Outputs


```md
# Monday 31st March 2025

## Word of the Day
Today's Word: "Innovation"
Today's Definition: "The process of creating something new or better."

# Weather
## High
  25°C
  Feels Like: 23°C
## Low
  18°C
  Feels Like: 17°C
## Humidity
  60%
## Wind
  10 km/h
## MSLP
  1013 hPa
## UV Index
  5
## Rain
10:00 50%
15:00 70%

# Today

## Calendar

## Todos

# News

# HackerNews

```

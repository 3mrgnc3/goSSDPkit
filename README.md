# goSSDPkit - Go SSDP Security Testing Kit

A Go implementation of SSDP-based security research and phishing attack tools for authorized penetration testing.

## Overview

This tool responds to SSDP multicast discover requests, posing as a generic UPNP device. Your spoofed device will magically appear in Windows Explorer on machines in your local network. Users who are tempted to open the device are shown a configurable phishing page. This page can load a hidden image over SMB, allowing you to capture or relay the NetNTLM challenge/response.

Templates are also provided to capture clear-text credentials via basic authentication and logon forms, and creating your own custom templates is quick and easy.

## Features

- **SSDP Multicast Listener**: Responds to SSDP discovery requests
- **HTTP Server**: Serves phishing pages and device XML descriptors  
- **Template System**: Configurable phishing templates
- **Credential Harvesting**: Captures NetNTLM hashes and clear-text credentials
- **XXE Detection**: Identifies potential XML External Entity vulnerabilities
- **Basic Authentication**: Optional credential prompt with realm configuration
- **Logging**: Comprehensive logging of all interactions

## Installation

### Prerequisites

- Go 1.21 or later
- Root/administrator privileges (required for binding to privileged ports)

### Building

```bash
# Clone the repository
git clone https://github.com/3mrgnc3/goSSDPkit
cd goSSDPkit

# Install dependencies
make deps

# Build the application
make build
```

### Cross-Platform Releases

To build release binaries for multiple platforms:

```bash
# Build for all supported platforms
./build_releases.sh v1.0.0

# Quick build for key platforms only  
./simple_build.sh v1.0.0
```

**Supported Platforms:**
- Linux: amd64, 386, arm64, armv7 (arm/7), mips, mipsle, mips64, mips64le, ppc64, ppc64le, s390x, riscv64
- Windows: amd64, 386, arm64
- macOS: amd64, arm64 (Apple Silicon) 
- FreeBSD: amd64, 386, arm64
- OpenBSD: amd64, 386
- NetBSD: amd64, 386
- Solaris: amd64
- Android: arm64, arm

Builds include embedded version information accessible via `-version` flag.

## Usage

### Basic Usage

```bash
# Basic run with default office365 template
sudo ./build/goSSDPkit eth0

# Use a specific template with redirect
sudo ./build/goSSDPkit wlan0 -t microsoft-azure -u https://azure.microsoft.com

# Enable basic authentication with custom realm
sudo ./build/goSSDPkit eth0 -t office365 -b -r "Corporate Portal"

# Run in analyze mode (no SSDP responses, testing only)
sudo ./build/goSSDPkit eth0 -a
```

### Command Line Options

```
Usage: goSSDPkit [options] <interface>

positional arguments:
  interface             Network interface to listen on

optional arguments:
  -p int                Port for HTTP server (default 8888)
  -t string             Name of a folder in the templates directory (default "office365")
  -s string             IP address of your SMB server (defaults to interface IP)
  -b                    Enable basic authentication and log credentials
  -r string             Realm for basic authentication (default "Microsoft Corporation")
  -u string             URL to redirect to after capturing credentials
  -a                    Run in analyze mode (no SSDP responses)
```

### Examples

```bash
# Capture NetNTLM hashes with Office365 template
sudo ./build/goSSDPkit eth0 -t office365 -u https://office.microsoft.com

# Use custom SMB server
sudo ./build/goSSDPkit wlan0 -t scanner -s 192.168.1.205

# Look for XXE vulnerabilities
sudo ./build/goSSDPkit eth0 -t xxe-smb
```

## Templates

The following templates are included:

- **office365**: Office365 login page for credential harvesting
- **scanner**: Corporate scanner with "new scans waiting" message
- **microsoft-azure**: Microsoft Azure login portal
- **bitcoin**: Bitcoin wallet interface
- **password-vault**: IT password vault interface
- **xxe-smb**: XXE vulnerability detection with SMB callback
- **xxe-exfil**: XXE vulnerability with file exfiltration attempt

### Creating Custom Templates

Each template directory must contain:

- `device.xml`: UPnP device descriptor (defines Windows Explorer appearance)
- `present.html`: Phishing page with template variables
- `service.xml`: UPnP service descriptor (optional)

Template variables available in HTML files:
- `{{.SMBServer}}`: SMB server IP for NetNTLM capture
- `{{.LocalIP}}`: Local server IP address
- `{{.LocalPort}}`: Local server port
- `{{.SessionUSN}}`: Unique session identifier
- `{{.RedirectURL}}`: Redirect URL after credential capture

## Project Structure

```
goSSDPkit/
├── cmd/goSSDPkit/        # Main application
├── pkg/
│   ├── ssdp/            # SSDP multicast listener
│   ├── upnp/            # HTTP server for UPnP/phishing
│   └── template/        # Template processing engine
├── templates/           # Phishing templates
├── reference_projects/  # Original Python and Go SSDP references
└── build/              # Compiled binaries
```

## Development

### Building and Testing

```bash
# Install dependencies
make deps

# Run tests
make test

# Check code quality
make check

# Build for multiple platforms
make build-all
```

### Reference Material

This project includes reference implementations in `reference_projects/`:
- `evil-ssdp/`: Original Python implementation
- `go-ssdp/`: Go SSDP library for networking patterns

## Security Considerations

This tool is designed for authorized security testing and research purposes only. Users are responsible for:

- Obtaining proper authorization before testing
- Complying with applicable laws and regulations
- Using the tool ethically and responsibly

## Logging

All significant events are logged to `logs/goSSDPkit.log`:
- Captured credentials (both basic auth and form submissions)
- XXE vulnerability detections
- Exfiltration attempts

## License

This project maintains compatibility with the original evil-ssdp license terms.

## Credits

- Original evil-ssdp by initstring (https://github.com/initstring/evil-ssdp)
- Go SSDP library by koron (https://github.com/koron/go-ssdp)  
- Go port implementation by 3mrgnc3 (https://github.com/3mrgnc3)
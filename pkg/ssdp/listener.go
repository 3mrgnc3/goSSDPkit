package ssdp

import (
	"fmt"
	"net"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/ipv4"
)

// Colors for console output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[91m"
	ColorGreen  = "\033[92m"
	ColorYellow = "\033[93m"
	ColorBlue   = "\033[94m"
)

// Console output prefixes
var (
	OkBox      = ColorBlue + "[*] " + ColorReset
	NoteBox    = ColorGreen + "[+] " + ColorReset
	WarnBox    = ColorYellow + "[!] " + ColorReset
	MSearchBox = ColorBlue + "[M-SEARCH]     " + ColorReset
	XMLBox     = ColorGreen + "[XML REQUEST]  " + ColorReset
	PhishBox   = ColorRed + "[PHISH HOOKED] " + ColorReset
	CredsBox   = ColorRed + "[CREDS GIVEN]  " + ColorReset
	XXEBox     = ColorRed + "[XXE VULN!!!!] " + ColorReset
	ExfilBox   = ColorRed + "[EXFILTRATION] " + ColorReset
	DetectBox  = ColorYellow + "[DETECTION]    " + ColorReset
)

// Listener represents an SSDP multicast listener
type Listener struct {
	sock         *net.UDPConn
	knownHosts   map[string]bool
	localIP      string
	localPort    int
	analyzeMode  bool
	sessionUSN   string
	validST      *regexp.Regexp
	mu           sync.RWMutex
}

// NewListener creates a new SSDP listener
func NewListener(localIP string, localPort int, analyzeMode bool) (*Listener, error) {
	// SSDP multicast address and port as defined by the spec
	ssdpPort := 1900
	mcastGroup := "239.255.255.250"
	
	// Create UDP address for multicast group
	mcastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", mcastGroup, ssdpPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}
	
	// Create listener address (bind to all interfaces on SSDP port)
	listenAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", ssdpPort))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve listen address: %w", err)
	}
	
	// Create UDP connection
	conn, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}
	
	// Get the interface for the local IP
	iface, err := getInterfaceByIP(localIP)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get interface for IP %s: %w", localIP, err)
	}
	
	// Create IPv4 packet connection for multicast operations
	pconn := ipv4.NewPacketConn(conn)
	
	// Join multicast group on the specific interface
	if err := pconn.JoinGroup(iface, mcastAddr); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to join multicast group on interface %s: %w", iface.Name, err)
	}
	
	// Set control message to receive destination info (not supported on Windows)
	if runtime.GOOS != "windows" {
		if err := pconn.SetControlMessage(ipv4.FlagDst, true); err != nil {
			fmt.Printf("%sWarning: failed to set control message (non-fatal): %v\n", WarnBox, err)
		}
	}
	
	// Enable SO_REUSEADDR to allow multiple processes to bind to same port
	if err := conn.SetReadBuffer(65536); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set read buffer: %w", err)
	}
	
	fmt.Printf("%sSSDP listener bound to interface %s (%s) on port %d\n", 
		OkBox, iface.Name, localIP, ssdpPort)
	
	// Regex for validating ST headers (same pattern as Python version)
	validST := regexp.MustCompile(`^[a-zA-Z0-9.\-_]+:[a-zA-Z0-9.\-_:]+$`)
	
	return &Listener{
		sock:        conn,
		knownHosts:  make(map[string]bool),
		localIP:     localIP,
		localPort:   localPort,
		analyzeMode: analyzeMode,
		sessionUSN:  generateSessionUSN(),
		validST:     validST,
	}, nil
}

// generateSessionUSN creates a random USN for this session
func generateSessionUSN() string {
	return fmt.Sprintf("uuid:%s-%s-%s-%s-%s",
		genRandom(8), genRandom(4), genRandom(4), genRandom(4), genRandom(12))
}

// genRandom generates a random hex string of specified length
func genRandom(length int) string {
	const chars = "abcdef0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

// getInterfaceByIP finds the network interface for the given IP address
func getInterfaceByIP(targetIP string) (*net.Interface, error) {
	// Special case for loopback - try different names based on OS
	if targetIP == "127.0.0.1" {
		// Try common loopback interface names
		loopbackNames := []string{"lo", "lo0", "Loopback Pseudo-Interface 1"}
		for _, name := range loopbackNames {
			if iface, err := net.InterfaceByName(name); err == nil {
				return iface, nil
			}
		}
		// If none found, fall through to search by IP
	}
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ipNet.IP.String() == targetIP {
					return &iface, nil
				}
			}
		}
	}
	
	return nil, fmt.Errorf("interface not found for IP %s", targetIP)
}

// SendLocation sends an SSDP response to the requester
func (l *Listener) SendLocation(addr net.Addr, requestedST string) error {
	url := fmt.Sprintf("http://%s:%d/ssdp/device-desc.xml", l.localIP, l.localPort)
	dateFormat := time.Now().UTC().Format(time.RFC1123)
	
	ssdpReply := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"CACHE-CONTROL: max-age=1800\r\n"+
		"DATE: %s\r\n"+
		"EXT:\r\n"+
		"LOCATION: %s\r\n"+
		"OPT: \"http://schemas.upnp.org/upnp/1/0/\"; ns=01\r\n"+
		"01-NLS: %s\r\n"+
		"SERVER: UPnP/1.0\r\n"+
		"ST: %s\r\n"+
		"USN: %s::%s\r\n"+
		"BOOTID.UPNP.ORG: 0\r\n"+
		"CONFIGID.UPNP.ORG: 1\r\n"+
		"\r\n\r\n",
		dateFormat, url, l.sessionUSN, requestedST, l.sessionUSN, requestedST)
	
	_, err := l.sock.WriteTo([]byte(ssdpReply), addr)
	return err
}

// ProcessData processes received SSDP data
func (l *Listener) ProcessData(data []byte, addr net.Addr) {
	remoteIP := strings.Split(addr.String(), ":")[0]
	dataStr := string(data)
	
	// Look for ST header in M-SEARCH request
	re := regexp.MustCompile(`(?i)\r\nST:(.*?)\r\n`)
	matches := re.FindStringSubmatch(dataStr)
	
	if strings.Contains(dataStr, "M-SEARCH") && len(matches) > 1 {
		requestedST := strings.TrimSpace(matches[1])
		
		if l.validST.MatchString(requestedST) {
			// Create unique key for this host/ST combination
			hostKey := fmt.Sprintf("%s_%s", remoteIP, requestedST)
			
			l.mu.Lock()
			if !l.knownHosts[hostKey] {
				fmt.Printf("%sNew Host %s, Service Type: %s\n", 
					MSearchBox, remoteIP, requestedST)
				l.knownHosts[hostKey] = true
			}
			l.mu.Unlock()
			
			// Send response if not in analyze mode
			if !l.analyzeMode {
				if err := l.SendLocation(addr, requestedST); err != nil {
					fmt.Printf("%sError sending SSDP response: %v\n", WarnBox, err)
				}
			}
		} else {
			fmt.Printf("%sOdd ST (%s) from %s. Possible detection tool!\n", 
				DetectBox, requestedST, remoteIP)
		}
	}
}

// Listen starts listening for SSDP multicast messages
func (l *Listener) Listen() error {
	buffer := make([]byte, 1024)
	
	fmt.Printf("%sSSDP listener started, waiting for M-SEARCH requests...\n", OkBox)
	
	for {
		n, addr, err := l.sock.ReadFromUDP(buffer)
		if err != nil {
			return fmt.Errorf("error reading UDP data: %w", err)
		}
		
		// Debug: log all received UDP packets
		dataStr := string(buffer[:n])
		if strings.Contains(dataStr, "M-SEARCH") {
			fmt.Printf("%sReceived M-SEARCH from %s (length: %d)\n", NoteBox, addr.String(), n)
		}
		
		// Process the received data
		l.ProcessData(buffer[:n], addr)
	}
}

// Close closes the SSDP listener
func (l *Listener) Close() error {
	return l.sock.Close()
}

// GetSessionUSN returns the session USN
func (l *Listener) GetSessionUSN() string {
	return l.sessionUSN
}
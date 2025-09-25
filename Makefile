BINARY_NAME=goSSDPkit
BUILD_DIR=build
MAIN_PACKAGE=./cmd/goSSDPkit

.PHONY: build clean test run install deps help

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Clean build artifacts
clean:
		@echo "Cleaning build artifacts and logs..."
	@rm -rf build/ releases/ logs/

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application with default settings
run: build
	@echo "Running $(BINARY_NAME) with eth0 interface..."
	sudo ./$(BUILD_DIR)/$(BINARY_NAME) eth0

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Check for Go modules that can be updated
check-updates:
	@echo "Checking for available updates..."
	go list -m -u all

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run Go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run staticcheck (requires staticcheck to be installed)
staticcheck:
	@echo "Running staticcheck..."
	staticcheck ./...

# Run all checks
check: fmt vet test

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Show available templates
show-templates:
	@echo "Available templates:"
	@ls -1 templates/

# Help target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  run          - Build and run with eth0 interface"
	@echo "  deps         - Install/update dependencies"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  check        - Run fmt, vet, and tests"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  show-templates - List available templates"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Example usage:"
	@echo "  make build"
	@echo "  sudo ./build/goSSDPkit eth0 -t office365"
	@echo "  sudo ./build/goSSDPkit wlan0 -t microsoft-azure -u https://azure.microsoft.com"
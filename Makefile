# Makefile for nextcloud_monitor Go project

# Variables
BINARY_NAME=nc_monitor
SOURCE_FILE=nextcloud_monitor.go
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
GO_VERSION=$(shell go version | cut -d' ' -f3)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.version=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Default target
.PHONY: all
all: build

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build the binary
build: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: $(BUILD_DIR)
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SOURCE_FILE)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SOURCE_FILE)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SOURCE_FILE)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -rf $(BUILD_DIR)

# Install the binary to system path
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete"

# Uninstall the binary from system path
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with verbose output
.PHONY: run-verbose
run-verbose: build
	@echo "Running $(BINARY_NAME) in verbose mode..."
	NEXTCLOUD_VERBOSE=true ./$(BUILD_DIR)/$(BINARY_NAME)

# Format Go code
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run Go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run tests (if any exist)
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Download and tidy dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	go list -json -m all | nancy sleuth

# Show project info
.PHONY: info
info:
	@echo "Project Information:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Go Version:  $(GO_VERSION)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Build Time:  $(BUILD_TIME)"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install binary to $(INSTALL_DIR)"
	@echo "  uninstall    - Remove binary from $(INSTALL_DIR)"
	@echo "  run          - Build and run the application"
	@echo "  run-verbose  - Build and run with verbose output"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  test         - Run tests"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  security     - Check for security vulnerabilities"
	@echo "  info         - Show project information"
	@echo "  help         - Show this help message"

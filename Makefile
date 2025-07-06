# Makefile for Peaks - Beautiful Terminal Bandwidth Monitor

# Application info
APP_NAME=peaks
VERSION?=1.0.0

# Build variables
BINARY_NAME=$(APP_NAME)
BINARY_UNIX=$(APP_NAME)_unix
BINARY_WINDOWS=$(APP_NAME).exe
BINARY_DARWIN=$(APP_NAME)_darwin

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
GOFLAGS=-trimpath

.PHONY: build build-all build-linux build-windows build-darwin run test clean deps fmt lint help

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for all platforms"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install/update dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter (requires golangci-lint)"
	@echo "  help         - Show this help message"

# Build the project
build:
	go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/peaks

# Build for multiple platforms
build-all: build-linux build-windows build-darwin

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_UNIX) ./cmd/peaks

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_WINDOWS) ./cmd/peaks

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_DARWIN) ./cmd/peaks

# Run the application
run:
	go run ./cmd/peaks

# Test the application
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME) $(BINARY_UNIX) $(BINARY_WINDOWS) $(BINARY_DARWIN)
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod verify

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install the binary
install:
	go install

# Development mode with auto-reload
dev:
	go run .

.PHONY: build build-all build-linux build-windows build-darwin run test clean deps fmt lint install dev

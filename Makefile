# Makefile for Peaks bandwidth monitor

# Build variables
BINARY_NAME=peaks
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Build the project
build:
	go build -o $(BINARY_NAME) -v

# Build for multiple platforms
build-all: build-linux build-windows build-darwin

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_UNIX) -v

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BINARY_WINDOWS) -v

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DARWIN) -v

# Run the application
run:
	go run .

# Test the application
test:
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_DARWIN)

# Install dependencies
deps:
	go mod tidy
	go mod verify

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install the binary
install:
	go install

# Development mode with auto-reload
dev:
	go run .

.PHONY: build build-all build-linux build-windows build-darwin run test clean deps fmt lint install dev

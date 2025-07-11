# Variables
BINARY_NAME := cli-recover
VERSION := v2.0.0-alpha
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -s -w"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# Build targets
.PHONY: all build clean test test-coverage install help

# Default target
all: clean build

build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Cross-platform builds
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64

build-darwin-amd64:
	@echo "Building for Darwin AMD64..."
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .

build-darwin-arm64:
	@echo "Building for Darwin ARM64..."
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .

build-linux-amd64:
	@echo "Building for Linux AMD64..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .

build-linux-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p dist
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .

build-windows-amd64:
	@echo "Building for Windows AMD64..."
	@mkdir -p dist
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

# Development
run:
	$(GOCMD) run . $(ARGS)

test:
	$(GOTEST) -v -cover ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html

install:
	$(GOCMD) install $(LDFLAGS) .

# Create checksums for releases
checksums:
	@cd dist && sha256sum * > checksums.txt

# Display version
version:
	@echo "$(VERSION)"

# Help
help:
	@echo "cli-recover v2.0 Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          - Build for current platform"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install        - Install to GOPATH/bin"
	@echo "  make run ARGS=...   - Run with arguments"
	@echo "  make version        - Show version"
	@echo ""
	@echo "Platform-specific builds:"
	@echo "  make build-darwin-amd64  - macOS Intel"
	@echo "  make build-darwin-arm64  - macOS Apple Silicon"
	@echo "  make build-linux-amd64   - Linux x64"
	@echo "  make build-linux-arm64   - Linux ARM64"
	@echo "  make build-windows-amd64 - Windows x64"
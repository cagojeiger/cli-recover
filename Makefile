# Variables
BINARY_NAME := cli-recover
# Git 태그 기반 자동 버전 감지 (태그가 없으면 dev)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build targets
.PHONY: help all build build-all clean test test-coverage deps run version

# Default target - show help
help:
	@echo "CLI-Recover Makefile Commands:"
	@echo ""
	@echo "  make build         - Build for current platform"
	@echo "  make build-all     - Build for all platforms"
	@echo "  make run           - Build and run version command"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make version       - Show current version"
	@echo "  make deps          - Download dependencies"
	@echo ""
	@echo "Platform-specific builds:"
	@echo "  make build-darwin-amd64  - Build for macOS Intel"
	@echo "  make build-darwin-arm64  - Build for macOS Apple Silicon"
	@echo "  make build-linux-amd64   - Build for Linux x86_64"
	@echo "  make build-linux-arm64   - Build for Linux ARM64"

all: clean deps build

deps:
	$(GOMOD) download
	$(GOMOD) tidy

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/cli-recover

build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/cli-recover

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/cli-recover

build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/cli-recover

build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/cli-recover

test:
	$(GOTEST) -v ./...

# Test with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*

run: build
	./$(BINARY_NAME) --version

# Create checksums for releases
checksums:
	sha256sum $(BINARY_NAME)-* > checksums.txt

# Display current version
version:
	@echo "Current version: $(VERSION)"
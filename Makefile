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
	@echo "  make build              - Build for current platform"
	@echo "  make build-all          - Build for all platforms"
	@echo "  make run                - Build and run version command"
	@echo "  make test               - Run tests"
	@echo "  make test-coverage      - Run tests with coverage (excluding TUI)"
	@echo "  make test-coverage-all  - Run tests with coverage (including TUI)"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make version            - Show current version"
	@echo "  make deps               - Download dependencies"
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

# Test with coverage (excluding TUI package)
test-coverage:
	@echo "Running tests with coverage (excluding TUI)..."
	@$(GOTEST) -coverprofile=coverage.out \
		$(shell go list ./... | grep -v "/internal/tui$$") || true
	@echo ""
	@echo "========== Coverage Summary =========="
	@$(GOCMD) tool cover -func=coverage.out | grep "total:" || echo "No coverage data"
	@echo ""
	@echo "Package-level coverage:"
	@$(GOCMD) tool cover -func=coverage.out | grep -E "^github.com/cagojeiger/cli-recover/(cmd|internal/kubernetes|internal/backup|internal/runner)" | grep -E "\s[0-9]+\.[0-9]+%" || true
	@echo "======================================"
	@echo ""
	@echo "Generating HTML coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

# Test with coverage including TUI (for reference)
test-coverage-all:
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

# TDD Development workflow
.PHONY: tdd test-unit test-integration lint dev-tools

test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -run Integration ./...

lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run 'make dev-tools' to install."; \
	fi

dev-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# TDD watch mode (requires entr)
tdd:
	@echo "Starting TDD watch mode..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -c go test -v ./...; \
	else \
		echo "entr not installed. Install it with: brew install entr (macOS) or apt-get install entr (Linux)"; \
	fi
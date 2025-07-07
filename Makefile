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
.PHONY: help all build build-all clean test test-coverage test-report test-quality test-fast deps run version

# Default target - show help
help:
	@echo "CLI-Recover Makefile Commands:"
	@echo ""
	@echo "  make build              - Build for current platform"
	@echo "  make build-all          - Build for all platforms"
	@echo "  make run                - Build and run version command"
	@echo "  make test               - Run tests (excluding legacy code)"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo "  make test-report        - Detailed coverage report by package"
	@echo "  make test-quality       - Check coverage meets quality standards"
	@echo "  make test-fast          - Run fast unit tests only"
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
	@echo "Running tests (excluding legacy code)..."
	$(GOTEST) -v ./cmd/... ./internal/...

# Test with coverage (excluding legacy code)
test-coverage:
	@echo "Running tests with coverage (excluding legacy code)..."
	@$(GOTEST) -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo ""
	@echo "========== Coverage Summary =========="
	@$(GOCMD) tool cover -func=coverage.out | grep "total:" || echo "No coverage data"
	@echo ""
	@echo "🎯 Target: 90% | Current: $$($(GOCMD) tool cover -func=coverage.out | grep "total:" | awk '{print $$3}' || echo "0%")"
	@echo ""
	@echo "Generating HTML coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

# Detailed coverage report by package
test-report:
	@echo "Generating detailed coverage report..."
	@$(GOTEST) -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo ""
	@echo "========== Package Coverage Details =========="
	@echo "Architecture Layers:"
	@$(GOCMD) tool cover -func=coverage.out | grep -E "github.com/cagojeiger/cli-recover/(cmd|internal/domain|internal/infrastructure|internal/application)" | sort
	@echo ""
	@echo "Summary by Layer:"
	@echo "  CMD Layer:           $$($(GOCMD) tool cover -func=coverage.out | grep "cmd/" | awk '{sum+=$$3; count++} END {if(count>0) printf "%.1f%%", sum/count; else print "0%"}')"
	@echo "  Domain Layer:        $$($(GOCMD) tool cover -func=coverage.out | grep "internal/domain" | awk '{sum+=$$3; count++} END {if(count>0) printf "%.1f%%", sum/count; else print "0%"}')"
	@echo "  Infrastructure:      $$($(GOCMD) tool cover -func=coverage.out | grep "internal/infrastructure" | awk '{sum+=$$3; count++} END {if(count>0) printf "%.1f%%", sum/count; else print "0%"}')"
	@echo "  Application:         $$($(GOCMD) tool cover -func=coverage.out | grep "internal/application" | awk '{sum+=$$3; count++} END {if(count>0) printf "%.1f%%", sum/count; else print "0%"}')"
	@echo ""
	@echo "Overall: $$($(GOCMD) tool cover -func=coverage.out | grep "total:" | awk '{print $$3}')"
	@echo "=============================================="

# Check if coverage meets quality standards
test-quality:
	@echo "Checking test quality standards..."
	@$(GOTEST) -coverprofile=coverage.out ./cmd/... ./internal/...
	@COVERAGE=$$($(GOCMD) tool cover -func=coverage.out | grep "total:" | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE >= 70" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage meets minimum standard (70%)"; \
		if [ $$(echo "$$COVERAGE >= 90" | bc -l) -eq 1 ]; then \
			echo "🏆 Excellent coverage! Target achieved (90%+)"; \
		else \
			echo "🎯 Good coverage. Target: 90%, Current: $$COVERAGE%"; \
		fi \
	else \
		echo "❌ Coverage below minimum standard. Target: 70%, Current: $$COVERAGE%"; \
		exit 1; \
	fi

# Fast unit tests only
test-fast:
	@echo "Running fast unit tests..."
	$(GOTEST) -short ./cmd/... ./internal/...

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
.PHONY: tdd test-unit test-integration lint dev-tools test-ci quality-gate coverage-check

test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./cmd/... ./internal/...

test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -run Integration ./cmd/... ./internal/...

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
		find ./cmd ./internal -name "*.go" | entr -c go test -v ./cmd/... ./internal/...; \
	else \
		echo "entr not installed. Install it with: brew install entr (macOS) or apt-get install entr (Linux)"; \
	fi

# Additional quality gates for CI/CD

test-ci:
	@echo "Running CI test suite..."
	@$(GOTEST) -race -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo "CI tests completed successfully"

quality-gate: test-quality lint
	@echo "✅ All quality gates passed!"

coverage-check:
	@echo "Coverage trend analysis..."
	@$(GOTEST) -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo "Current coverage: $$($(GOCMD) tool cover -func=coverage.out | grep "total:" | awk '{print $$3}')"
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

# Default target - show help
.PHONY: help
help:
	@echo "CLI-Recover - Essential Commands"
	@echo "================================"
	@echo ""
	@echo "Daily Development:"
	@echo "  make fmt       - Format Go code"
	@echo "  make test      - Run unit tests"
	@echo "  make build     - Build binary for current platform"
	@echo ""
	@echo "Quality Checks (before commit):"
	@echo "  make check     - Run fmt, vet, and mod tidy"
	@echo "  make coverage  - Test coverage report (target: 90%)"
	@echo "  make lint      - Run golangci-lint"
	@echo ""
	@echo "Development:"
	@echo "  make tdd       - Watch mode for TDD"
	@echo "  make run       - Build and run version"
	@echo ""
	@echo "CI/CD:"
	@echo "  make ci        - Full CI pipeline"
	@echo "  make quality   - Quality gate checks"
	@echo ""
	@echo "Release:"
	@echo "  make version   - Show current version"
	@echo "  make build-all - Build for all platforms"
	@echo "  make checksums - Generate checksums"

# Essential targets
.PHONY: all fmt test build

all: check test build

# Daily development commands
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...

test:
	@echo "Running tests..."
	@$(GOTEST) -v -short ./cmd/... ./internal/...

build:
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/cli-recover

# Quality checks
.PHONY: check coverage lint

check: fmt
	@echo "Running code checks..."
	@$(GOCMD) vet ./...
	@$(GOMOD) tidy
	@echo "✅ All checks passed"

coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo ""
	@echo "========== Coverage Summary =========="
	@$(GOCMD) tool cover -func=coverage.out | grep "total:" || echo "No coverage data"
	@echo ""
	@COVERAGE=$$($(GOCMD) tool cover -func=coverage.out | grep "total:" | awk '{print $$3}' | sed 's/%//'); \
	echo "🎯 Target: 90% | Current: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE >= 90" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage target achieved!"; \
	else \
		echo "⚠️  Coverage below target (90%)"; \
	fi
	@echo ""
	@echo "Generating HTML report: coverage.html"
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed."; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Development tools
.PHONY: tdd run clean deps

tdd:
	@echo "Starting TDD watch mode..."
	@if command -v entr >/dev/null 2>&1; then \
		find ./cmd ./internal -name "*.go" | entr -c go test -v ./cmd/... ./internal/...; \
	else \
		echo "⚠️  entr not installed."; \
		echo "Install with: brew install entr (macOS) or apt-get install entr (Linux)"; \
	fi

run: build
	@./$(BINARY_NAME) --version

clean:
	@echo "Cleaning build artifacts..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-* coverage.out coverage.html

deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# CI/CD
.PHONY: ci quality

ci: check test coverage lint
	@echo ""
	@echo "✅ CI pipeline completed successfully"

quality:
	@echo "Running quality gate checks..."
	@$(MAKE) check
	@$(MAKE) coverage
	@COVERAGE=$$($(GOCMD) tool cover -func=coverage.out | grep "total:" | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE >= 90" | bc -l) -eq 1 ]; then \
		echo "✅ Quality gate: PASSED (coverage: $$COVERAGE%)"; \
	else \
		echo "❌ Quality gate: FAILED (coverage: $$COVERAGE%, required: 90%)"; \
		exit 1; \
	fi
	@$(MAKE) lint

# Release targets
.PHONY: version build-all checksums

version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(GIT_COMMIT)"
	@echo "Built:   $(BUILD_DATE)"

build-all:
	@echo "Building for all platforms..."
	@echo "  → Darwin AMD64..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/cli-recover
	@echo "  → Darwin ARM64..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/cli-recover
	@echo "  → Linux AMD64..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/cli-recover
	@echo "  → Linux ARM64..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/cli-recover
	@echo "✅ All platforms built successfully"

checksums:
	@echo "Generating checksums..."
	@sha256sum $(BINARY_NAME)-* > checksums.txt
	@echo "✅ Checksums saved to checksums.txt"
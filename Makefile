# Variables
BINARY := cli-pipe
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

# Default target
.DEFAULT_GOAL := build
.PHONY: build test test-coverage coverage-html coverage-func lint clean version release release-build

# Core targets
build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/cli-pipe

test:
	go test -v -cover ./...

# Generate coverage profile and show report
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Show function-level coverage
coverage-func:
	@if [ -f coverage.out ]; then \
		go tool cover -func=coverage.out; \
	else \
		echo "No coverage.out file found. Run 'make test-coverage' first."; \
	fi

lint:
	go vet ./...
	gofmt -l .

clean:
	rm -f $(BINARY)
	rm -f *.out
	rm -rf dist/

version:
	@echo $(VERSION)

# Release targets
release: clean
	@if [ -z "$(TAG)" ]; then echo "Usage: make release TAG=v2.0.0"; exit 1; fi
	git tag -a $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)
	@echo "Release $(TAG) created and pushed!"
	@echo "GitHub Actions will now build and publish the release."

release-build:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/cli-pipe
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/cli-pipe
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/cli-pipe
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 ./cmd/cli-pipe
	cd dist && sha256sum * > checksums.txt
# Variables
BINARY := cli-recover
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

# Default target
.DEFAULT_GOAL := build
.PHONY: build test lint clean version release release-build

# Core targets
build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test -v -cover ./...

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
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .
	cd dist && sha256sum * > checksums.txt
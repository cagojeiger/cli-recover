# Reusable Patterns

## Version Injection Pattern
```makefile
VERSION := v0.1.0
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
```

```go
var version = "dev"  // Will be overridden by ldflags
```

## Cobra CLI Structure Pattern
```go
var rootCmd = &cobra.Command{
    Use:     "app-name",
    Short:   "Brief description",
    Long:    `Detailed description`,
    Version: version,
}
```

## Makefile Cross-Platform Build Pattern
```makefile
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)

build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64
```

## GitHub Actions Release Pattern
```yaml
on:
  push:
    tags:
      - 'v*'
      
- name: Build binaries
  run: |
    VERSION=${GITHUB_REF#refs/tags/}
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION"
```

## Context Engineering Update Pattern
1. Create/Update relevant `.memory/short-term/` during work
2. Move to `.memory/long-term/` when complete
3. Create `.checkpoint/` at milestones
4. Keep `.planning/` updated with next steps
# cli-recover v2.0

A simple, isolated Kubernetes backup and restore tool.

## Philosophy

> "Isolation > Reusability"  
> "Simplicity > Complexity"

## Status

🚧 **v2.0.0-alpha** - Complete rewrite in progress

This is a fresh start based on lessons learned from v1. We're building a simpler, more focused tool with complete provider isolation.

## Features (Planned)

- ✅ Single provider: filesystem
- ✅ CLI-only interface 
- ✅ Zero external dependencies
- ✅ Progress reporting built-in
- ✅ Test-driven development

## Installation

### From Source

```bash
go install github.com/cagojeiger/cli-recover@latest
```

### Pre-built Binaries

Coming soon...

## Usage

```bash
# Show version
cli-recover version

# Backup (coming soon)
cli-recover backup <namespace> <pod> <path> -o backup.tar

# Restore (coming soon)  
cli-recover restore <namespace> <pod> <path> -i backup.tar
```

## Development

### Prerequisites

- Go 1.21+
- kubectl (for runtime)
- Access to a Kubernetes cluster

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run with coverage
make test-coverage
```

### Project Structure

```
cli-recover/
├── main.go           # Entry point
├── backup.go         # Backup logic (TBD)
├── restore.go        # Restore logic (TBD)
├── progress.go       # Progress reporting (TBD)
└── Makefile         # Build automation
```

## Design Principles

1. **No shared interfaces** - Each component is completely isolated
2. **Direct implementation** - No unnecessary abstractions
3. **TDD from day one** - Test first, implement second
4. **Copy over share** - Duplication is better than wrong abstraction

## Roadmap

See [.planning/00-roadmap.md](.planning/00-roadmap.md) for detailed phases.

- [x] Phase 1: Minimal foundation
- [ ] Phase 2: Filesystem backup (TDD)
- [ ] Phase 3: Progress reporting
- [ ] Phase 4: Filesystem restore
- [ ] Phase 5: CLI polish
- [ ] Phase 6: Integrity & safety
- [ ] Phase 7: Performance & scale
- [ ] Phase 8: v2.0 release

## Contributing

This project follows strict TDD practices. Please ensure:

1. Write tests before implementation
2. Keep complexity below 20/100
3. No external dependencies in v2.0
4. Follow the isolation principle

## License

[Apache 2.0](LICENSE)
# Technology Stack

## Core Technologies

### Language & Runtime
- **Go 1.24.3**: Primary development language
- **Standard Library**: JSON parsing, file operations, command execution

### CLI Framework
- **Cobra**: Command-line interface structure
- **pflag**: Command-line flag parsing

### TUI Framework
- **Bubble Tea**: Terminal user interface framework
- **Lipgloss**: Styling and layout for terminal UIs
- **Termenv**: Terminal environment detection

### Testing Framework
- **Go testing**: Built-in testing framework
- **teatest**: Bubble Tea TUI testing utilities
- **Golden Files**: Mock data for kubectl responses

## Development Tools

### Build & Dependencies
- **Go Modules**: Dependency management
- **Makefile**: Build automation scripts

### Code Quality
- **go fmt**: Code formatting
- **go vet**: Static analysis
- **golint**: Style checking (implied)

### Testing & Coverage
- **go test**: Unit and integration testing
- **go tool cover**: Test coverage analysis
- **CI/CD**: GitHub Actions (inferred from structure)

## External Integrations

### Kubernetes
- **kubectl**: Command-line tool for K8s cluster interaction
- **JSON API**: Kubernetes REST API responses
- **tar**: Unix archiving tool for backup creation

### File System
- **Golden Files**: Test data in `testdata/kubectl/`
- **Temporary Files**: For backup operations

## Architecture Decisions

### Why Bubble Tea?
- Modern, reactive TUI framework
- Excellent testing support with teatest
- Clean model-view-update architecture

### Why Golden Files?
- Deterministic testing without K8s cluster
- Fast CI/CD execution
- Reproducible test results

### Why Internal Packages?
- Clear separation of concerns
- Testable business logic
- Follows Go project layout standards
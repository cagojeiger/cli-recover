# Architecture

## Package Structure
- **cmd/cli-recover/**
  - main.go: Entry point
  - backup_filesystem.go: Backup logic
  
- **internal/kubernetes/**
  - client.go: K8s operations
  - types.go: Resource types
  
- **internal/runner/**
  - runner.go: Command execution interface
  - golden.go: Test runner
  
- **internal/tui/**
  - model.go: State management
  - view.go: 4-stage layout
  - screens.go: UI components
  - handlers_*.go: Input handling
  - executor.go: Command execution
  - command_builder.go: Type-safe builder

## Key Patterns
- **Runner Interface**: Testable command execution
- **CommandBuilder**: Incremental command construction
- **4-Stage Layout**: Header | Content | Command | Footer
- **Screen States**: Main → Namespace → Pod → Directory → Options

## Data Flow
- **CLI**: Args → Validate → Execute → Output
- **TUI**: Navigate → Build → Preview → Execute

## Current Issues
- StreamingExecutor blocks UI
- No async command execution
- Missing progress feedback
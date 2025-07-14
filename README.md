# cli-pipe

Unix command pipeline orchestrator - Make every CLI task trackable and reproducible

## Overview

cli-pipe is a tool that allows you to define and execute Unix command pipelines using YAML files. It automatically monitors execution (bytes, lines, time) and logs all pipeline runs for debugging and auditing.

## Features

- 📝 **YAML-based pipeline definition** - Define complex command sequences in readable YAML
- 🔗 **Explicit stream connections** - Clear data flow with named input/output streams
- 📊 **Automatic monitoring** - Always tracks bytes processed, lines, and execution time
- 📁 **Persistent logging** - All runs are logged to `~/.cli-pipe/logs/`
- 🏗️ **Clean architecture** - Simplified design with configuration-based approach
- ✅ **High test coverage** - 90%+ test coverage across all packages

## Installation

```bash
go build -o cli-pipe ./cmd/cli-pipe
```

## Quick Start

1. Initialize configuration:
```bash
cli-pipe init
```

2. Run a pipeline:
```bash
cli-pipe run <pipeline.yaml>
```

## Pipeline YAML Format

```yaml
name: pipeline-name
description: Pipeline description

steps:
  - name: step1
    run: command to execute
    output: stream-name
    
  - name: step2
    run: another command
    input: stream-name
    output: result
```

## Examples

See the `examples/` directory for sample pipelines:

- `hello-world.yaml` - Basic pipeline with text transformation
- `word-count.yaml` - Count words in generated text
- `file-processing.yaml` - Process and analyze files
- `date-time.yaml` - Date/time formatting
- `simple-test.yaml` - Minimal test pipeline
- `backup.yaml` - Create compressed backups

### Running an example:

```bash
./cli-pipe run examples/hello-world.yaml
```

Example output:
```
Executing pipeline: hello-world
Logging to: /home/user/.cli-pipe/logs/hello-world_20250714_090000

Command: echo "Hello, World!" | tr 'a-z' 'A-Z' | sed 's/WORLD/CLI-PIPE/'

HELLO, CLI-PIPE!

==================================================
Pipeline completed
• Duration: 5.2ms
• Bytes processed: 17 B
• Lines processed: 1
• Status: Success
• Logs: /home/user/.cli-pipe/logs/hello-world_20250714_090000
```

## Configuration

cli-pipe stores its configuration in `~/.cli-pipe/config.yaml`:

```yaml
version: 1
logs:
  directory: /home/user/.cli-pipe/logs
  retention_days: 7
```

You can customize the log directory and retention period by editing this file.

## Architecture

The project follows a simplified clean architecture:

```
internal/
├── config/        # Configuration management
├── logger/        # Structured logging system
│   ├── logger.go     # Core logger interface and implementation
│   ├── rotator.go    # Log file rotation with compression
│   └── cleaner.go    # Old log cleanup functionality
└── pipeline/      # Core pipeline execution logic
    ├── builder.go    # Command building
    ├── executor.go   # Pipeline execution with monitoring
    ├── monitor.go    # Unified monitoring (bytes, lines, time)
    ├── parser.go     # YAML parsing
    └── pipeline.go   # Data structures
cmd/cli-pipe/      # CLI entry point
```

Key design decisions:
- All pipelines are automatically monitored and logged
- Configuration-based approach instead of command-line flags
- Unified monitoring combines bytes, lines, and time tracking
- Simplified execution path with no feature toggles

## Development

This project was developed using Test-Driven Development (TDD):

1. Write failing tests first
2. Implement minimal code to pass
3. Refactor while keeping tests green

### Running tests:

```bash
go test ./... -v -cover
```

### Building:

```bash
go build -o cli-pipe ./cmd/cli-pipe
```

## License

MIT
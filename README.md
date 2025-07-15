# cli-pipe

Unix command pipeline orchestrator - Make every CLI task trackable and reproducible

## Overview

cli-pipe is a tool that allows you to define and execute Unix command pipelines using YAML files. It automatically monitors execution (bytes, lines, time) and logs all pipeline runs for debugging and auditing.

## Features

- 📝 **YAML-based pipeline definition** - Define complex command sequences in readable YAML
- 🔗 **Explicit stream connections** - Clear data flow with named input/output streams
- 🌳 **Tree-structured pipelines** - Support for branching pipelines (one input, multiple outputs)
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

### Linear Pipelines
- `hello-world.yaml` - Basic pipeline with text transformation
- `word-count.yaml` - Count words in generated text
- `file-processing.yaml` - Process and analyze files
- `date-time.yaml` - Date/time formatting
- `simple-test.yaml` - Minimal test pipeline
- `backup.yaml` - Create compressed backups

### Tree-Structured Pipelines (NEW!)
- `tree-simple-branch.yaml` - Simple branching (one output to two consumers)
- `tree-multi-branch.yaml` - Multiple branching (one output to three consumers)
- `tree-multi-level.yaml` - Multi-level tree (branching after branching)
- `tree-complex.yaml` - Complex tree with mixed branches and isolated steps

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

## Tree-Structured Pipelines

cli-pipe now supports tree-structured pipelines, allowing you to split data flow to multiple consumers:

### Simple Branching Example

```yaml
name: data-analysis
steps:
  - name: fetch-data
    run: curl -s https://api.example.com/data
    output: raw_data
    
  - name: backup
    run: gzip > backup.gz
    input: raw_data
    
  - name: analyze
    run: jq .users
    input: raw_data
```

In this example, the output from `fetch-data` is sent to both `backup` and `analyze` steps simultaneously using `tee`.

### Multi-Level Trees

You can create complex trees with multiple levels:

```yaml
name: log-processing
steps:
  - name: read-logs
    run: cat server.log
    output: logs
    
  - name: extract-errors
    run: grep ERROR
    input: logs
    output: errors
    
  - name: extract-warnings
    run: grep WARN
    input: logs
    output: warnings
    
  - name: count-errors
    run: wc -l
    input: errors
    
  - name: alert-errors
    run: mail -s "Errors found" admin@example.com
    input: errors
    
  - name: summarize-warnings
    run: sort | uniq -c
    input: warnings
```

### Visual Pipeline Structure

When executing tree pipelines, cli-pipe displays the structure:

```
Pipeline structure:
└── [read-logs] cat server.log
    ├── [extract-errors] grep ERROR
    │   ├── [count-errors] wc -l
    │   └── [alert-errors] mail -s "Errors found" admin@example.com
    └── [extract-warnings] grep WARN
        └── [summarize-warnings] sort | uniq -c
```

### How It Works

Tree pipelines use Unix process substitution and `tee` to efficiently split data:
- Linear pipelines: `cmd1 | cmd2 | cmd3`
- Branching: `cmd1 | tee >(cmd2) >(cmd3) > /dev/null`
- Multi-level: `cmd1 | tee >(cmd2 | cmd4) >(cmd3 | cmd5) > /dev/null`

### Limitations

- Each step can have only one input (no merging)
- No circular dependencies allowed
- All branches execute in parallel

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
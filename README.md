# cli-pipe

Unix command pipeline orchestrator - Make every CLI task trackable and reproducible

## Overview

cli-pipe is a tool that allows you to define and execute Unix command pipelines using YAML files. It provides explicit input/output stream management, execution logging, and pipeline orchestration.

## Features

- 📝 **YAML-based pipeline definition** - Define complex command sequences in readable YAML
- 🔗 **Explicit stream connections** - Clear data flow with named input/output streams
- 📊 **Execution logging** - Track every step of your pipeline execution
- 🏗️ **Clean architecture** - Hexagonal architecture with TDD approach
- ✅ **High test coverage** - 90%+ test coverage across all packages

## Installation

```bash
go build -o cli-pipe ./cmd/cli-pipe
```

## Usage

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

### Running an example:

```bash
./cli-pipe run examples/hello-world.yaml
```

## Architecture

The project follows hexagonal (clean) architecture:

- **Domain layer** - Core business logic (entities, value objects)
- **Application layer** - Use cases (pipeline execution)
- **Infrastructure layer** - External interfaces (YAML parsing, CLI)

## Development

This project was developed using Test-Driven Development (TDD):

1. Write failing tests first
2. Implement minimal code to pass
3. Refactor while keeping tests green

### Running tests:

```bash
go test ./... -v -cover
```

## License

MIT
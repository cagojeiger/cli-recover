# Technical Decisions

## 2025-07-06: CLI Framework Selection
- **Decision**: Cobra (github.com/spf13/cobra)
- **Rationale**:
  * 41k+ GitHub stars
  * Industry standard (kubectl, docker, helm)
  * Built-in version flag support
  * Excellent subcommand structure for future
- **Alternatives Considered**:
  * urfave/cli (23k stars) - Simpler but less features
  * kingpin - Deprecated

## 2025-07-06: Project Structure
- **Decision**: Simple start with flat structure
- **Rationale**:
  * Avoid over-engineering
  * main.go contains all code initially
  * Will expand to internal/ when needed
- **Reference**: "Start simple, always with a flat structure"

## 2025-07-06: Cross-Platform Support
- **Decision**: Support macOS and Linux from start
- **Platforms**:
  * darwin/amd64 (Intel Mac)
  * darwin/arm64 (Apple Silicon)
  * linux/amd64
  * linux/arm64
- **Rationale**: Kubernetes runs on Linux servers
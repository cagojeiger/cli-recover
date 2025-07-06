# Working Context

## Current Environment
- **Repository**: cli-recover (Kubernetes backup tool)
- **Working Directory**: /Users/kangheeyong/project/cli-recover
- **Go Module**: github.com/cagojeiger/cli-restore
- **Go Version**: 1.24.3

## Project State
- **Structure**:  Refactored to standard Go layout
- **Build Status**:  Builds successfully
- **Test Status**:  12/12 tests passing
- **Coverage**: L 0.0% (needs improvement to 90%+)

## Recent Changes
- Moved from monolithic `tui.go` (922 lines) to modular structure
- Created internal packages: kubernetes/, runner/, tui/
- Updated all imports to use correct module path
- Moved testdata to root level with updated relative paths

## Active Files
### Recently Modified
- `cmd/cli-restore/main.go`: Simplified to use internal packages
- `internal/tui/*.go`: Split TUI logic into focused files
- `internal/kubernetes/*.go`: K8s operations and types
- `internal/runner/*.go`: Command execution abstraction

### Test Files
- `cmd/cli-restore/main_test.go`: Updated with new imports
- `cmd/cli-restore/tui_test.go`: TUI integration tests
- Golden files in `testdata/kubectl/`: Mock kubectl responses

## Dependencies
- Bubble Tea ecosystem for TUI
- Cobra for CLI structure
- Standard library for core functionality
- No external K8s dependencies (uses kubectl binary)

## Current Challenges
- Test coverage reporting shows 0% due to main() function only
- Need to add unit tests for internal packages
- Some functions may exceed 50-line limit
- CLI mode flags and debug mode not yet implemented
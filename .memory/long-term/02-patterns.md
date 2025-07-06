# Reusable Patterns

## Architecture Patterns

### Strategy Pattern for Testing
**Use Case**: Command execution with test/production variants
**Implementation**:
```go
type Runner interface {
    Run(cmd string, args ...string) ([]byte, error)
}

// Production
type ShellRunner struct{}
func (r *ShellRunner) Run(cmd string, args ...string) ([]byte, error) {
    return exec.Command(cmd, args...).Output()
}

// Testing  
type GoldenRunner struct { dir string }
func (r *GoldenRunner) Run(cmd string, args ...string) ([]byte, error) {
    // Read from golden files
}
```
**Benefits**: Zero external dependencies in tests, easy switching

### Package Dependency Hierarchy
**Pattern**: Clear unidirectional dependencies
**Structure**:
```
cmd/cli-restore (main)
    ↓
internal/tui (UI logic)
    ↓
internal/kubernetes (business logic)
    ↓
internal/runner (execution abstraction)
```
**Benefits**: No circular imports, clear responsibility layers

## TUI Patterns

### Screen State Management
**Pattern**: Enum-based screen transitions with centralized state
**Implementation**:
```go
type Screen int
const (
    ScreenMain Screen = iota
    ScreenNamespaceList
    // ...
)

type Model struct {
    screen Screen
    // All screen states in one place
}
```
**Benefits**: Single source of truth, easy navigation logic

### Progressive Disclosure UI
**Pattern**: Tab-based category navigation
**Use Case**: Complex option configuration
**Implementation**: Tab switching with category-specific option lists
**Benefits**: Reduces cognitive load, organizes related options

### Keyboard Handling Patterns
**Pattern**: Consistent key bindings across screens
**Standard bindings**:
- `j/k`: Up/down navigation
- `Enter`: Selection/confirmation
- `Space`: Toggle/special action
- `Tab`: Category switching
- `b/Esc`: Back navigation
- `q`: Quit
**Benefits**: Familiar Vi-style navigation, muscle memory

## Testing Patterns

### Golden File Testing
**Pattern**: Mock external command responses
**File naming**: `{command}-{args-sanitized}.golden`
**Sanitization**: Replace special characters, normalize paths
**Environment switching**: `USE_GOLDEN=true` for test mode
**Benefits**: Fast tests, no external dependencies

### TUI Integration Testing
**Pattern**: Complete user journey simulation
**Tools**: teatest framework for Bubble Tea
**Structure**:
1. Setup model with test data
2. Send key sequences
3. Wait for expected output
4. Verify final state
**Benefits**: Catches integration issues, validates UX flows

### Table-Driven Tests
**Pattern**: Struct slices for multiple test cases
**Template**:
```go
tests := []struct {
    name    string
    input   InputType
    want    OutputType
    wantErr bool
}{
    {"case 1", input1, output1, false},
    {"error case", badInput, nil, true},
}
```
**Benefits**: Easy to add cases, clear test intent

## Error Handling Patterns

### Context-Preserving Error Wrapping
**Pattern**: Add context while preserving original error
**Implementation**: `fmt.Errorf("operation failed: %w", originalErr)`
**Benefits**: Clear error chains, easier debugging

### Early Return Pattern
**Pattern**: Check errors immediately and return
**Template**:
```go
result, err := operation()
if err != nil {
    return nil, fmt.Errorf("context: %w", err)
}
// Continue with result
```
**Benefits**: Reduces nesting, clear error handling

## Code Organization Patterns

### Feature-Based File Organization
**Pattern**: Group related functions in focused files
**Example**: `handlers.go`, `screens.go`, `options.go`
**Benefits**: Easy to locate functionality, logical grouping

### Exported Interface, Unexported Implementation
**Pattern**: Public interface with private implementations
**Use Case**: Runner interface with ShellRunner/GoldenRunner
**Benefits**: Clean API, implementation flexibility
# Architectural Decisions

## Major Design Decisions

### Package Structure (2025-01-06)
**Decision**: Split monolithic TUI into internal packages
**Rationale**: 
- CLAUDE.md compliance (files <500 lines)
- Better separation of concerns
- Improved testability
**Implementation**: 
- `internal/kubernetes/`: K8s operations
- `internal/runner/`: Command execution abstraction  
- `internal/tui/`: UI components
**Result**: Reduced complexity from ~65 to ~45

### Testing Strategy
**Decision**: Golden file testing for K8s interactions
**Rationale**:
- No K8s cluster required for CI/CD
- Deterministic test results
- Fast test execution
**Implementation**: Mock kubectl responses in `testdata/`
**Trade-offs**: Need to maintain test data manually

### TUI Framework Choice
**Decision**: Bubble Tea for terminal interface
**Rationale**:
- Modern reactive architecture
- Excellent testing support (teatest)
- Clean model-view-update pattern
**Alternatives considered**: tcell, termui
**Result**: Rich interactive experience with reliable testing

### Command Execution Pattern
**Decision**: Strategy pattern for runner interface
**Rationale**:
- Easy testing with mock implementations
- Clean separation between test and production
**Implementation**: Runner interface with Golden/Shell implementations
**Benefits**: No external dependencies in tests

## Technical Decisions

### Error Handling
**Pattern**: Wrap errors with context using fmt.Errorf
**Example**: `fmt.Errorf("failed to get pods: %w", err)`
**Benefit**: Clear error trails for debugging

### State Management
**Pattern**: Centralized state in TUI model
**Rationale**: Single source of truth for UI state
**Implementation**: Model struct with all screen states

### File Organization
**Pattern**: Feature-based file organization within packages
**Example**: `handlers.go`, `screens.go`, `options.go` in tui/
**Benefit**: Easy to locate related functionality

## Rejected Approaches

### Direct K8s API Usage
**Rejected**: Using K8s Go client libraries
**Reason**: Complex dependencies, harder testing
**Chosen**: kubectl binary execution with JSON parsing

### Single Package Structure
**Rejected**: Keeping all code in main package
**Reason**: CLAUDE.md violations, poor separation
**Chosen**: Standard Go internal package layout
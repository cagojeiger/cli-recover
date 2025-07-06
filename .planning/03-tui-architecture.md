# TUI Architecture Design

## Overview
k9s 스타일의 전문적인 TUI 시스템 아키텍처

## Core Architecture

### 1. Layout System
```
┌─ Header (2 lines) ──────────┐
│ Title & Cluster Info        │ 
│ Navigation Breadcrumb       │
├─ Main Content (dynamic) ────┤
│ List-based Interface        │
├─ Command Preview (2-3 lines)┤
│ $ Generated command         │
├─ Footer (1 line) ──────────┤
│ Contextual shortcuts        │
└─────────────────────────────┘
```

### 2. Directory Structure
```
internal/
├── tui/
│   ├── core/
│   │   ├── app.go           # Bubble Tea application
│   │   ├── layout.go        # Layout system
│   │   ├── navigation.go    # Screen navigation
│   │   ├── state.go         # State management
│   │   └── types.go         # Common types
│   ├── views/
│   │   ├── action/          # Action selection view
│   │   ├── target/          # Target selection view
│   │   ├── pod/             # Pod backup views
│   │   │   ├── list.go      # Pod list
│   │   │   ├── paths.go     # Path selection
│   │   │   └── options.go   # Backup options
│   │   ├── mongodb/         # MongoDB backup views
│   │   ├── execution/       # Execution & progress
│   │   └── common/          # Shared components
│   ├── components/
│   │   ├── list.go          # List component
│   │   ├── table.go         # Table component
│   │   ├── input.go         # Input component
│   │   └── progress.go      # Progress bar
│   ├── cmd/
│   │   ├── builder.go       # Command builder
│   │   └── executor.go      # Command executor
│   ├── binary/
│   │   ├── manager.go       # Binary management
│   │   ├── embedded.go      # Embedded binaries
│   │   └── injection.go     # Pod injection logic
│   └── styles/
│       └── theme.go         # Color scheme & styles
├── backup/
│   ├── strategy/
│   │   ├── analyzer.go      # Capacity analysis
│   │   ├── selector.go      # Strategy selection
│   │   └── executor.go      # Execution logic
│   └── services/
│       ├── mongodb.go       # MongoDB specific
│       ├── minio.go         # MinIO specific
│       └── postgres.go      # PostgreSQL specific
```

### 3. State Management (Functional)

```go
// Immutable state
type AppState struct {
    CurrentView   ViewType
    Navigation    NavigationStack
    CommandParts  CommandBuilder
    Selections    SelectionState
    Execution     ExecutionState
}

// State transitions
type Transition func(AppState) AppState

// View interface
type View interface {
    ID() string
    Render(width, height int) string
    HandleEvent(event Event) (View, Command)
    DataRequirements() []DataRequest
}
```

### 4. Command Pattern Implementation

```go
type CommandBuilder struct {
    Action   string              // backup, restore, verify
    Target   string              // pod, mongodb, minio
    Resource string              // specific pod/db name
    Options  map[string]string   // flags and values
}

// Generates: cli-restore backup pod nginx /data --namespace prod
func (cb CommandBuilder) Build() string
```

### 5. Navigation Flow

```
Start
  │
  ├─> Action Selection (backup/restore/verify)
  │     │
  │     ├─> Target Selection (pod/mongodb/minio/...)
  │     │     │
  │     │     ├─> Resource Selection (specific pod/db)
  │     │     │     │
  │     │     │     ├─> Configuration (paths/options)
  │     │     │     │     │
  │     │     │     │     └─> Execution
  │     │     │     │
  │     │     │     └─> Back
  │     │     │
  │     │     └─> Back
  │     │
  │     └─> Back
  │
  └─> Exit
```

### 6. Key Bindings

#### Global
- `j/k` or `↑/↓`: Navigate up/down
- `Enter`: Select/Confirm
- `Space`: Toggle selection
- `/`: Search/Filter
- `b` or `Esc`: Back
- `?`: Help
- `q`: Quit

#### Context-specific
- `n`: Change namespace
- `a`: Select all
- `x`: Clear selections
- `p`: Pattern selection
- `v`: Verbose mode
- `c`: Cancel operation

### 7. View Types

#### List View
- Single column selection
- Search/filter support
- Keyboard navigation

#### Table View  
- Multi-column data
- Sortable columns
- Row selection

#### File Browser
- Tree structure
- Size information
- Multi-select support

#### Progress View
- Real-time updates
- Log streaming
- Pause/cancel support

### 8. Extensibility

#### Adding New Target
1. Add to target list
2. Create view module
3. Register command builder
4. No changes to core system

Example:
```go
// Register new target
RegisterTarget("elasticsearch", ElasticsearchTarget{
    Name: "Elasticsearch",
    Category: "Databases",
    ViewFactory: NewElasticsearchView,
    CommandBuilder: BuildElasticsearchCommand,
})
```

### 9. Error Handling

- Graceful degradation
- Clear error messages
- Retry mechanisms
- Fallback to CLI mode

### 10. Performance Considerations

- Lazy loading for large lists
- Debounced search
- Minimal redraws
- Efficient state updates

### 11. Binary Management System

#### Embedded Binary Architecture
```go
type BinaryManager struct {
    embedded  map[string][]byte  // Embedded binaries
    cache     string            // Local cache dir
    injected  map[string]string // Pod -> binary path
}

type BinaryStrategy int

const (
    UseLocal BinaryStrategy = iota
    UseEmbedded
    UsePodInternal
    RequireDownload
)

// Determine best binary strategy
func (bm *BinaryManager) DetermineStrategy(
    tool string, 
    pod *PodInfo,
) (BinaryStrategy, error) {
    // 1. Check local availability
    if bm.hasLocalTool(tool) {
        return UseLocal, nil
    }
    
    // 2. Check pod internal
    if pod.HasTool(tool) {
        return UsePodInternal, nil
    }
    
    // 3. Check embedded
    if bm.hasEmbedded(tool) {
        return UseEmbedded, nil
    }
    
    // 4. Require download
    return RequireDownload, nil
}
```

### 12. Backup Strategy Selection

#### Automatic Strategy Based on Capacity
```go
type BackupAnalyzer struct {
    kubectl KubectlClient
}

type BackupStrategy struct {
    Method      string // "streaming", "pod-internal", "port-forward"
    Reason      string
    SpaceNeeded int64
    TimeEstimate time.Duration
    Commands    []string
}

func (ba *BackupAnalyzer) AnalyzeBackup(
    service ServiceType,
    pod PodInfo,
    dataSize int64,
) (*BackupStrategy, error) {
    availableSpace := ba.getPodAvailableSpace(pod)
    
    // Size-based decision
    if dataSize > availableSpace * 0.8 {
        return &BackupStrategy{
            Method: "streaming",
            Reason: "Insufficient pod storage",
            Commands: ba.buildStreamingCommands(service, pod),
        }, nil
    }
    
    if dataSize < 10*GB && availableSpace > dataSize*2 {
        return &BackupStrategy{
            Method: "pod-internal",
            Reason: "Small dataset, faster internally",
            SpaceNeeded: dataSize * 2,
        }, nil
    }
    
    return &BackupStrategy{
        Method: "streaming",
        Reason: "Default safe strategy",
    }, nil
}
```
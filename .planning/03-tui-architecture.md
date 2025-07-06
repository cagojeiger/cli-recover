# TUI Architecture Design

## Overview
k9s 스타일의 전문적인 TUI - 단순화된 접근 (v0.2.0)

## Simplified Architecture (복잡도: 20)

### 1. Single File Start
- main.go에 모든 TUI 로직 포함
- 500줄 초과 시 분리
- Golden File 기반 개발

### 2. Core Components (main.go 내부)
```go
// Bubble Tea Model
type Model struct {
    screen     Screen      // current screen
    namespaces []string    // kubectl data
    pods       []Pod
    selected   int
    runner     Runner      // golden or shell
}

// Minimal screens
type Screen int
const (
    MainMenu Screen = iota
    NamespaceList
    PodList
    Executing
)
```

## Future Architecture (v0.3.0+)
*아래는 향후 확장 시 고려사항입니다*

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

### 2. Directory Structure (향후 고려)
```
# v0.3.0 이후 필요시 분리
internal/
├── tui/          # UI 관련
├── k8s/          # Kubectl 통합
└── backup/       # 백업 로직

# 현재 v0.2.0
cli-restore/
├── main.go       # 모든 코드
├── main_test.go  # TDD
└── testdata/     # Golden files
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
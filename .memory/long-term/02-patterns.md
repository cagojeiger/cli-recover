# 재사용 가능한 패턴

## UI 컴포넌트 패턴

### 1. Generic List Component
```go
type ListComponent[T any] struct {
    items      []T
    selected   int
    focused    bool
    height     int
    renderer   func(item T, selected bool) string
    onSelect   func(item T) tea.Cmd
    filter     string
}

func (l *ListComponent[T]) Update(msg tea.Msg) (*ListComponent[T], tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if l.selected > 0 {
                l.selected--
            }
        case "down", "j":
            if l.selected < len(l.items)-1 {
                l.selected++
            }
        case "enter":
            if l.onSelect != nil && l.selected < len(l.items) {
                return l, l.onSelect(l.items[l.selected])
            }
        }
    }
    return l, nil
}

func (l *ListComponent[T]) View() string {
    var b strings.Builder
    start := 0
    end := len(l.items)
    
    // 스크롤링 처리
    if l.height > 0 && len(l.items) > l.height {
        // 선택된 항목을 중앙에 유지
        if l.selected >= l.height/2 {
            start = l.selected - l.height/2
            end = start + l.height
        }
    }
    
    for i := start; i < end && i < len(l.items); i++ {
        b.WriteString(l.renderer(l.items[i], i == l.selected))
        b.WriteString("\n")
    }
    
    return b.String()
}
```

### 2. Form Component
```go
type FormField struct {
    Label       string
    Value       string
    Type        FieldType // text, password, select, checkbox
    Options     []string  // for select
    Validator   func(string) error
    Required    bool
}

type FormComponent struct {
    fields   []FormField
    selected int
    errors   map[int]string
}

func (f *FormComponent) Validate() bool {
    f.errors = make(map[int]string)
    valid := true
    
    for i, field := range f.fields {
        if field.Required && field.Value == "" {
            f.errors[i] = "This field is required"
            valid = false
        } else if field.Validator != nil {
            if err := field.Validator(field.Value); err != nil {
                f.errors[i] = err.Error()
                valid = false
            }
        }
    }
    
    return valid
}
```

### 3. Table Component
```go
type Column struct {
    Title string
    Width int
    Align Alignment // Left, Right, Center
}

type TableComponent[T any] struct {
    columns  []Column
    rows     []T
    renderer func(row T, col int) string
    selected int
    sorted   int // column index for sorting
}

func (t *TableComponent[T]) View() string {
    var b strings.Builder
    
    // Header
    for _, col := range t.columns {
        b.WriteString(padString(col.Title, col.Width, col.Align))
        b.WriteString(" ")
    }
    b.WriteString("\n")
    
    // Separator
    for _, col := range t.columns {
        b.WriteString(strings.Repeat("─", col.Width))
        b.WriteString(" ")
    }
    b.WriteString("\n")
    
    // Rows
    for i, row := range t.rows {
        for j, col := range t.columns {
            content := t.renderer(row, j)
            if i == t.selected {
                content = highlightStyle.Render(content)
            }
            b.WriteString(padString(content, col.Width, col.Align))
            b.WriteString(" ")
        }
        b.WriteString("\n")
    }
    
    return b.String()
}
```

## 상태 관리 패턴

### 1. State Machine
```go
type State int

const (
    StateIdle State = iota
    StateLoading
    StateError
    StateSuccess
)

type StateMachine struct {
    current     State
    transitions map[State][]State
}

func (sm *StateMachine) CanTransition(to State) bool {
    allowed, ok := sm.transitions[sm.current]
    if !ok {
        return false
    }
    
    for _, state := range allowed {
        if state == to {
            return true
        }
    }
    return false
}
```

### 2. Event Bus
```go
type EventType string

type Event struct {
    Type      EventType
    Timestamp time.Time
    Data      interface{}
}

type EventBus struct {
    subscribers map[EventType][]chan Event
    mu          sync.RWMutex
}

func (eb *EventBus) Subscribe(eventType EventType) <-chan Event {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    ch := make(chan Event, 10)
    eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
    return ch
}

func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    
    for _, ch := range eb.subscribers[event.Type] {
        select {
        case ch <- event:
        default:
            // 채널이 가득 찬 경우 스킵
        }
    }
}
```

## 비즈니스 로직 패턴

### 1. Service Layer
```go
type BackupService struct {
    repo      BackupRepository
    executor  CommandExecutor
    validator BackupValidator
}

func (s *BackupService) CreateBackup(ctx context.Context, req BackupRequest) (*BackupJob, error) {
    // 1. 검증
    if err := s.validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 2. 엔티티 생성
    job := &BackupJob{
        ID:        generateID(),
        Status:    JobStatusPending,
        CreatedAt: time.Now(),
    }
    
    // 3. 저장
    if err := s.repo.Save(job); err != nil {
        return nil, fmt.Errorf("failed to save job: %w", err)
    }
    
    // 4. 실행
    go s.executeJob(ctx, job)
    
    return job, nil
}
```

### 2. Repository with Cache
```go
type CachedRepository struct {
    base  Repository
    cache map[string]*CacheEntry
    mu    sync.RWMutex
    ttl   time.Duration
}

type CacheEntry struct {
    data      interface{}
    expiresAt time.Time
}

func (r *CachedRepository) Get(id string) (interface{}, error) {
    // 캐시 확인
    r.mu.RLock()
    if entry, ok := r.cache[id]; ok && entry.expiresAt.After(time.Now()) {
        r.mu.RUnlock()
        return entry.data, nil
    }
    r.mu.RUnlock()
    
    // 캐시 미스: DB에서 가져오기
    data, err := r.base.Get(id)
    if err != nil {
        return nil, err
    }
    
    // 캐시 업데이트
    r.mu.Lock()
    r.cache[id] = &CacheEntry{
        data:      data,
        expiresAt: time.Now().Add(r.ttl),
    }
    r.mu.Unlock()
    
    return data, nil
}
```

## 에러 처리 패턴

### 1. Error Types
```go
type ErrorCode string

const (
    ErrNotFound      ErrorCode = "NOT_FOUND"
    ErrInvalidInput  ErrorCode = "INVALID_INPUT"
    ErrUnauthorized  ErrorCode = "UNAUTHORIZED"
    ErrInternal      ErrorCode = "INTERNAL"
)

type AppError struct {
    Code      ErrorCode
    Message   string
    Details   map[string]interface{}
    Cause     error
    Timestamp time.Time
}

func (e AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
```

### 2. Error Handler
```go
type ErrorHandler struct {
    logger Logger
}

func (h *ErrorHandler) Handle(err error) tea.Cmd {
    var appErr AppError
    if errors.As(err, &appErr) {
        h.logger.Error("Application error", 
            "code", appErr.Code,
            "message", appErr.Message,
            "details", appErr.Details,
        )
        
        // UI에 에러 메시지 전달
        return func() tea.Msg {
            return ErrorMsg{Error: appErr}
        }
    }
    
    // 예상치 못한 에러
    h.logger.Error("Unexpected error", "error", err)
    return func() tea.Msg {
        return ErrorMsg{
            Error: AppError{
                Code:    ErrInternal,
                Message: "An unexpected error occurred",
                Cause:   err,
            },
        }
    }
}

## 2025-01-07 추가 패턴

### Provider Registry Pattern
```go
type Registry struct {
    factories map[string]func() Provider
    mu        sync.RWMutex
}

func (r *Registry) RegisterFactory(name string, factory func() Provider) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.factories[name]; exists {
        return fmt.Errorf("provider %s already registered", name)
    }
    
    r.factories[name] = factory
    return nil
}
```
**사용 예**:
- 백업 provider 등록
- 복원 provider 등록
- 동적 provider 로딩

### Adapter Pattern for CLI
```go
type BackupAdapter struct {
    registry *backup.Registry
}

func (a *BackupAdapter) ExecuteBackup(providerName string, cmd *cobra.Command, args []string) error {
    // 1. Provider 생성
    provider, err := a.registry.Create(providerName)
    
    // 2. 옵션 빌드
    opts, err := a.buildOptions(providerName, cmd, args)
    
    // 3. 실행
    return provider.Execute(context.Background(), opts)
}
```
**장점**:
- CLI와 도메인 로직 분리
- 테스트 용이
- 다양한 CLI 프레임워크 지원 가능

### Metadata Store Pattern
```go
type Store interface {
    Save(metadata *Metadata) error
    Get(id string) (*Metadata, error)
    List() ([]*Metadata, error)
    Delete(id string) error
}

type FileStore struct {
    baseDir string
    mu      sync.RWMutex
}
```
**특징**:
- 인터페이스 기반
- 파일 시스템 구현
- 향후 DB로 교체 가능

### Progress Monitoring Pattern
```go
func (a *Adapter) monitorProgress(provider Provider, done <-chan bool) {
    progressCh := provider.StreamProgress()
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-done:
            return
        case progress := <-progressCh:
            // Update display
        case <-ticker.C:
            // Refresh display
        }
    }
}
```
**용도**:
- 비동기 진행률 표시
- 실시간 업데이트
- 리소스 효율적

### Domain Model Separation
```go
// Domain layer
internal/domain/
├── backup/
│   ├── provider.go
│   ├── types.go
│   └── registry.go
├── restore/
│   ├── provider.go
│   ├── types.go
│   └── registry.go
├── metadata/
│   └── store.go
└── job/
    ├── job.go
    ├── repository.go
    └── status.go

// Infrastructure layer
internal/infrastructure/
├── kubernetes/
│   ├── client.go
│   └── executor.go
├── process/
│   └── executor.go
└── storage/
    └── file_repository.go

// Application layer
cmd/cli-recover/adapters/
├── backup_adapter.go
├── restore_adapter.go
├── list_adapter.go
└── job_adapter.go
```
**이점**:
- 명확한 계층 분리
- 의존성 방향 제어
- 도메인 로직 보호

## 2025-01-07 백그라운드 실행 패턴

### Job Domain Model
```go
type Job struct {
    ID          string
    PID         int        // 프로세스 ID
    Command     string
    Args        []string
    Status      JobStatus
    Progress    int
    Output      *RingBuffer // 메모리 효율적 출력 관리
    StartTime   time.Time
    EndTime     time.Time
    Error       error
}

type JobStatus string

const (
    JobStatusPending   JobStatus = "pending"
    JobStatusRunning   JobStatus = "running"
    JobStatusCompleted JobStatus = "completed"
    JobStatusFailed    JobStatus = "failed"
    JobStatusCancelled JobStatus = "cancelled"
)
```

### Background Execution Pattern
```go
func (s *JobService) StartBackground(cmd *cobra.Command, args []string) (*Job, error) {
    // 1. Job 생성
    job := &Job{
        ID:      generateJobID(),
        Command: os.Args[0],
        Args:    append(args, "--job-id", job.ID),
        Status:  JobStatusPending,
    }
    
    // 2. Job 저장
    if err := s.repo.Save(job); err != nil {
        return nil, err
    }
    
    // 3. 백그라운드 프로세스 시작
    process := exec.Command(job.Command, job.Args...)
    process.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,
    }
    
    if err := process.Start(); err != nil {
        return nil, err
    }
    
    // 4. PID 저장
    job.PID = process.Process.Pid
    job.Status = JobStatusRunning
    s.repo.Update(job)
    
    return job, nil
}
```

### File-based Job Repository
```go
type FileJobRepository struct {
    baseDir string
}

func (r *FileJobRepository) Save(job *Job) error {
    data, err := json.Marshal(job)
    if err != nil {
        return err
    }
    
    path := filepath.Join(r.baseDir, "jobs", job.ID+".json")
    return os.WriteFile(path, data, 0644)
}

func (r *FileJobRepository) GetByPID(pid int) (*Job, error) {
    files, err := os.ReadDir(filepath.Join(r.baseDir, "jobs"))
    if err != nil {
        return nil, err
    }
    
    for _, file := range files {
        job, err := r.load(file.Name())
        if err != nil {
            continue
        }
        if job.PID == pid {
            return job, nil
        }
    }
    
    return nil, ErrJobNotFound
}
```

### Cleanup Pattern
```go
type CleanupService struct {
    baseDir string
    logger  Logger
}

func (s *CleanupService) CleanOlderThan(duration time.Duration) error {
    cutoff := time.Now().Add(-duration)
    
    // Clean logs
    logDir := filepath.Join(s.baseDir, "logs")
    if err := s.cleanDirectory(logDir, cutoff); err != nil {
        s.logger.Error("Failed to clean logs", "error", err)
    }
    
    // Clean completed jobs
    jobDir := filepath.Join(s.baseDir, "jobs")
    if err := s.cleanCompletedJobs(jobDir, cutoff); err != nil {
        s.logger.Error("Failed to clean jobs", "error", err)
    }
    
    return nil
}
```
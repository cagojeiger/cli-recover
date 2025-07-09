# 코딩 패턴 및 컨벤션

## 아키텍처 패턴

### 헥사고날 아키텍처 (Ports & Adapters)
```go
// Port (도메인에서 정의)
type BackupRepository interface {
    Save(job *BackupJob) error
    FindByID(id string) (*BackupJob, error)
    List() ([]*BackupJob, error)
}

// Adapter (인프라에서 구현)
type FileBackupRepository struct {
    basePath string
}

func (r *FileBackupRepository) Save(job *BackupJob) error {
    // 파일 시스템에 저장
}
```

### Component 패턴 (UI)
```go
type Component interface {
    // Bubble Tea 인터페이스 준수
    Init() tea.Cmd
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
    
    // 컴포넌트 라이프사이클
    Focus() tea.Cmd
    Blur() tea.Cmd
    SetSize(width, height int)
}

// 예시: 리스트 컴포넌트
type ListComponent[T any] struct {
    items    []T
    selected int
    focused  bool
    renderer func(T, bool) string
}
```

### Repository 패턴
```go
// 도메인 모델과 영속성 분리
type JobRepository interface {
    Create(job *BackupJob) error
    Update(job *BackupJob) error
    Delete(id string) error
    FindByID(id string) (*BackupJob, error)
    FindAll() ([]*BackupJob, error)
    FindByStatus(status JobStatus) ([]*BackupJob, error)
}
```

### Command 패턴
```go
// 사용자 액션을 명령으로 캡슐화
type Command interface {
    Execute(ctx context.Context) error
    Undo() error
    CanExecute() bool
}

type CreateBackupCommand struct {
    service BackupService
    request BackupRequest
}
```

### Builder 패턴
```go
// 복잡한 객체 생성을 단계별로
type BackupJobBuilder struct {
    job *BackupJob
}

func NewBackupJobBuilder() *BackupJobBuilder {
    return &BackupJobBuilder{
        job: &BackupJob{},
    }
}

func (b *BackupJobBuilder) WithNamespace(ns string) *BackupJobBuilder {
    b.job.Namespace = ns
    return b
}

func (b *BackupJobBuilder) Build() (*BackupJob, error) {
    // 유효성 검사 후 반환
}
```

### Ring Buffer 패턴
```go
// 메모리 효율적인 순환 버퍼
type RingBuffer struct {
    data     []string
    size     int
    writePos int
    readPos  int
    mu       sync.RWMutex
}

func (rb *RingBuffer) Write(line string) {
    rb.mu.Lock()
    defer rb.mu.Unlock()
    
    rb.data[rb.writePos] = line
    rb.writePos = (rb.writePos + 1) % rb.size
}
```

### Factory 패턴
```go
// 백업 타입별 생성 로직 캡슐화
type BackupTypeFactory interface {
    Create(typeName string) (BackupType, error)
}

type DefaultBackupTypeFactory struct {
    registry map[string]func() BackupType
}

func (f *DefaultBackupTypeFactory) Register(name string, creator func() BackupType) {
    f.registry[name] = creator
}
```

### Strategy 패턴
```go
// 알고리즘 교체 가능
type CompressionStrategy interface {
    Compress(data []byte) ([]byte, error)
    Extension() string
}

type GzipStrategy struct{}
type BzipStrategy struct{}
type NoCompressionStrategy struct{}
```

### Observer 패턴 (이벤트 기반)
```go
// 상태 변화 알림
type EventBus interface {
    Subscribe(eventType EventType, handler EventHandler)
    Publish(event Event)
}

type EventHandler func(event Event)

type Event struct {
    Type      EventType
    Timestamp time.Time
    Data      interface{}
}
```

## 코딩 컨벤션

### 에러 처리
```go
// 도메인 에러 정의
type BackupError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

func (e BackupError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

// 사용 예시
if err := validatePath(path); err != nil {
    return BackupError{
        Code:    InvalidPath,
        Message: "invalid backup path",
        Cause:   err,
    }
}
```

### 네이밍 규칙
- 인터페이스: 동사+er (예: Reader, Writer, BackupExecutor)
- 구현체: 형용사+인터페이스명 (예: FileBackupRepository)
- 메서드: 동사로 시작 (예: CreateBackup, ValidateOptions)
- 상수: 대문자 스네이크 케이스 (예: MAX_BUFFER_SIZE)

### 테스트 작성
```go
// 테이블 주도 테스트
func TestRingBuffer_Write(t *testing.T) {
    tests := []struct {
        name     string
        size     int
        writes   []string
        expected []string
    }{
        {
            name:     "normal write",
            size:     3,
            writes:   []string{"a", "b", "c"},
            expected: []string{"a", "b", "c"},
        },
        {
            name:     "overflow write",
            size:     3,
            writes:   []string{"a", "b", "c", "d"},
            expected: []string{"b", "c", "d"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            rb := NewRingBuffer(tt.size)
            for _, w := range tt.writes {
                rb.Write(w)
            }
            assert.Equal(t, tt.expected, rb.GetAll())
        })
    }
}
```

### 의존성 주입
```go
// 생성자에서 의존성 주입
func NewBackupService(
    repo BackupRepository,
    executor CommandExecutor,
    logger Logger,
) *BackupService {
    return &BackupService{
        repo:     repo,
        executor: executor,
        logger:   logger,
    }
}
```

## 2025-01-09 현재 적용 패턴

### Factory 패턴 (Registry 단순화)
```go
// 단순한 Factory 함수로 변경
func CreateBackupProvider(name string) (backup.Provider, error) {
    switch name {
    case "filesystem":
        return filesystem.NewProvider(executor.New()), nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}
```

### Adapter 패턴 (CLI 통합)
```go
// CLI → Domain 브릿지
type BackupAdapter struct {
    registry *backup.Registry
}

func (a *BackupAdapter) ExecuteBackup(providerName string, cmd *cobra.Command, args []string) error {
    provider := a.registry.Create(providerName)
    opts := a.buildOptions(cmd, args)
    return provider.Execute(ctx, opts)
}
```

### Progress Streaming 패턴
```go
// 비동기 진행률 처리
func monitorProgress(provider Provider, done <-chan bool) {
    progressCh := provider.StreamProgress()
    ticker := time.NewTicker(500 * time.Millisecond)
    
    for {
        select {
        case <-done:
            return
        case progress := <-progressCh:
            updateDisplay(progress)
        }
    }
}
```

### Metadata Store 인터페이스
```go
// 저장소 추상화
type Store interface {
    Save(metadata *Metadata) error
    Get(id string) (*Metadata, error)
    List() ([]*Metadata, error)
    Delete(id string) error
}

// 파일 시스템 구현
type FileStore struct {
    baseDir string
    mu      sync.RWMutex
}
```

### TDD 패턴
```go
// Mock 기반 테스트
type MockProvider struct {
    mock.Mock
}

func TestAdapter_Execute(t *testing.T) {
    // Given
    mockProvider := new(MockProvider)
    mockProvider.On("Execute", mock.Anything, opts).Return(nil)
    
    // When
    err := adapter.Execute(cmd, args)
    
    // Then
    assert.NoError(t, err)
    mockProvider.AssertExpectations(t)
}
```

### 안전한 리팩토링 패턴
```go
// 1. 호환성 테스트 먼저 작성
func TestBackupCompatibility(t *testing.T) {
    oldCmd := newFilesystemBackupCmd()
    newCmd := newBackupCommand()
    
    // 기능 동일성 검증
    assert.Equal(t, oldCmd.Use, newCmd.Use)
    // 모든 플래그 검증
    for _, flag := range oldFlags {
        assert.NotNil(t, newCmd.Flags().Lookup(flag))
    }
}

// 2. 테스트 통과 후 안전하게 제거
// 3. 점진적 마이그레이션
```

### 테스트 커버리지 개선 패턴
```go
// 누락된 함수 찾기
func TestNewRestoreAdapter(t *testing.T) {
    // 생성자 테스트
    adapter := NewRestoreAdapter(registry)
    assert.NotNil(t, adapter)
}

// Edge case 테스트
func TestSanitizeTargetPath(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty path", "", "/"},
        {"with spaces", "/my data", "/my data"},
        // 실제 동작 기반 테스트
    }
}
```

## CLI 디자인 패턴 (2025-01-09)

### 플래그 레지스트리 패턴
```go
// 중앙 집중식 플래그 관리
var Registry = struct {
    Namespace   string
    Output      string
    Force       string
    Compression string
    // ...
}{
    Namespace:   "n",
    Output:      "o",
    Force:       "f",
    Compression: "c",
}

// 컴파일 타임 검증
func init() {
    if err := validateNoDuplicates(); err != nil {
        panic(fmt.Sprintf("Flag conflict: %v", err))
    }
}
```

### 하이브리드 인자 처리 패턴
```go
// Positional args + flags 처리
func buildOptions(cmd *cobra.Command, args []string) (Options, error) {
    opts := Options{}
    
    // 1. Positional args 먼저
    if len(args) >= 1 {
        opts.Pod = args[0]
    }
    if len(args) >= 2 {
        opts.Path = args[1]
    }
    
    // 2. Flags로 오버라이드
    if pod, _ := cmd.Flags().GetString("pod"); pod != "" {
        opts.Pod = pod
    }
    
    return opts, nil
}
```

### 구조화된 에러 처리 패턴
```go
// 사용자 친화적 에러
type CLIError struct {
    What   string
    Why    string
    How    string
    SeeDoc string
}

func handleError(err error) {
    var cliErr *CLIError
    if errors.As(err, &cliErr) {
        fmt.Fprintf(os.Stderr, "❌ Error: %s\n", cliErr.What)
        fmt.Fprintf(os.Stderr, "   Reason: %s\n", cliErr.Why)
        fmt.Fprintf(os.Stderr, "   Fix: %s\n", cliErr.How)
        if cliErr.SeeDoc != "" {
            fmt.Fprintf(os.Stderr, "   See: %s\n", cliErr.SeeDoc)
        }
    }
}
```

### 진행률 통합 패턴
```go
// 백업 파일 쓰기와 진행률 통합
type backupProgressWriter struct {
    writer   io.Writer
    current  int64
    total    int64
    reporter backup.ProgressReporter
}

func (w *backupProgressWriter) Write(p []byte) (n int, err error) {
    n, err = w.writer.Write(p)
    w.current += int64(n)
    
    // 진행률 보고
    w.reporter.Report(backup.Progress{
        BytesProcessed: w.current,
        TotalBytes:     w.total,
        CurrentFile:    w.currentFile,
    })
    
    return n, err
}
```

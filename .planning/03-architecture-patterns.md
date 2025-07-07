# 아키텍처 패턴 설계

## 목표
- Kubernetes Pod 백업 도구의 확장성과 유지보수성 확보
- kubectl 의존성 격리
- 테스트 용이성 향상
- 새로운 백업 타입 추가 용이성

## 핵심 패턴: Hexagonal Architecture + Plugin Pattern

### 1. 전체 구조

```
┌─────────────────────────────────────────────────────────┐
│                    Presentation Layer                    │
│                   (Bubble Tea TUI)                      │
└─────────────────────────┬───────────────────────────────┘
                          │ Port (Interface)
┌─────────────────────────▼───────────────────────────────┐
│                    Domain Core                          │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐ │
│  │BackupService│  │ JobManager   │  │ BackupType    │ │
│  │             │  │              │  │ (Plugin)      │ │
│  └─────────────┘  └──────────────┘  └───────────────┘ │
└─────────────────────────┬───────────────────────────────┘
                          │ Port (Interface)
┌─────────────────────────▼───────────────────────────────┐
│                 Infrastructure Adapters                  │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐ │
│  │ Kubectl     │  │ FileSystem   │  │ Config        │ │
│  │ Adapter     │  │ Adapter      │  │ Adapter       │ │
│  └─────────────┘  └──────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### 2. 핵심 인터페이스 (Ports)

#### Domain Core Interfaces
```go
// 백업 서비스 포트
type BackupService interface {
    CreateBackup(ctx context.Context, req BackupRequest) (*Job, error)
    ListJobs() ([]*Job, error)
    GetJob(id string) (*Job, error)
    CancelJob(id string) error
}

// Kubernetes 접근 포트
type KubernetesPort interface {
    GetNamespaces() ([]string, error)
    GetPods(namespace string) ([]Pod, error)
    GetContainers(namespace, pod string) ([]string, error)
    ExecInPod(namespace, pod, container string, cmd []string) (io.ReadCloser, error)
}

// 작업 저장소 포트
type JobRepository interface {
    Save(job *Job) error
    Update(job *Job) error
    FindByID(id string) (*Job, error)
    FindAll() ([]*Job, error)
}

// 백업 타입 인터페이스 (플러그인)
type BackupType interface {
    Name() string
    Description() string
    BuildCommand(target Target, options Options) []string
    ValidateOptions(options Options) error
    GetOptionsForm() []FormField
}
```

### 3. Plugin Pattern for 백업 타입

```go
// 백업 타입 레지스트리
type BackupTypeRegistry struct {
    mu    sync.RWMutex
    types map[string]BackupType
}

func (r *BackupTypeRegistry) Register(bt BackupType) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.types[bt.Name()] = bt
}

func (r *BackupTypeRegistry) Get(name string) (BackupType, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    bt, ok := r.types[name]
    return bt, ok
}

// 초기화 시 등록
func InitBackupTypes(registry *BackupTypeRegistry) {
    registry.Register(NewFilesystemBackup())
    registry.Register(NewMinIOBackup())
    registry.Register(NewMongoDBBackup())
}
```

### 4. Event-Driven Job Management

```go
// 작업 이벤트
type JobEvent interface {
    JobID() string
    Type() EventType
    Timestamp() time.Time
}

type EventType int

const (
    JobCreated EventType = iota
    JobStarted
    JobProgress
    JobCompleted
    JobFailed
    JobCancelled
)

// 이벤트 버스
type JobEventBus interface {
    Subscribe(handler func(JobEvent)) func() // unsubscribe 함수 반환
    Publish(event JobEvent)
}

// 구현 예시
type jobEventBus struct {
    mu          sync.RWMutex
    handlers    []func(JobEvent)
}

func (eb *jobEventBus) Subscribe(handler func(JobEvent)) func() {
    eb.mu.Lock()
    eb.handlers = append(eb.handlers, handler)
    index := len(eb.handlers) - 1
    eb.mu.Unlock()
    
    // unsubscribe 함수 반환
    return func() {
        eb.mu.Lock()
        defer eb.mu.Unlock()
        eb.handlers = append(eb.handlers[:index], eb.handlers[index+1:]...)
    }
}
```

### 5. Adapter 구현 예시

```go
// Kubectl Adapter
type kubectlAdapter struct {
    kubectlPath string
    timeout     time.Duration
}

func (k *kubectlAdapter) GetNamespaces() ([]string, error) {
    cmd := exec.Command(k.kubectlPath, "get", "namespaces", "-o", "json")
    // ... 실행 및 파싱
}

func (k *kubectlAdapter) ExecInPod(namespace, pod, container string, command []string) (io.ReadCloser, error) {
    args := []string{"exec", "-n", namespace}
    if container != "" {
        args = append(args, "-c", container)
    }
    args = append(args, pod, "--")
    args = append(args, command...)
    
    cmd := exec.Command(k.kubectlPath, args...)
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()
    
    return stdout, nil
}
```

### 6. 의존성 주입 구조

```go
// 애플리케이션 초기화
type App struct {
    // Ports
    backupService BackupService
    kubeClient    KubernetesPort
    jobRepo       JobRepository
    
    // Infrastructure
    eventBus      JobEventBus
    registry      *BackupTypeRegistry
}

func NewApp() *App {
    // Infrastructure 생성
    kubeAdapter := NewKubectlAdapter()
    fileRepo := NewFileJobRepository("~/.cli-recover/jobs")
    eventBus := NewJobEventBus()
    registry := NewBackupTypeRegistry()
    
    // 백업 타입 등록
    InitBackupTypes(registry)
    
    // Domain 서비스 생성
    jobManager := NewJobManager(fileRepo, eventBus)
    backupService := NewBackupService(kubeAdapter, jobManager, registry)
    
    return &App{
        backupService: backupService,
        kubeClient:    kubeAdapter,
        jobRepo:       fileRepo,
        eventBus:      eventBus,
        registry:      registry,
    }
}
```

## 장점

1. **테스트 용이성**: 모든 외부 의존성이 인터페이스로 모킹 가능
2. **확장성**: 새 백업 타입을 플러그인으로 추가
3. **유지보수성**: 각 레이어의 책임이 명확
4. **kubectl 독립성**: adapter 교체로 다른 런타임 지원 가능

## 복잡도 평가: 40/100 ⚠️
- 필요한 추상화만 도입
- 실용적인 수준의 분리
- 과도한 인터페이스 지양
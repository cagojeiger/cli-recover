# 재사용 가능한 패턴

## 비즈니스 로직 패턴

### Service Layer
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

### Repository with Cache
```go
type CachedRepository struct {
    base  Repository
    cache map[string]*CacheEntry
    mu    sync.RWMutex
    ttl   time.Duration
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

### Error Types
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

## Provider Registry Pattern
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

## Adapter Pattern for CLI
```go
type BackupAdapter struct {
    registry *backup.Registry
    logger   logger.Logger
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

## Metadata Store Pattern
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

## Progress Monitoring Pattern
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

## Logger Integration Pattern

### Provider에 로거 추가
```go
type Provider struct {
    executor Executor
    logger   logger.Logger
}

func NewProvider(executor Executor) *Provider {
    return &Provider{
        executor: executor,
        logger:   log.WithField("provider", "filesystem"),
    }
}

func (p *Provider) Execute(ctx context.Context, opts backup.Options) error {
    p.logger.Info("Starting filesystem backup", 
        log.F("namespace", opts.Namespace),
        log.F("pod", opts.PodName),
        log.F("path", opts.Path),
    )
    
    // 백업 실행...
    
    if err != nil {
        p.logger.Error("Backup failed", 
            log.F("error", err),
            log.F("pod", opts.PodName),
        )
        return err
    }
    
    p.logger.Info("Backup completed successfully",
        log.F("size", size),
        log.F("duration", time.Since(start)),
    )
    
    return nil
}
```

### CLI 플래그로 로거 설정
```go
rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
    level, _ := cmd.Flags().GetString("log-level")
    output, _ := cmd.Flags().GetString("log-output")
    logFile, _ := cmd.Flags().GetString("log-file")
    
    cfg := logger.DefaultConfig()
    cfg.Level = level
    cfg.Output = output
    if logFile != "" {
        cfg.FilePath = logFile
    }
    
    if err := logger.InitializeFromConfig(cfg); err != nil {
        logger.Error("Failed to configure logger", logger.F("error", err))
    }
}
```

### 조건부 로깅
```go
// Verbose 모드에서만 로깅
if logger.GetLevel() <= logger.DebugLevel {
    logger.Debug("Detailed information", 
        logger.F("data", complexData),
    )
}

// 대량 데이터는 요약
logger.Info("Processing items", 
    logger.F("count", len(items)),
    logger.F("first", items[0]),
)
```

## Job Domain Model Pattern
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

## Background Execution Pattern
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

## File-based Job Repository
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

## Cleanup Pattern
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
# Logger Integration Examples

이 문서는 cli-recover 프로젝트에 로거를 통합하는 방법을 보여줍니다.

## 1. Provider에 로거 추가하기

### Provider 인터페이스 수정 (예시)

```go
// Provider defines the interface for backup providers
type Provider interface {
    // ... 기존 메서드들 ...
    
    // SetLogger sets the logger for the provider
    SetLogger(logger logger.Logger)
}
```

### Filesystem Provider에 로거 추가

```go
package filesystem

import (
    "github.com/cagojeiger/cli-recover/internal/domain/logger"
    log "github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
)

type Provider struct {
    executor Executor
    logger   logger.Logger
}

func NewProvider(executor Executor) *Provider {
    return &Provider{
        executor: executor,
        logger:   log.WithField("provider", "filesystem"), // 기본 로거
    }
}

func (p *Provider) SetLogger(l logger.Logger) {
    p.logger = l.WithField("provider", "filesystem")
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

## 2. CLI 명령에 로거 통합

### 메인 함수에서 로거 초기화

```go
package main

import (
    "github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
    "github.com/spf13/cobra"
)

func main() {
    // 환경 변수에서 로거 설정 로드
    if err := logger.InitializeFromEnv(); err != nil {
        // 초기화 실패 시 기본 로거 사용
        logger.Error("Failed to initialize logger from env", logger.F("error", err))
    }
    
    rootCmd := &cobra.Command{
        Use:   "cli-recover",
        Short: "Kubernetes backup and restore tool",
    }
    
    // 글로벌 플래그 추가
    rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
    rootCmd.PersistentFlags().String("log-output", "console", "Log output (console, file, both)")
    rootCmd.PersistentFlags().String("log-file", "", "Log file path")
    
    // 플래그 기반 로거 재설정
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
    
    // ... 명령 추가 ...
    
    if err := rootCmd.Execute(); err != nil {
        logger.Fatal("Command failed", logger.F("error", err))
    }
}
```

### Backup 명령에서 로거 사용

```go
func NewBackupCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "backup",
        Short: "Perform a backup",
        RunE: func(cmd *cobra.Command, args []string) error {
            logger := log.WithField("command", "backup")
            logger.Debug("Starting backup command")
            
            // 옵션 파싱
            namespace, _ := cmd.Flags().GetString("namespace")
            
            // 백업 서비스 생성
            service := backup.NewService(
                backup.DefaultRegistry,
                metadata.NewFileStorage(metadataPath),
            )
            
            // 백업 실행
            logger.Info("Executing backup", 
                log.F("type", backupType),
                log.F("namespace", namespace),
            )
            
            result, err := service.CreateBackup(ctx, req)
            if err != nil {
                logger.Error("Backup failed", log.F("error", err))
                return err
            }
            
            logger.Info("Backup completed", 
                log.F("id", result.ID),
                log.F("size", result.Size),
            )
            
            return nil
        },
    }
    
    return cmd
}
```

## 3. 서비스 레이어에 로거 추가

```go
package backup

import (
    "github.com/cagojeiger/cli-recover/internal/domain/logger"
    log "github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
)

type Service struct {
    registry *ProviderRegistry
    storage  metadata.Storage
    logger   logger.Logger
}

func NewService(registry *ProviderRegistry, storage metadata.Storage) *Service {
    return &Service{
        registry: registry,
        storage:  storage,
        logger:   log.WithField("service", "backup"),
    }
}

func (s *Service) CreateBackup(ctx context.Context, req CreateBackupRequest) (*CreateBackupResult, error) {
    s.logger.Debug("Creating backup", 
        log.F("type", req.Type),
        log.F("namespace", req.Namespace),
    )
    
    // Provider 가져오기
    provider, err := s.registry.Get(req.Type)
    if err != nil {
        s.logger.Error("Failed to get provider", 
            log.F("type", req.Type),
            log.F("error", err),
        )
        return nil, err
    }
    
    // Provider에 로거 설정
    if p, ok := provider.(interface{ SetLogger(logger.Logger) }); ok {
        p.SetLogger(s.logger)
    }
    
    // ... 백업 실행 ...
    
    return result, nil
}
```

## 4. 기존 fmt.Printf 대체

### Before:
```go
fmt.Printf("Starting backup for pod %s in namespace %s\n", podName, namespace)
fmt.Printf("Error: %v\n", err)
```

### After:
```go
logger.Info("Starting backup", 
    logger.F("pod", podName),
    logger.F("namespace", namespace),
)
logger.Error("Operation failed", logger.F("error", err))
```

## 5. 조건부 로깅

```go
// Verbose 모드에서만 로깅
if logger.GetLevel() <= logger.DebugLevel {
    logger.Debug("Detailed information", 
        logger.F("data", complexData),
    )
}

// 진행 상황 로깅
logger.WithField("progress", progress).Info("Backup progress")
```

## 6. 테스트에서 로거 사용

```go
func TestBackupService(t *testing.T) {
    // 테스트용 로거 생성 (출력 없음)
    testLogger := logger.NewConsoleLogger(logger.FatalLevel, false)
    logger.SetGlobalLogger(testLogger)
    
    // 또는 특정 서비스에만 주입
    service := NewService(registry, storage)
    service.SetLogger(testLogger)
    
    // 테스트 실행...
}
```

## 7. 로그 출력 예시

### Console (기본):
```
2025-01-07 14:30:15.123 [INFO ] Starting backup pod=nginx-pod namespace=default
2025-01-07 14:30:15.456 [DEBUG] Executing tar command cmd="kubectl exec nginx-pod -- tar -czf - /data"
2025-01-07 14:30:20.789 [INFO ] Backup completed size=1048576 duration=5.666s
```

### JSON (파일 로깅):
```json
{"timestamp":"2025-01-07T14:30:15.123Z","level":"INFO","message":"Starting backup","pod":"nginx-pod","namespace":"default"}
{"timestamp":"2025-01-07T14:30:20.789Z","level":"INFO","message":"Backup completed","size":1048576,"duration":"5.666s"}
```

## 8. 환경 변수 설정

```bash
# 로그 레벨 설정
export CLI_RECOVER_LOG_LEVEL=debug

# 파일로 로깅
export CLI_RECOVER_LOG_OUTPUT=file
export CLI_RECOVER_LOG_FILE=/var/log/cli-recover.log

# JSON 포맷 사용
export CLI_RECOVER_LOG_FORMAT=json

# 콘솔 색상 비활성화
export CLI_RECOVER_LOG_COLOR=false
```

## 9. 로그 로테이션

로그 파일은 자동으로 로테이션됩니다:
- 기본 최대 크기: 100MB
- 기본 보관 기간: 7일
- 로테이션된 파일명: `cli-recover.log.20250107-143015`

## 10. 성능 고려사항

```go
// 비싼 연산은 필요할 때만 수행
if logger.GetLevel() <= logger.DebugLevel {
    // Debug 레벨일 때만 실행
    expensiveData := computeExpensiveData()
    logger.Debug("Expensive data", logger.F("data", expensiveData))
}

// 대량 데이터는 요약
logger.Info("Processing items", 
    logger.F("count", len(items)),
    logger.F("first", items[0]),
)
```

이러한 패턴을 따라 점진적으로 로거를 통합하면 됩니다.
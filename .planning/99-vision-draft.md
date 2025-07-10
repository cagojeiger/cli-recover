# CLI-Recover 2.0 Vision Draft

## Purpose
이 문서는 CLI-Recover의 새로운 아키텍처 비전을 담은 임시 문서입니다.
기존 컨텍스트와 병합되기 전의 실험적 아이디어를 자유롭게 탐색합니다.

## 핵심 통찰
- **Filesystem restore ≈ cp**: tar 복원은 본질적으로 파일 복사
- **Provider 다양성**: MongoDB, MinIO, PostgreSQL은 완전히 다른 패러다임
- **공통 인터페이스의 한계**: 강제된 추상화는 오히려 복잡도 증가
- **CLI/TUI 통합 가능성**: 메타데이터 기반 자동 UI 생성

## 새로운 비전

### 1. Provider 독립 아키텍처
```
internal/
├── providers/
│   ├── filesystem/
│   │   ├── tar_backup.go      # tar 기반 백업
│   │   ├── cp_restore.go      # cp 스타일 복원
│   │   ├── cli_definition.go  # CLI 명령 정의
│   │   └── README.md          # Provider 특화 문서
│   │
│   ├── mongodb/
│   │   ├── mongodump.go       # mongodump 래핑
│   │   ├── mongorestore.go    # mongorestore 래핑
│   │   ├── cli_definition.go  # MongoDB 특화 옵션
│   │   └── README.md
│   │
│   └── minio/
│       ├── s3_backup.go       # S3 API 활용
│       ├── s3_sync.go         # 동기화 방식 복원
│       ├── cli_definition.go  # S3 특화 옵션
│       └── README.md
```

### 2. CLI/TUI 통합 시스템

#### CLI Definition 표준
```go
type CLIDefinition struct {
    Provider string
    Commands []CommandDef
}

type CommandDef struct {
    Name        string
    Description string
    Args        []ArgDef
    Flags       []FlagDef
    Handler     HandlerFunc
    Validator   ValidatorFunc
}

type FlagDef struct {
    Name        string
    Type        string // "string", "bool", "int", "duration"
    Default     interface{}
    Required    bool
    Description string
    Choices     []string // for enum types
}
```

#### 자동 TUI 생성
```go
// CLI 정의에서 TUI Form 자동 생성
func (cmd CommandDef) GenerateTUIForm() *tview.Form {
    form := tview.NewForm()
    
    // Args → Input Fields
    for _, arg := range cmd.Args {
        form.AddInputField(arg.Name, "", 40, nil, nil)
    }
    
    // Flags → 적절한 UI 컴포넌트
    for _, flag := range cmd.Flags {
        switch flag.Type {
        case "bool":
            form.AddCheckbox(flag.Name, flag.Default.(bool), nil)
        case "string":
            if len(flag.Choices) > 0 {
                form.AddDropDown(flag.Name, flag.Choices, 0, nil)
            } else {
                form.AddInputField(flag.Name, flag.Default.(string), 40, nil, nil)
            }
        }
    }
    
    return form
}
```

### 3. Provider 플러그인 시스템

#### Provider 인터페이스 최소화
```go
// 각 Provider는 최소한의 인터페이스만 구현
type Provider interface {
    Name() string
    CLIDefinition() CLIDefinition
}

// 실제 동작은 Provider별로 특화
type FilesystemProvider struct {
    kubectlClient KubectlClient
}

type MongoDBProvider struct {
    mongoClient MongoClient
}
```

### 4. 진행률 표시 통합
```go
// 모든 Provider가 공통으로 사용하는 진행률 시스템
type ProgressReporter interface {
    Start(operation string, total int64)
    Update(current int64, message string)
    Complete()
    Error(err error)
}

// CLI와 TUI 모두에서 동작
type UnifiedProgressReporter struct {
    cliReporter *CLIProgressBar
    tuiReporter *TUIProgressWidget
}
```

## 마이그레이션 전략

### Phase 4: 아키텍처 리팩토링
1. Provider 독립 구조로 전환
2. 기존 코드를 새 구조로 이동
3. 공통 부분만 shared 패키지로

### Phase 5: CLI/TUI 통합
1. CLI Definition 표준 구현
2. 자동 TUI 생성 시스템
3. 통합 테스트

### Phase 6: Provider 확장
1. MongoDB Provider
2. MinIO Provider
3. PostgreSQL Provider

## 장기 비전

### 1. 플러그인 에코시스템
- 사용자가 Provider 추가 가능
- Go 플러그인 또는 WASM
- Provider 마켓플레이스

### 2. 선언적 백업 정책
```yaml
# backup-policy.yaml
providers:
  - type: filesystem
    schedule: "0 2 * * *"
    retention: 7d
    targets:
      - namespace: production
        pods: app-*
        paths: ["/data", "/config"]
```

### 3. 중앙 관리 대시보드
- 모든 백업 상태 모니터링
- 복원 이력 추적
- 알림 시스템

## 위험 요소 및 대응

### 1. 기존 사용자 호환성
- v1 명령어 호환 레이어 제공
- 마이그레이션 가이드 제공

### 2. 복잡도 증가
- Provider별 독립은 코드 중복 가능
- 하지만 각 Provider 특성에 최적화 가능

### 3. 테스트 커버리지
- Provider별 독립 테스트
- 통합 테스트 자동화

## 결론
이 비전은 CLI-Recover를 단순한 백업 도구에서 
확장 가능한 데이터 보호 플랫폼으로 진화시킵니다.

각 Provider의 특성을 존중하면서도
사용자에게는 일관된 경험을 제공합니다.
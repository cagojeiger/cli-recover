# 명령어 패턴과 구조

## 기본 명령어 패턴

### POSIX/GNU 표준 패턴
```
APPNAME VERB NOUN [OPTIONS] [ARGUMENTS]
```

### cli-recover 패턴
```
cli-recover [COMMAND] [SUBCOMMAND] [ARGS] [FLAGS]
```

## 명령어 실행 플로우

### Backup 명령 실행 흐름
```mermaid
sequenceDiagram
    participant U as User
    participant C as CLI Parser
    participant V as Validator
    participant P as Provider
    participant K as Kubernetes
    participant F as File System
    
    U->>C: cli-recover backup filesystem nginx /data
    C->>C: Parse command & flags
    C->>V: Validate arguments
    V->>V: Check required fields
    V->>V: Validate flag combinations
    V-->>C: Validation result
    
    alt Validation Failed
        C-->>U: Error with usage help
    else Validation Passed
        C->>P: Create filesystem provider
        P->>P: Build options from args/flags
        P->>K: kubectl exec -n <ns> <pod> -- tar -cvf - <path>
        K->>K: Execute in pod
        K-->>P: Stream tar data
        
        par Progress Monitoring
            P-->>C: Progress updates
            C-->>U: Display progress bar
        and Data Writing
            P->>F: Write to temp file
            F->>F: Calculate checksum
            F->>F: Atomic rename
        end
        
        P-->>C: Result (size, checksum)
        C-->>U: Success summary
    end
```

### Restore 명령 실행 흐름
```mermaid
sequenceDiagram
    participant U as User
    participant C as CLI Parser
    participant V as Validator
    participant P as Provider
    participant K as Kubernetes
    participant F as File System
    
    U->>C: cli-recover restore filesystem nginx backup.tar
    C->>C: Parse command & flags
    C->>V: Validate arguments
    
    alt Missing --force flag
        V-->>C: Dangerous operation warning
        C-->>U: Show warning & suggest --force
    else Has --force flag
        V-->>C: Validation passed
        C->>P: Create restore provider
        P->>F: Validate backup file
        F-->>P: File exists & readable
        P->>K: cat backup.tar | kubectl exec -i <pod> -- tar -xvf -
        K-->>P: Stream status
        P-->>C: Progress updates
        C-->>U: Display progress
        P-->>C: Restore complete
        C-->>U: Success message
    end
```

## 플래그와 인자 처리 패턴

### 하이브리드 인자 처리
```mermaid
graph TD
    Start[명령 입력] --> Parse[인자 파싱]
    Parse --> Check{플래그와 Positional<br/>모두 있나?}
    
    Check -->|Yes| Priority[플래그 우선순위 적용]
    Check -->|No| CheckType{어떤 타입?}
    
    CheckType -->|플래그만| UseFlags[플래그 값 사용]
    CheckType -->|Positional만| UseArgs[Positional 사용]
    CheckType -->|둘 다 없음| Error[필수 인자 누락 에러]
    
    Priority --> Merge[값 병합]
    UseFlags --> Validate[유효성 검증]
    UseArgs --> Validate
    Merge --> Validate
    
    Validate -->|성공| Execute[명령 실행]
    Validate -->|실패| ShowHelp[도움말 표시]
```

### 플래그 우선순위 규칙
```mermaid
flowchart LR
    subgraph "우선순위 (높음 → 낮음)"
        E[환경변수] --> F[플래그]
        F --> P[Positional Args]
        P --> C[설정 파일]
        C --> D[기본값]
    end
```

## 명령어 계층 구조

### 명령어 트리
```mermaid
graph TD
    CLI[cli-recover] --> B[backup]
    CLI --> R[restore]
    CLI --> L[list]
    CLI --> LOG[logs]
    CLI --> I[init]
    CLI --> T[tui]
    
    B --> BFS[filesystem]
    B --> BMG[🔮 mongodb]
    B --> BMO[🔮 minio]
    
    R --> RFS[filesystem]
    R --> RMG[🔮 mongodb]
    R --> RMO[🔮 minio]
    
    L --> LB[backups]
    L --> LJ[🔮 jobs]
    
    LOG --> LLIST[list]
    LOG --> LSHOW[show]
    LOG --> LTAIL[tail]
    LOG --> LCLEAN[clean]
    
    style BMG stroke-dasharray: 5 5
    style BMO stroke-dasharray: 5 5
    style RMG stroke-dasharray: 5 5
    style RMO stroke-dasharray: 5 5
    style LJ stroke-dasharray: 5 5
```

## 에러 처리 패턴

### 에러 플로우
```mermaid
stateDiagram-v2
    [*] --> Parsing: User Input
    Parsing --> Validation: Parse Success
    Parsing --> ParseError: Parse Failed
    
    Validation --> Execution: Valid
    Validation --> ValidationError: Invalid
    
    Execution --> Success: No Error
    Execution --> RuntimeError: Error
    
    ParseError --> ShowUsage: Display Help
    ValidationError --> ShowExample: Show Correct Usage
    RuntimeError --> ShowFix: Suggest Solution
    
    ShowUsage --> [*]
    ShowExample --> [*]
    ShowFix --> [*]
    Success --> [*]
```

### 에러 메시지 구조
```mermaid
graph TD
    Error[에러 발생] --> Type{에러 타입}
    
    Type -->|Parse| P[구문 에러]
    Type -->|Validation| V[검증 에러]
    Type -->|Runtime| R[실행 에러]
    
    P --> P1[잘못된 플래그/인자]
    P --> P2[알 수 없는 명령]
    
    V --> V1[필수 인자 누락]
    V --> V2[잘못된 값]
    V --> V3[충돌하는 옵션]
    
    R --> R1[권한 문제]
    R --> R2[리소스 없음]
    R --> R3[네트워크 오류]
    
    P1 --> Fix1[올바른 사용법 예시]
    V1 --> Fix2[필수 인자 안내]
    R1 --> Fix3[권한 해결 방법]
```

## Cobra 명령 구조

### 명령어 정의 패턴
```go
// 기본 구조
cmd := &cobra.Command{
    Use:   "subcommand [args]",
    Short: "짧은 설명",
    Long:  `긴 설명...`,
    Args:  cobra.ExactArgs(2), // 인자 검증
    RunE:  executeFunction,     // 실행 함수
}

// 플래그 추가
cmd.Flags().StringP("output", "o", "", "Output file")
cmd.Flags().BoolP("force", "f", false, "Force operation")
```

### 실행 함수 패턴
```go
func executeBackup(provider string, cmd *cobra.Command, args []string) error {
    // 1. 옵션 빌드
    opts := buildOptions(cmd, args)
    
    // 2. 검증
    if err := validateOptions(opts); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // 3. Provider 생성
    p := createProvider(provider)
    
    // 4. 실행
    result, err := p.Execute(context.Background(), opts)
    if err != nil {
        return fmt.Errorf("execution failed: %w", err)
    }
    
    // 5. 결과 표시
    displayResult(result)
    return nil
}
```

## 명령어 확장 패턴

### 새 Provider 추가 시
```mermaid
graph LR
    A[새 Provider 정의] --> B[Domain 인터페이스 구현]
    B --> C[Infrastructure 구현]
    C --> D[Factory 등록]
    D --> E[CLI 명령 추가]
    E --> F[테스트 작성]
    F --> G[문서 업데이트]
```

### 새 명령어 추가 시
```mermaid
flowchart TD
    A[요구사항 분석] --> B{명령 타입?}
    B -->|Action| C[동사형 명령]
    B -->|Query| D[명사형 명령]
    
    C --> E[backup, restore, init]
    D --> F[list, logs, status]
    
    E --> G[Provider 패턴 적용]
    F --> H[직접 구현]
    
    G --> I[플래그 정의]
    H --> I
    I --> J[로직 구현]
    J --> K[테스트 추가]
```

## 모범 사례

### 1. 명령어 이름
- 동작은 동사로: `backup`, `restore`, `init`
- 조회는 명사로: `list`, `logs`
- 명확하고 짧게: `list backups` (not `show-all-backup-files`)

### 2. 플래그 설계
- 짧은 형식은 자주 사용하는 것만
- 위험한 작업은 긴 형식만: `--force`, `--dry-run`
- Boolean 플래그는 긍정형으로: `--verbose` (not `--quiet`)

### 3. 인자 순서
- 중요도 순: `<필수> <선택>`
- 자연스러운 순서: `<source> <destination>`
- 일관성 유지: 모든 명령에서 동일한 순서

### 4. 에러 처리
- 명확한 에러 메시지
- 해결 방법 제시
- 관련 명령어 안내
- 종료 코드 일관성
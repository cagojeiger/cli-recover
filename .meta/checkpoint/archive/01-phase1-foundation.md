# Phase 0: Foundation - 완료 상태

## 목표
- 단일 Unix 명령 실행 및 로깅
- 기본 CLI 인터페이스
- 실행 이력 저장 및 조회

## 완성된 아키텍처

```
┌─────────────────────────────────────┐
│         CLI Interface               │
│     cli-pipe run "command"          │
└────────────────┬────────────────────┘
                 ▼
        ┌────────────────┐
        │  Application   │
        │   UseCase      │
        └────────┬───────┘
                 ▼
    ┌────────────┴────────────┐
    ▼                         ▼
┌─────────┐           ┌──────────────┐
│Executor │           │    Logger    │
│(os/exec)│           │ (JSON file)  │
└─────────┘           └──────────────┘
    │                         │
    ▼                         ▼
Unix Command            ~/.cli-pipe/
                         └── operations/
                             └── 2024-01-14/
                                 └── xxxxx.json
```

## 프로젝트 구조

```
cli-pipe/
├── cmd/
│   └── cli-pipe/
│       ├── main.go
│       ├── root.go
│       ├── run.go
│       ├── history.go
│       └── show.go
├── internal/
│   ├── domain/
│   │   ├── command.go      # Command 타입 정의
│   │   ├── operation.go    # Operation 타입 정의
│   │   └── errors.go       # 도메인 에러
│   ├── application/
│   │   ├── executor.go     # 실행 유즈케이스
│   │   └── query.go        # 조회 유즈케이스
│   └── infrastructure/
│       ├── process/
│       │   └── executor.go # Unix 명령 실행
│       └── storage/
│           └── logger.go   # JSON 파일 저장
├── go.mod
├── go.sum
└── Makefile
```

## 동작하는 기능

### 1. 명령 실행
```bash
$ cli-pipe run "ls -la"
Operation ID: 2024-01-14-123456-abc
Status: Running...

total 24
drwxr-xr-x  6 user  staff   192 Jan 14 10:30 .
drwxr-xr-x  5 user  staff   160 Jan 14 10:00 ..
-rw-r--r--  1 user  staff  1234 Jan 14 10:30 main.go

Status: Success
Duration: 0.12s
Exit Code: 0
```

### 2. 실행 이력 조회
```bash
$ cli-pipe history
ID                        Command         Status    Duration    Time
2024-01-14-123456-abc    ls -la          success   0.12s       10:30:45
2024-01-14-123455-xyz    pwd             success   0.01s       10:29:30
2024-01-14-123454-def    cat README.md   failed    0.05s       10:28:15
```

### 3. 상세 정보 조회
```bash
$ cli-pipe show 2024-01-14-123456-abc
Operation: 2024-01-14-123456-abc
Command: ls -la
Status: success
Started: 2024-01-14 10:30:45
Duration: 0.12s
Exit Code: 0

=== STDOUT ===
total 24
drwxr-xr-x  6 user  staff   192 Jan 14 10:30 .
...

=== STDERR ===
(empty)
```

### 4. 재실행
```bash
$ cli-pipe replay 2024-01-14-123456-abc
Replaying: ls -la
Operation ID: 2024-01-14-134567-new
Status: Running...
```

## 저장 형식

### Operation JSON
```json
{
  "id": "2024-01-14-123456-abc",
  "command": {
    "executable": "ls",
    "args": ["-la"],
    "raw": "ls -la"
  },
  "status": "success",
  "exit_code": 0,
  "started_at": "2024-01-14T10:30:45Z",
  "completed_at": "2024-01-14T10:30:45.12Z",
  "duration_ms": 120,
  "stdout": "total 24\ndrwxr-xr-x...",
  "stderr": "",
  "environment": {
    "pwd": "/Users/user/project",
    "user": "user"
  }
}
```

## 핵심 인터페이스

```go
// domain/operation.go
type Operation struct {
    ID          string
    Command     Command
    Status      Status
    ExitCode    int
    StartedAt   time.Time
    CompletedAt time.Time
    Stdout      string
    Stderr      string
}

// application/executor.go
type Executor interface {
    Execute(ctx context.Context, cmd string) (*Operation, error)
}

// application/query.go
type OperationQuery interface {
    FindByID(id string) (*Operation, error)
    List(filter Filter) ([]*Operation, error)
}
```

## 테스트 커버리지
- Domain: 100%
- Application: 95%
- Infrastructure: 85%
- Integration: 기본 시나리오

## 다음 Phase로의 연결점
- 단일 명령 → 여러 명령 순차 실행 (Pipeline)
- 하드코딩된 명령 → YAML 파일 정의
- 단순 실행 → 변수 지원
- 로컬 실행만 → 다양한 컨텍스트
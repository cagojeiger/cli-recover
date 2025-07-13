# cli-pipe 코딩 패턴

## 코드 구조 패턴
### 헥사고날 아키텍처
- **Domain**: 비즈니스 로직 (Pipeline, Operation)
- **Application**: 유즈케이스 (Execute, Replay)
- **Infrastructure**: 외부 연동 (Unix commands, Storage)

### 패키지 구조
```
internal/
  domain/         # 핵심 타입, 인터페이스
  application/    # 비즈니스 로직
  infrastructure/ # 구현체
    executor/     # 명령 실행
    storage/      # 로그 저장
cmd/
  cli-pipe/      # 메인 엔트리
```

## 파이프라인 정의 패턴
### 기본 구조
```yaml
pipeline:
  name: descriptive-name
  version: "1.0"
  steps:
    - command: "..."
      output: stream|file|variable
```

### 스트림 분기 패턴
```yaml
steps:
  - command: "source command"
    output: stream
    tee:
      - command: "sha256sum"
        output_to: context.checksum
      - command: "pv -pterb"
        output: stderr
      - command: "gzip"
        output: file
```

### 에러 처리 패턴
```yaml
steps:
  - command: "risky command"
    on_error: abort|continue|retry
    retry:
      count: 3
      delay: 5s
```

## Go 코딩 패턴
### 인터페이스 우선 설계
```go
// 인터페이스 정의
type Executor interface {
    Execute(ctx context.Context, cmd Command) (*Result, error)
}

// 구현은 infrastructure에
type UnixExecutor struct{}
```

### 에러 처리
```go
// 명확한 에러 타입
type PipelineError struct {
    Step   string
    Cause  error
    Output string
}

// 에러 래핑
if err != nil {
    return nil, fmt.Errorf("step %s failed: %w", step.Name, err)
}
```

### Context 사용
```go
// 타임아웃과 취소 지원
ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
defer cancel()
```

## 로깅 패턴
### 구조화된 로깅
```go
log.WithFields(log.Fields{
    "operation_id": op.ID,
    "step":        step.Name,
    "duration":    duration,
}).Info("Step completed")
```

### 로그 레벨
- **Debug**: 상세 실행 정보
- **Info**: 주요 이벤트
- **Warn**: 복구 가능한 문제
- **Error**: 실행 실패

## 스트림 처리 패턴
### TeeReader 사용
```go
// 읽으면서 동시에 처리
checksumReader := io.TeeReader(input, hasher)
progressReader := io.TeeReader(checksumReader, counter)
```

### 버퍼 관리
```go
// 고정 크기 버퍼 사용
buffer := make([]byte, 32*1024)
io.CopyBuffer(dst, src, buffer)
```

## 테스트 패턴
### 테이블 기반 테스트
```go
tests := []struct {
    name     string
    pipeline Pipeline
    want     Result
    wantErr  bool
}{
    // test cases...
}
```

### Mock 사용
```go
type MockExecutor struct {
    ExecuteFunc func(Command) (*Result, error)
}
```

## 설정 관리 패턴
### 우선순위
1. CLI 플래그
2. 환경 변수
3. 설정 파일
4. 기본값

### 네이밍 규칙
- 환경변수: `CLI_PIPE_*`
- 설정 키: `snake_case`
- Go 변수: `CamelCase`

## 동시성 패턴
### 병렬 실행
```go
var wg sync.WaitGroup
for _, step := range parallel {
    wg.Add(1)
    go func(s Step) {
        defer wg.Done()
        execute(s)
    }(step)
}
wg.Wait()
```

### 채널 사용
```go
results := make(chan Result, len(steps))
errors := make(chan error, 1)
```

## 보안 패턴
### 명령 주입 방지
```go
// exec.Command 사용 (shell 거치지 않음)
cmd := exec.Command("tar", "cf", "-", path)
```

### 민감 정보 처리
```go
// 로그에서 마스킹
masked := regexp.MustCompile(`password=\S+`).
    ReplaceAllString(cmd, "password=***")
```
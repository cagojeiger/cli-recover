# 재사용 가능한 패턴

## 핵심 패턴

### Provider 패턴
```go
type Provider interface {
    Name() string
    Execute(ctx context.Context, opts Options) error
    EstimateSize(opts Options) (int64, error)
    StreamProgress() <-chan Progress
}
```
- 백업/복원 provider 확장성
- 인터페이스 기반 설계
- 진행률 스트리밍 지원

### Factory 패턴 (Registry 대체)
```go
func CreateBackupProvider(name string) (Provider, error) {
    switch name {
    case "filesystem":
        return filesystem.NewProvider(executor), nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}
```
- 단순한 switch 문 사용
- 복잡한 Registry 제거
- 타입 안전성 확보

### 원자적 파일 쓰기
```go
tempFile := outputFile + ".tmp"
defer func() {
    if !success {
        os.Remove(tempFile)
    }
}()
// 쓰기 완료 후
os.Rename(tempFile, outputFile)
```
- 백업 무결성 보장
- OS 레벨 원자성 활용
- 실패 시 자동 정리

### 진행률 모니터링
```go
type ProgressWriter struct {
    writer     io.Writer
    current    int64
    total      int64
    reporter   ProgressReporter
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
    n, err = pw.writer.Write(p)
    pw.current += int64(n)
    pw.reporter.Update(pw.current, pw.total)
    return
}
```
- io.Writer 인터페이스 활용
- 스트리밍 중 실시간 추적
- 다중 환경 지원

### TUI CLI 래퍼
```go
func (b *BackupFlow) executeBackup() {
    args := []string{"backup", "filesystem", b.pod, b.path}
    cmd := exec.Command("cli-recover", args...)
    
    stdout, _ := cmd.StdoutPipe()
    go b.streamOutput(stdout)
    
    cmd.Run()
}
```
- TUI는 CLI 명령 실행
- 비즈니스 로직 중복 제거
- 일관된 동작 보장

## CLI 패턴

### 플래그 레지스트리
```go
var Registry = struct {
    Namespace   string
    Output      string
    Force       string
    // ...
}{
    Namespace: "n",
    Output:    "o", 
    Force:     "f",
}
```
- 중앙 집중식 관리
- 컴파일 타임 충돌 검증
- 타입 안전성

### 하이브리드 인자 처리
```go
func buildOptions(cmd *cobra.Command, args []string) Options {
    opts := Options{}
    // Positional args
    if len(args) >= 1 {
        opts.Pod = args[0]
    }
    // Flag overrides
    if pod, _ := cmd.Flags().GetString("pod"); pod != "" {
        opts.Pod = pod
    }
    return opts
}
```
- Positional args 우선 파싱
- Flag로 오버라이드 가능
- kubectl 스타일 유연성

### 구조화된 에러 처리
```go
type CLIError struct {
    What   string // 무엇이 잘못됐나
    Why    string // 왜 발생했나
    How    string // 어떻게 해결하나
    SeeDoc string // 추가 문서
}
```
- 사용자 친화적 메시지
- 해결책 제시
- 셀프 서비스 유도

# 진행률 보고 구현 가이드

## 기본 구현 패턴

### 1. ProgressReporter 구조체

```go
package progress

import (
    "fmt"
    "io"
    "os"
    "strings"
    "time"
    
    "golang.org/x/term"
    "github.com/cagojeiger/cli-recover/internal/domain/logger"
)

type Reporter struct {
    writer      io.Writer      // 터미널 출력용 (보통 os.Stderr)
    logger      logger.Logger  // 구조화된 로그용
    isTerminal  bool          // 터미널 환경 여부
    lastLog     time.Time     // 마지막 로그 시간
    logInterval time.Duration  // 로그 간격 (기본 10초)
    progressCh  chan<- Progress // TUI용 채널 (옵션)
}

type Progress struct {
    Current   int64
    Total     int64
    Message   string
    Speed     float64 // bytes per second
    StartTime time.Time
}

func NewReporter(w io.Writer, l logger.Logger) *Reporter {
    return &Reporter{
        writer:      w,
        logger:      l,
        isTerminal:  term.IsTerminal(int(os.Stderr.Fd())),
        logInterval: 10 * time.Second,
    }
}
```

### 2. Update 메서드 구현

```go
func (r *Reporter) Update(current, total int64, message string) {
    // 1. 진행률 계산
    var percent int
    if total > 0 {
        percent = int(current * 100 / total)
    }
    
    // 2. 터미널 출력 (항상)
    if r.isTerminal {
        r.writeTerminal(current, total, percent, message)
    }
    
    // 3. 로그 출력 (주기적으로)
    now := time.Now()
    if now.Sub(r.lastLog) >= r.logInterval {
        r.writeLog(current, total, percent, message)
        r.lastLog = now
    }
    
    // 4. TUI 채널 전송 (있다면)
    if r.progressCh != nil {
        select {
        case r.progressCh <- Progress{
            Current: current,
            Total:   total,
            Message: message,
        }:
        default: // non-blocking
        }
    }
}

func (r *Reporter) writeTerminal(current, total int64, percent int, message string) {
    // 터미널 너비 가져오기
    width := 80 // 기본값
    if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil {
        width = w
    }
    
    // 진행 바 생성
    barWidth := 20
    filled := barWidth * percent / 100
    bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
    
    // 포맷팅
    var output string
    if total > 0 {
        output = fmt.Sprintf("\r%s [%s] %d%% (%s/%s)",
            message, bar, percent,
            formatBytes(current), formatBytes(total))
    } else {
        output = fmt.Sprintf("\r%s %s processed",
            message, formatBytes(current))
    }
    
    // 라인 클리어 (이전 내용 지우기)
    fmt.Fprint(r.writer, "\r" + strings.Repeat(" ", width))
    fmt.Fprint(r.writer, output)
}

func (r *Reporter) writeLog(current, total int64, percent int, message string) {
    fields := []logger.Field{
        logger.F("message", message),
        logger.F("bytes_current", current),
        logger.F("bytes_total", total),
    }
    
    if total > 0 {
        fields = append(fields, logger.F("percent", percent))
    }
    
    r.logger.Info("Progress update", fields...)
}
```

### 3. io.TeeReader 패턴 활용

```go
// 다운로드 예시
func downloadWithProgress(url string, dest string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    // 진행률 리포터 생성
    reporter := NewReporter(os.Stderr, logger.Get())
    
    // 진행률 추적용 Writer
    pw := &progressWriter{
        reporter: reporter,
        total:    resp.ContentLength,
        message:  "Downloading",
    }
    
    // TeeReader로 읽으면서 진행률 추적
    reader := io.TeeReader(resp.Body, pw)
    
    // 파일에 저장
    file, err := os.Create(dest + ".tmp")
    if err != nil {
        return err
    }
    defer file.Close()
    
    if _, err := io.Copy(file, reader); err != nil {
        return err
    }
    
    // 원자적 이동
    return os.Rename(dest+".tmp", dest)
}

// progressWriter는 io.Writer 인터페이스 구현
type progressWriter struct {
    reporter *Reporter
    written  int64
    total    int64
    message  string
}

func (pw *progressWriter) Write(p []byte) (int, error) {
    n := len(p)
    pw.written += int64(n)
    pw.reporter.Update(pw.written, pw.total, pw.message)
    return n, nil
}
```

### 4. 백업 작업에 적용

```go
// filesystem.go 수정 예시
func (p *Provider) Execute(ctx context.Context, opts backup.Options) error {
    // ... 기존 코드 ...
    
    // 진행률 리포터 생성
    reporter := progress.NewReporter(os.Stderr, p.logger)
    
    // 크기 추정
    estimatedSize, _ := p.EstimateSize(opts)
    
    // 진행률 추적 Writer
    pw := &progressWriter{
        reporter: reporter,
        total:    estimatedSize,
        message:  "Creating backup",
    }
    
    // 임시 파일에 쓰기
    tempFile := opts.OutputFile + ".tmp"
    outputFile, err := os.Create(tempFile)
    if err != nil {
        return err
    }
    defer outputFile.Close()
    
    // TeeReader로 진행률 추적하며 복사
    reader := io.TeeReader(stdout, pw)
    if _, err := io.Copy(outputFile, reader); err != nil {
        os.Remove(tempFile)
        return err
    }
    
    // 완료 메시지
    reporter.Complete("Backup created successfully")
    
    // 원자적 이동
    return os.Rename(tempFile, opts.OutputFile)
}
```

### 5. 환경별 처리

```go
// CI/CD 환경 감지
func isCI() bool {
    // 일반적인 CI 환경 변수 체크
    ciVars := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "GITLAB_CI"}
    for _, v := range ciVars {
        if os.Getenv(v) != "" {
            return true
        }
    }
    return false
}

// 진행률 모드 결정
func determineProgressMode() string {
    if !term.IsTerminal(int(os.Stderr.Fd())) {
        return "log"  // 파이프나 리다이렉션
    }
    if isCI() {
        return "ci"   // CI 환경
    }
    return "interactive" // 일반 터미널
}
```

### 6. 유틸리티 함수들

```go
// 바이트를 사람이 읽기 쉬운 형식으로
func formatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 남은 시간 계산
func calculateETA(current, total int64, startTime time.Time) time.Duration {
    if current == 0 || total == 0 {
        return 0
    }
    
    elapsed := time.Since(startTime)
    rate := float64(current) / elapsed.Seconds()
    remaining := float64(total-current) / rate
    
    return time.Duration(remaining) * time.Second
}

// 전송 속도 계산
func calculateSpeed(bytes int64, duration time.Duration) string {
    if duration == 0 {
        return "0 B/s"
    }
    
    bytesPerSec := float64(bytes) / duration.Seconds()
    return formatBytes(int64(bytesPerSec)) + "/s"
}
```

### 7. 완료 및 에러 처리

```go
func (r *Reporter) Complete(message string) {
    if r.isTerminal {
        // 진행 바 지우고 완료 메시지
        fmt.Fprintf(r.writer, "\r%s\n", strings.Repeat(" ", 80))
        fmt.Fprintf(r.writer, "✓ %s\n", message)
    }
    
    r.logger.Info("Operation completed", logger.F("message", message))
}

func (r *Reporter) Error(err error, message string) {
    if r.isTerminal {
        // 진행 바 지우고 에러 메시지
        fmt.Fprintf(r.writer, "\r%s\n", strings.Repeat(" ", 80))
        fmt.Fprintf(r.writer, "✗ %s: %v\n", message, err)
    }
    
    r.logger.Error("Operation failed", 
        logger.F("message", message),
        logger.F("error", err.Error()))
}
```

## 테스트 작성

```go
func TestProgressReporter(t *testing.T) {
    // Mock writer와 logger
    var buf bytes.Buffer
    mockLogger := &testLogger{}
    
    reporter := NewReporter(&buf, mockLogger)
    reporter.isTerminal = false // 테스트에서는 비터미널로
    
    // 진행률 업데이트
    reporter.Update(50, 100, "Testing")
    
    // 로그 확인
    assert.Contains(t, mockLogger.LastMessage, "Progress update")
    assert.Equal(t, int64(50), mockLogger.LastFields["bytes_current"])
}
```

## 모범 사례

### DO ✅
- 3초 이상 작업에는 항상 진행률 표시
- 환경에 맞는 출력 방식 자동 선택
- 정확한 크기 정보 제공 (가능한 경우)
- 에러 발생 시 진행률 정리

### DON'T ❌
- 너무 빈번한 업데이트 (100ms 미만)
- 터미널 감지 없이 \r 사용
- 진행률 때문에 성능 저하
- 부정확한 예측 시간 표시

## 통합 체크리스트

- [ ] ProgressReporter 인터페이스 구현
- [ ] 환경 감지 로직 추가
- [ ] io.TeeReader 패턴 적용
- [ ] 터미널/로그 동시 출력
- [ ] 단위 테스트 작성
- [ ] 에러 처리 및 정리
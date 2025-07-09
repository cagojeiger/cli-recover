# Phase 3-1: Restore 명령어 긴급 수정

## 개요
- **목적**: restore 명령어가 멈춰있는 문제 해결 및 문서 스펙 준수
- **우선순위**: P0 (긴급)
- **복잡도**: 35/100
- **예상 시간**: 4-6시간
- **날짜**: 2025-01-09

## 문제 분석

### 1. 핵심 문제: 바이너리 스트리밍 실패
```go
// 현재 문제 코드 (restore.go:131)
outputCh, errorCh := p.executor.Stream(ctx, tarCmd)
// Stream()은 텍스트 기반이라 tar 바이너리 데이터 손상
```

### 2. 명령어 구성 문제
```go
// 현재: sh -c "cat backup.tar | kubectl exec ..."
// 문제: 셸을 통한 파이프로 stdin 연결 실패
```

### 3. 진행률 표시 문제
- tar -x는 기본적으로 출력이 없음
- verbose 옵션(-v) 누락
- stderr 모니터링 실패

### 4. 3초 규칙 위반
- DelayedReporter로 인해 초기 3초간 무반응
- 사용자는 프로그램이 멈춘 것으로 오해

## 해결 방안

### 1. 새로운 RestoreExecutor 구현
```go
// internal/infrastructure/kubernetes/restore_executor.go
type RestoreExecutor struct {
    executor CommandExecutor
}

func (r *RestoreExecutor) ExecuteRestore(ctx context.Context, backupFile string, kubectlArgs []string) error {
    // 1. 백업 파일 열기
    file, err := os.Open(backupFile)
    if err != nil {
        return fmt.Errorf("failed to open backup file: %w", err)
    }
    defer file.Close()
    
    // 2. kubectl exec 명령어 직접 실행
    cmd := exec.CommandContext(ctx, "kubectl", kubectlArgs...)
    cmd.Stdin = file  // 파일을 직접 stdin으로 연결
    
    // 3. stderr 캡처 (진행률용)
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return err
    }
    
    // 4. 실행
    if err := cmd.Start(); err != nil {
        return err
    }
    
    // 5. stderr 모니터링 (별도 고루틴)
    go r.monitorProgress(stderr)
    
    return cmd.Wait()
}
```

### 2. restore.go 수정
```go
// buildTarCommand 수정 - verbose 옵션 추가
kubectlArgs = append(kubectlArgs, "-xvf")  // -v 추가

// Execute 메서드 개선
func (p *RestoreProvider) Execute(ctx context.Context, opts restore.Options) (*restore.RestoreResult, error) {
    // 타임아웃 설정
    ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
    defer cancel()
    
    // ... 검증 로직 ...
    
    // 새로운 executor 사용
    executor := &RestoreExecutor{executor: p.executor}
    err := executor.ExecuteRestore(ctx, opts.BackupFile, kubectlArgs)
    
    // ...
}
```

### 3. 진행률 표시 개선
```go
func (r *RestoreExecutor) monitorProgress(stderr io.Reader) {
    scanner := bufio.NewScanner(stderr)
    fileCount := 0
    
    for scanner.Scan() {
        line := scanner.Text()
        // tar -v 출력 파싱: "x path/to/file"
        if strings.HasPrefix(line, "x ") {
            fileCount++
            fileName := strings.TrimPrefix(line, "x ")
            
            r.progressCh <- restore.Progress{
                Current: int64(fileCount),
                Message: fmt.Sprintf("Restoring: %s", fileName),
            }
        }
    }
}
```

### 4. 즉각적 피드백
```go
// restore_logic.go의 monitorRestoreProgress 수정
func monitorRestoreProgress(provider restore.Provider, estimatedSize int64, done <-chan bool, verbose bool) {
    // DelayedReporter 대신 즉시 표시
    reporter := progress.NewAutoReporterWithDelay(os.Stderr, false)  // false = no delay
    
    // 시작 메시지 즉시 표시
    reporter.Start("Restore", estimatedSize)
    
    // ...
}
```

## 테스트 계획

### 1. 바이너리 파일 복원 테스트
```go
func TestRestoreBinaryFile(t *testing.T) {
    // 바이너리 tar 파일 생성
    // 복원 실행
    // 파일 무결성 검증
}
```

### 2. 진행률 표시 테스트
```go
func TestRestoreProgress(t *testing.T) {
    // Mock stderr 출력
    // 진행률 업데이트 검증
}
```

### 3. 타임아웃 테스트
```go
func TestRestoreTimeout(t *testing.T) {
    // 느린 복원 시뮬레이션
    // 타임아웃 발생 검증
}
```

## 위험 요소 및 대응

### 1. 기존 동작 호환성
- 위험: 기존 스크립트가 깨질 수 있음
- 대응: 출력 형식은 유지, 내부 구현만 변경

### 2. 대용량 파일 처리
- 위험: 메모리 사용량 증가
- 대응: 스트리밍 처리 유지, 버퍼 크기 최적화

### 3. 네트워크 불안정
- 위험: 중간에 연결 끊김
- 대응: 재시도 로직, 체크포인트 기능 (향후)

## 검증 기준

### 성공 기준
1. restore 명령어가 즉시 반응 (3초 이내)
2. 진행률이 실시간으로 표시됨
3. 바이너리 파일이 손상 없이 복원됨
4. 10분 이상 걸리는 작업에서 타임아웃 발생

### 성능 기준
- CPU 사용률: 기존 대비 +5% 이내
- 메모리 사용량: 파일 크기와 무관하게 일정
- 복원 속도: 기존과 동일 또는 개선

## 구현 순서

1. RestoreExecutor 구현 (새 파일)
2. restore.go의 Execute 메서드 수정
3. 진행률 모니터링 개선
4. 타임아웃 및 에러 처리 추가
5. 테스트 작성 및 실행
6. 문서 업데이트

## 롤백 계획

만약 문제 발생 시:
1. git revert로 즉시 롤백
2. 기존 Stream() 방식으로 임시 복구
3. 문제 분석 후 재시도

## 관련 문서
- /docs/backup-integrity/00-overview.md
- /docs/progress-reporting/00-overview.md
- /docs/cli-design-guide/04-implementation-guide.md
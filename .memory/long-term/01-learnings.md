# Bubble Tea 프레임워크 학습 사항

## 주요 제약사항

### 1. Goroutine 사용 제한
**문제**: 별도 goroutine에서 program.Send() 호출 시 panic
**해결**: tea.Cmd 사용
```go
// 잘못된 방식
go func() {
    result := doWork()
    program.Send(result) // 위험!
}()

// 올바른 방식
func doWorkCmd() tea.Cmd {
    return func() tea.Msg {
        return doWork()
    }
}
```

### 2. 상태 불변성
**원칙**: Update는 새 Model을 반환해야 함
```go
// 잘못된 방식
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.counter++ // 포인터가 아니면 작동 안함
    return m, nil
}

// 올바른 방식
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.counter++ // 값 복사본 수정
    return m, nil // 새 복사본 반환
}
```

### 3. WindowSizeMsg 처리
**중요**: 터미널 크기 변경 시 자동으로 발생
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    // 레이아웃 재계산
```

## 메시지 패턴

### 1. 타입 안전 메시지
```go
// 명확한 타입 정의
type BackupStartedMsg struct {
    JobID string
}

type BackupProgressMsg struct {
    JobID    string
    Progress int
}

type BackupCompletedMsg struct {
    JobID   string
    Success bool
    Error   error
}
```

### 2. tea.Cmd 배치
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // 여러 명령 수집
    cmds = append(cmds, checkStatusCmd())
    cmds = append(cmds, updateUICmd())
    
    // Batch로 실행
    return m, tea.Batch(cmds...)
}
```

### 3. 컴포넌트 포커스 관리
```go
type Component interface {
    Focus() tea.Cmd
    Blur() tea.Cmd
    Focused() bool
}
```

## 성능 최적화

### 1. 렌더링 최적화
**문제**: 매 Update마다 전체 View 호출
**해결**:
- 상태 변경 최소화
- 복잡한 계산 캐싱
- strings.Builder 사용

### 2. 메시지 배치 처리
```go
// 여러 업데이트를 하나로
type BatchUpdateMsg struct {
    Updates []interface{}
}
```

## 디버깅 팁

### 1. 로그 파일 사용
```go
func debugLog(format string, args ...interface{}) {
    f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer f.Close()
    fmt.Fprintf(f, format+"\n", args...)
}
```

### 2. 패닉 복구
```go
defer func() {
    if r := recover(); r != nil {
        debugLog("PANIC: %v", r)
    }
}()
```

## teatest 활용법

### 1. 기본 테스트
```go
func TestModel(t *testing.T) {
    m := NewModel()
    tm := teatest.NewTestModel(t, m)
    
    // 키 입력 시뮬레이션
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    
    // 결과 검증
    fm := tm.FinalModel().(Model)
    assert.Equal(t, ScreenNext, fm.screen)
}
```

### 2. 시나리오 테스트
```go
func TestBackupFlow(t *testing.T) {
    tm := teatest.NewTestModel(t, NewModel())
    
    // 1. 백업 타입 선택
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    
    // 2. 네임스페이스 선택
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    
    // 최종 상태 검증
    fm := tm.FinalModel().(Model)
    assert.NotNil(t, fm.activeJob)
}
```

## 흔한 실수

### 1. program.Send() 오용
- Update 내부에서 Send 호출 금지
- 대신 tea.Cmd 반환

### 2. 무한 루프
- Update에서 항상 Cmd 반환 시 주의
- 조건부로 Cmd 반환

### 3. View에서 상태 변경
- View는 읽기 전용
- 상태 변경은 Model과 Update에서만

## 2025-01-07 추가 학습사항

### God Object 안티패턴
**문제**: TUI Model이 115+ 필드로 비대해짐
**증상**:
- 단일 책임 원칙 위반
- 테스트 어려움
- 유지보수 곤란
**교훈**:
- 상태를 도메인별로 분리
- 컴포지션 사용
- 관심사 분리 철저히

### 메모리 누수 이슈
**문제**: 무제한 출력 버퍼로 OOM 발생
**원인**:
- []string으로 모든 출력 저장
- 장시간 실행 시 메모리 증가
**해결**:
- Ring Buffer 도입
- 최근 N개 라인만 유지
**교훈**:
- 리소스 제한 항상 고려
- 무제한 성장 자료구조 주의

### 테스트 가능한 구조
**문제**: TUI 통합 테스트 실패
**원인**:
- 외부 의존성 직접 사용
- Mock 불가능한 구조
**해결**:
- Provider 패턴 도입
- 인터페이스 기반 설계
- 의존성 주입
**효과**:
- 단위 테스트 가능
- 통합 테스트 격리

### 프로세스 관리
**문제**: 좀비 프로세스 생성
**원인**:
- exec.Command 정리 누락
- Context 취소 처리 없음
**해결**:
```go
defer func() {
    if cmd.Process != nil {
        cmd.Process.Kill()
    }
}()
```
**교훈**:
- 리소스 정리 항상 defer
- Context timeout 설정

### CLI vs TUI 복잡도
**발견**: CLI가 TUI보다 테스트하기 쉬움
**이유**:
- 입출력 명확
- 상태 관리 단순
- Mock 용이
**결론**:
- 핵심 기능은 CLI로
- TUI는 얇은 래퍼로

### 코드 중복 제거 (TDD)
**상황**: backup_filesystem.go와 backup_new.go 중복
**접근**:
1. 호환성 테스트 먼저 작성
2. 기능 동일성 검증
3. 안전한 제거
**효과**:
- 유지보수 부담 감소
- 코드베이스 단순화
**교훈**:
- 항상 테스트 먼저
- 점진적 리팩토링
- 레거시 제거 시 신중히

### 로그 시스템 구현
**날짜**: 2025-01-07
**설계 결정**:
- 인터페이스 기반 설계 (도메인 레이어)
- 다중 출력 지원 (Console, File, Both)
- 구조화된 로깅 (Field 기반)
- 자동 로그 로테이션

**구현 패턴**:
```go
// 편의 함수로 깔끔한 API
logger.Info("Starting backup", 
    logger.F("pod", podName),
    logger.F("namespace", namespace),
)
```

**모범 사례**:
1. 컨텍스트 로깅: WithField/WithFields
2. 레벨 기반 필터링: 비싼 연산 보호
3. 글로벌 + 의존성 주입 병행
4. 환경 변수 지원

**테스트 전략**:
- 콘솔: os.Stderr 캡처
- 파일: TempDir 사용
- 로테이션: 작은 MaxSize

**교훈**:
- 단순한 인터페이스가 최고
- 테스트 가능한 설계 중요
- 성능 고려 (레벨 체크)
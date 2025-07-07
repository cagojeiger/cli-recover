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
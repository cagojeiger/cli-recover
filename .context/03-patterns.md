# Patterns

## Testing
- **Golden Files**: testdata/kubectl/*.golden
- **Mock Runner**: 명령 실행 모킹
- **Table-Driven**: 테스트 케이스 배열

## Architecture
- **Interface Design**: Runner, Executor 인터페이스
- **Command Builder**: 타입 안전 명령 구성
- **Elm Architecture**: Model/Update/View (Bubble Tea)

## Bubble Tea Patterns

### Core Rules
- **Never use goroutines**: Bubble Tea가 동시성 관리
- **Use tea.Cmd for I/O**: 모든 I/O는 비동기로
- **Keep Update/View fast**: 블로킹 작업 금지
- **Message-based communication**: 상태 변경은 메시지로만

### Correct Patterns
```go
// ✅ 올바른 비동기 작업
func doAsyncWork() tea.Cmd {
    return func() tea.Msg {
        result := longRunningOperation()
        return resultMsg{result}
    }
}

// ✅ exec.Command 모니터링
func executeCommand(args []string) tea.Cmd {
    return func() tea.Msg {
        cmd := exec.Command(args[0], args[1:]...)
        stdout, _ := cmd.StdoutPipe()
        
        cmd.Start()
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            // Program.Send()로 진행상황 전송
            program.Send(outputMsg{scanner.Text()})
        }
        
        err := cmd.Wait()
        return doneMsg{err}
    }
}
```

### Anti-patterns
```go
// ❌ 절대 하면 안됨
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    go doSomething() // 금지!
    time.Sleep(1 * time.Second) // 금지!
    return m, nil
}

// ❌ Init에서 고루틴 생성
func (m Model) Init() tea.Cmd {
    go m.startBackgroundTask() // 금지!
    return nil
}
```

### Command Patterns
- **tea.Batch**: 여러 명령 동시 실행
- **tea.Sequence**: 순차 실행
- **tea.Tick/Every**: 타이머 작업
- **Program.Send()**: 외부에서 메시지 전송

### Best Practices
- 명령은 함수를 반환하는 함수로 구현
- 채널 사용 시 tea.Cmd 내부에서만
- 긴 작업은 별도 메시지로 진행상황 전송
- Context는 tea.Cmd에 전달하여 취소 처리

## Code Organization
- **Handler Separation**: action, navigation, helpers
- **Screen States**: iota 상수로 화면 관리
- **Internal Packages**: API 경계 명확화

## Error Handling
- **Error Wrapping**: fmt.Errorf("context: %w", err)
- **User Messages**: 기술 에러 → 친화적 메시지

## Standards
- 함수 < 50줄
- 파일 < 500줄
- 인터페이스: -er 접미사
- 테스트: Test- 접두사
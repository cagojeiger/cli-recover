# Patterns

## Testing
- **Golden Files**: testdata/kubectl/*.golden
- **Mock Runner**: 명령 실행 모킹
- **Table-Driven**: 테스트 케이스 배열

## Architecture
- **Interface Design**: Runner, Executor 인터페이스
- **Command Builder**: 타입 안전 명령 구성
- **Elm Architecture**: Model/Update/View (Bubble Tea)

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
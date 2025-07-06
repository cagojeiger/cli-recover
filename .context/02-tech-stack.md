# Tech Stack

## Core
- **Go 1.21+**: 단일 바이너리, 크로스 컴파일
- **Cobra**: CLI 프레임워크
- **Bubble Tea**: TUI 프레임워크 (Elm 아키텍처)
- **Lipgloss**: 터미널 스타일링

## Dependencies
- kubectl (필수)
- Bubbles (textinput 컴포넌트)
- Go standard library

## Intentionally Avoided
- client-go (무거움, kubectl로 충분)
- 복잡한 UI 라이브러리
- ORM/Database
- 설정 파일 라이브러리

## Testing
- Go testing package
- Golden files
- Mock interfaces

## Platform Support
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (미지원)
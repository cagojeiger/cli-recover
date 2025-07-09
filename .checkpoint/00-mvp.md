# MVP Checkpoint

## Date: 2025-01-06
## Version: 1.0.0-beta

## Working Features
- TUI 네비게이션 (4-stage layout)
- 네임스페이스/Pod 선택
- 파일시스템 브라우저
- 백업 옵션 설정
- kubectl 명령어 생성

## Critical Issues
- **TUI 멈춤**: StreamingExecutor 블로킹
- **백업 미실행**: 명령어 생성만 가능
- **커버리지**: 44.3% (목표 90%)

## Next Goals (v1.0.0)
1. TUI 비동기 실행 구현
2. 실제 백업 실행
3. 진행률 표시
4. 테스트 커버리지 90%

## Tech Stack
- Go 1.21+
- Bubble Tea TUI
- kubectl (필수)
- No K8s client-go

## Platform Support
- ✅ macOS (arm64, amd64)
- ✅ Linux (arm64, amd64)
- ⚠️ Windows (미테스트)

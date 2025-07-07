# Legacy TUI Backup

## Backup Date
2025-01-07

## Reason
CLI-First 전략으로 전환하면서 기존 TUI 코드를 백업

## Contents
- `cmd/`: 기존 CLI 명령 (backup filesystem 포함)
- `internal/`: TUI 및 Kubernetes 통합 코드
- `go.mod`, `go.sum`: 의존성 정보
- `README.md`: 프로젝트 문서

## Working Features
- filesystem backup CLI 완전 동작
- TUI 4단계 화면 플로우
- kubectl exec 통합
- 진행률 모니터링

## Known Issues
- God Object 안티패턴 (Model 115+ 필드)
- 메모리 누수 (무제한 출력 버퍼)
- 테스트 불가능한 구조

## Recovery
필요시 이 디렉토리의 내용을 프로젝트 루트로 복사하여 복원 가능
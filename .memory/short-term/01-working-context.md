# Working Context

## Environment
- Go 1.24.3
- Branch: feature/tui-backup
- Coverage: 44.3% (cmd: 24.7%)

## Recent Changes
- MongoDB/MinIO 제거 완료
- 테스트 커버리지 개선 (6.8% → 24.7%)
- TUI 비동기 실행 필요 확인

## Key Files
- internal/tui/executor.go (블로킹 이슈)
- internal/tui/model.go (상태 관리)
- cmd/cli-recover/*_test.go (테스트)

## Technical Debt
- TUI 패키지 테스트 제외
- StreamingExecutor 블로킹
- 일부 함수 50줄 초과
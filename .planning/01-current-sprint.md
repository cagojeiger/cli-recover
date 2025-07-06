# Current Sprint

## TUI Async Execution
- Start: 2025-01-06
- Status: 🔄 진행중
- 복잡도: 30/100

## Backlog

### P0 - Critical
- BackupProgressMsg/CompleteMsg 타입 정의
- executeBackupCmd() → tea.Cmd 구현
- model.Update() 비동기 처리

### P1 - High
- 실시간 출력 스트리밍
- 진행률 표시
- Ctrl+C 취소 지원

### P2 - Medium
- kubectl stderr 캡처
- 에러 파싱/메시지

## Definition of Done
- [ ] UI 반응성 유지
- [ ] 실시간 진행률
- [ ] 깔끔한 취소
- [ ] 에러 처리

## Progress
- ✅ StreamingExecutor 블로킹 이슈 확인
- ✅ 아키텍처 결정 (tea.Cmd)
- 📋 메시지 타입 설계
- 📋 executor 리팩토링
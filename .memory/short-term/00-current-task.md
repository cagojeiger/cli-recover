# Current Task

## TUI 비동기 실행
- 상태: 계획 중
- 우선순위: P0
- 복잡도: 30/100

## 문제
- StreamingExecutor가 UI 블로킹
- 백업 중 입력 불가
- 취소 불가능

## 해결책
- tea.Cmd 패턴 사용
- BackupProgressMsg/CompleteMsg 타입
- executeBackupCmd() 구현

## 완료 기준
- [ ] UI 반응성 유지
- [ ] 실시간 진행률
- [ ] Ctrl+C 취소
- [ ] 에러 피드백
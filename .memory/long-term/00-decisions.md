# Key Decisions

## Active Decisions

### MongoDB/MinIO 제거 (2025-01-06)
- 파일시스템 백업만 지원
- 복잡도 30% 감소
- Occam's Razor 준수

### TUI 비동기 실행 (2025-01-06)
- tea.Cmd 패턴 선택
- 복잡도: 30/100
- 백그라운드 프로세스는 향후 고려

### kubectl 사용 (유효)
- client-go 대신 kubectl 직접 사용
- 무거운 의존성 회피
- 트레이드오프: kubectl 필수

### Interface 설계 (유효)
- Runner, Executor 인터페이스
- 테스트 가능성 확보
- 확장성 고려
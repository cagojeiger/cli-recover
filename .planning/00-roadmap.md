# Roadmap

## Vision
- K8s 백업 도구를 직관적 TUI로 제공
- kubectl 명령어 교육적 가치

## Completed Phases
- **Phase 1-5** ✅: TUI 기본, 파일브라우저, 백업옵션, 리팩토링, 테스트 개선

## Phase 6: TUI Async 🔄 CURRENT
- tea.Cmd 패턴으로 비동기 실행
- 실시간 진행률 업데이트
- Ctrl+C 취소 지원
- UI 반응성 유지

## Phase 7: Production Features 📋 NEXT
- 실제 백업 실행 (현재는 명령어만)
- 진행률 인디케이터
- 에러 복구/재시도
- 설정 파일 지원
- 다양한 출력 포맷

## Phase 8: Advanced Features 📋 FUTURE
- 백그라운드 프로세스 관리
- 백업 스케줄링
- 클라우드 스토리지 연동
- 암호화 기능
- 웹 인터페이스

## Phase 9: Ecosystem 📋 FUTURE
- K8s Operator
- Helm chart
- 모니터링 연동
- 플러그인 시스템
- 멀티클러스터

## Metrics
- 커버리지: 44.3% → 90% 목표
- 함수 크기: <50줄 (대부분 달성)
- 파일 크기: <500줄 ✅
# 현재 작업 상태

## 진행 중인 작업
- TUI 아키텍처 리팩토링 계획 수립
- CLAUDE.md 기반 문서화 진행

## 완료된 작업
- [x] 백업 실행 버그 수정 (cli-recover 중복 제거)
- [x] UI 중복 제거 (Controls, 타이틀)
- [x] Progress 파싱 단순화
- [x] Job Manager 네비게이션 개선
- [x] 전체 명령어 표시 수정

## 다음 단계
- [ ] Ring Buffer 구현 (TDD)
- [ ] 비즈니스 로직 분리
- [ ] 컴포넌트 추출 시작

## 현재 브랜치
- feature/tui-backup

## 주요 이슈
1. 메모리 누수: BackupJob.Output 무제한 증가
2. God Object: Model struct 115+ 필드
3. 테스트 부재: 30% 커버리지
4. UI 반응성: 전체 화면 다시 그리기

## 중요 결정사항
- Bubble Tea 유지 (마이그레이션 비용 > 이익)
- 점진적 리팩토링 접근
- TDD 우선 적용 (비-UI 로직)
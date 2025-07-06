# [IN PROGRESS] Task: TUI Backup Feature Implementation

## Objective
- TUI 기반 대화형 백업 도구 구현
- `cli-restore tui` 명령어로 네임스페이스/Pod 선택
- 최종적으로 `cli-restore backup` CLI 명령어 구성

## Background
- v0.1.0 성공적으로 릴리즈 완료
- 새로운 요구사항: Pod 파일 백업, MinIO 업로드, MongoDB 백업
- 자동완성 대신 TUI 접근법 선택 (의존성 최소화)

## Progress
- [x] 브랜치 생성 (feature/tui-backup)
- [x] 메모리 문서 업데이트
- [ ] Survey 의존성 추가
- [ ] 프로젝트 구조 확장
- [ ] kubectl 래퍼 함수 구현
- [ ] TUI 서브커맨드 구현
- [ ] 백업 실행 로직 구현

## Target Commands
- `cli-restore tui`: 대화형 모드
- `cli-restore backup <pod> <path>`: 직접 실행 모드

## Current Focus
- Phase 1: Pod 파일 백업 기능만 구현
- Survey 라이브러리 사용 (가장 가벼운 TUI)
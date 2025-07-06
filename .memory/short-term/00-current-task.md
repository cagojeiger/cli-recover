# [IN PROGRESS] Task: TUI Complete Redesign

## Objective
- k9s 스타일의 전문적인 TUI 구현
- 명령어 패턴: `cli-restore [action] [target] [options]`
- TUI에서 명령어 구성부터 실행까지 완전 통합

## Background
- v0.1.0 릴리즈 완료 (기본 버전 명령어)
- Survey 기반 간단한 TUI 구현 완료
- 사용자 피드백: 더 전문적이고 통합된 TUI 필요
- k9s 같은 풀스크린 TUI로 전환 결정

## Architecture Decisions
- Survey → Bubble Tea 전환 (풀스크린 TUI)
- 헤더/메인/프리뷰/풋터 표준 레이아웃
- 리스트 기반 UI (박스 UI 제거)
- 함수형 아키텍처 + 독립적 뷰 시스템

## Design Principles
- 명령어 패턴 일관성 유지
- 확장 가능한 타겟 시스템 (pod, mongodb, postgres, minio...)
- vim 스타일 키보드 네비게이션
- 실시간 명령어 프리뷰 및 TUI 내 실행

## Current Status
- [x] Survey 기반 프로토타입 완료
- [x] TUI UX/UI 전체 설계 완료
- [x] 아키텍처 결정사항 확정
- [ ] Bubble Tea 마이그레이션
- [ ] 레이아웃 시스템 구현
- [ ] 뷰 시스템 구현
- [ ] 명령어 빌더/실행기 구현
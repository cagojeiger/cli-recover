# Project Context

## Purpose
- K8s Pod 파일시스템 백업 TUI 도구
- kubectl exec 복잡성 해결
- 직관적 인터페이스 + 교육적 가치

## Current State
- MVP 완성 (명령어 생성만)
- 테스트 커버리지: 44.3%
- 주요 이슈: TUI 비동기 실행 필요

## Working Features
- 네임스페이스/Pod 선택
- 파일시스템 브라우저
- 백업 옵션 설정
- kubectl 명령어 생성
- 명령어 미리보기

## Missing Features
- 실제 백업 실행
- 진행률 표시
- 백그라운드 실행
- 에러 복구
- 설정 파일

## Core Values
- **단순함**: Occam's Razor 준수
- **UX**: 직관적 키보드 네비게이션
- **교육**: kubectl 명령어 학습
- **품질**: 90%+ 테스트 목표

## Constraints
- Go 1.21+
- kubectl 필수
- K8s 클러스터 접근

## Current Focus
- TUI 비동기 실행 (StreamingExecutor 블로킹 해결)
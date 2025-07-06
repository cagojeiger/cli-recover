# Technology Stack

## Core Language
- Go 1.21+
- 선택 이유: Kubernetes 생태계 표준, 단일 바이너리 배포

## CLI Framework
- Cobra (github.com/spf13/cobra)
- 선택 이유:
  * 41k+ stars, 업계 표준
  * kubectl, docker 등 주요 CLI 도구 사용
  * 서브커맨드 구조 지원

## TUI Framework
- Bubble Tea (github.com/charmbracelet/bubbletea)
- 선택 이유:
  * Elm 아키텍처 기반 (React-like)
  * 풀스크린 TUI 지원
  * 실시간 업데이트 가능
  * k9s 수준의 전문적 UI 구현 가능
- 관련 라이브러리:
  * Bubbles: TUI 컴포넌트 라이브러리
  * Lipgloss: 스타일링 라이브러리
- 이전 시도:
  * Survey: 프롬프트 기반으로 한계 있음

## Future Dependencies
- k8s.io/client-go: Kubernetes API 클라이언트
- k8s.io/cli-runtime: kubectl 유틸리티 함수

## Build & Distribution
- Makefile: 빌드 자동화
- GitHub Actions: CI/CD
- GitHub Releases: 바이너리 배포
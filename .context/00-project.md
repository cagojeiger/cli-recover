# Project: CLI-Restore

## Goal
- Kubernetes Pod 파일/폴더 백업 도구
- tar 분할 압축 후 로컬 복사

## Current Phase
- v0.1.0: 버전 명령어 구현 완료 ✓
- v0.2.0: TUI 기반 Pod 백업 기능 구현 중

## Target Features
- `cli-restore --version`: 버전 확인 ✓
- `cli-restore tui`: 대화형 백업 도구 (구현 중)
- `cli-restore backup <pod> <path>`: Pod 백업 직접 실행 (구현 중)

## Constraints
- 크로스 플랫폼 지원:
  * macOS (darwin/amd64, darwin/arm64)
  * Linux (linux/amd64, linux/arm64)
- Go 1.21+ 사용
- 바이너리 배포 (GitHub Releases)
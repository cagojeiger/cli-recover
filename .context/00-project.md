# Project: CLI-Restore

## Goal
- Kubernetes Pod 파일/폴더 백업 도구
- tar 분할 압축 후 로컬 복사

## Current Phase
- v0.1.0: 버전 명령어만 구현
- 최소 기능으로 시작 (Simple Start)

## Target Features
- `cli-restore --version`: 버전 확인
- `cli-restore backup <pod> <path>`: Pod 백업 (향후)

## Constraints
- 크로스 플랫폼 지원:
  * macOS (darwin/amd64, darwin/arm64)
  * Linux (linux/amd64, linux/arm64)
- Go 1.21+ 사용
- 바이너리 배포 (GitHub Releases)
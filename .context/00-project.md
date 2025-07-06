# Project: CLI-Restore

## Goal
- Kubernetes 환경 통합 백업/복원 도구
- 지원 대상:
  * Pod 파일시스템 (tar 분할 압축)
  * MongoDB (mongodump/mongorestore)
  * MinIO (mc mirror)
  * PostgreSQL (pg_dump/pg_restore)
  * Redis, MySQL 등 확장 가능

## Current Phase
- v0.1.0: 버전 명령어 구현 완료 ✓
- v0.2.0: 전문적인 TUI 시스템 개발 중
  * Survey 기반 프로토타입 완료
  * Bubble Tea 기반 재설계 진행 중

## Target Features
- `cli-restore --version`: 버전 확인 ✓
- `cli-restore [action] [target] [options]`: 통합 명령어 패턴
  * Actions: backup, restore, verify, schedule, history
  * Targets: pod, mongodb, postgres, mysql, redis, minio, s3 등
- TUI 모드: k9s 스타일 풀스크린 인터페이스
- CLI 모드: 직접 명령어 실행

## Constraints
- 크로스 플랫폼 지원:
  * macOS (darwin/amd64, darwin/arm64)
  * Linux (linux/amd64, linux/arm64)
- Go 1.21+ 사용
- 바이너리 배포 (GitHub Releases)
- Bitnami Helm 차트 호환성
- 오프라인 환경 지원 (임베디드 바이너리)

## Key Features
- 용량 기반 백업 전략 자동 선택
- Port forward 자동 관리
- 진행률 실시간 모니터링
- 도구 자동 감지 및 대체 전략
- k9s 스타일 전문적 TUI
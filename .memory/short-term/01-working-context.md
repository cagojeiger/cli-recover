# 현재 작업 컨텍스트

## 전략 전환 ✨
- **TUI 중심 → CLI 우선 개발로 전환** (2025-01-07)
- 이유: 동작하는 CLI가 이미 있고, 더 실용적
- 원칙: "Make it work → Make it right → Make it pretty"

## 현재 동작 중
- filesystem 백업 CLI 완전 동작 ✓
  ```bash
  ./cli-recover backup filesystem <pod> <path> --namespace <ns>
  ```
- kubectl exec + tar 통합 완료
- 진행률 모니터링 구현
- 크기 추정 & ETA 계산

## 완료된 작업 (2025-01-07)
### 전략 및 설계
- CLI-First 전략 결정 ✓
- 전략적 결정사항 문서화 ✓
- CLI 중심 로드맵 재작성 ✓
- 아키텍처 패턴 설계 (Hexagonal + Plugin) ✓

### 구현 완료 (TDD 방식)
- 도메인 타입 정의 (Progress, Options, BackupError) ✓
- BackupProvider 인터페이스 ✓
- Provider 레지스트리 시스템 ✓
- Kubernetes 추상화 계층 ✓
  - KubeClient, CommandExecutor 인터페이스
  - KubectlClient (JSON 파싱)
  - OSCommandExecutor
- Filesystem Provider 완전 구현 ✓
  - 모든 테스트 통과
  - 진행률 스트리밍
  - 압축/exclude 옵션
- CLI 프레임워크 통합 ✓
  - Provider 초기화 시스템
  - CLI 어댑터 레이어
  - 새로운 명령 구조 (`backup <type>`)
  - 통합 테스트 작성

## 진행 중인 작업
- CLI 프레임워크 통합 완료 ✅
- 문서 업데이트 진행 중 🔄
- Git 상태 정리 예정

## 참조 문서
- `.memory/long-term/04-strategic-decisions.md`: CLI-First 전략 결정
- `.planning/00-roadmap.md`: CLI 중심 개발 로드맵
- `.planning/03-architecture-patterns.md`: 아키텍처 설계
- `.memory/long-term/03-architecture-decisions.md`: 아키텍처 근거
- `.planning/05-cli-phase1-progress.md`: 진행 상황 추적 ✨ NEW

## 다음 작업 (수정된 계획)

### 1. Restore 기능 구현 🆕
```bash
cli-recover restore filesystem <pod> <backup-file> [options]
```
- RestoreProvider 인터페이스 설계
- Filesystem restore 구현
- 진행률 추적

### 2. List/Status 명령 🆕
```bash
cli-recover list backups [--namespace <ns>]
cli-recover status <job-id>
```
- 메타데이터 저장 시스템
- 백업 이력 관리

### 3. Provider 확장 (Phase 5로 이동) ⏭️
- MinIO/MongoDB는 TUI 완성 후 구현
- 복잡도 관리 및 효율성을 위한 결정

## 핵심 파일들

### CLI 구조 (현재)
- `cmd/cli-recover/main.go`: 진입점
- `cmd/cli-recover/backup_filesystem.go`: filesystem 백업 구현

### 리팩토링 대상
- 명령 체계를 cobra 등으로 구조화
- Provider 패턴으로 백업 타입 추상화
- 도메인 레이어 분리

### 재사용 가능한 자산
- kubectl 실행 로직
- 진행률 계산 알고리즘
- 파일 크기 추정 로직
- 에러 처리 패턴

## 중요 결정사항
- **CLI 우선**: 모든 기능을 CLI로 먼저 구현
- **TUI는 래퍼**: CLI 위의 얇은 레이어로
- **플러그인 패턴**: BackupProvider로 확장성
- **점진적 개발**: 동작하는 것부터 시작

## 브랜치 상태
- 현재: feature/tui-backup
- CLI 중심 개발로 방향 전환됨
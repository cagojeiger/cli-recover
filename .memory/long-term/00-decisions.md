# Technical Decisions

## 2025-07-06: CLI Framework Selection
- **Decision**: Cobra (github.com/spf13/cobra)
- **Rationale**:
  * 41k+ GitHub stars
  * Industry standard (kubectl, docker, helm)
  * Built-in version flag support
  * Excellent subcommand structure for future
- **Alternatives Considered**:
  * urfave/cli (23k stars) - Simpler but less features
  * kingpin - Deprecated

## 2025-07-06: Project Structure
- **Decision**: Simple start with flat structure
- **Rationale**:
  * Avoid over-engineering
  * main.go contains all code initially
  * Will expand to internal/ when needed
- **Reference**: "Start simple, always with a flat structure"

## 2025-07-06: Cross-Platform Support
- **Decision**: Support macOS and Linux from start
- **Platforms**:
  * darwin/amd64 (Intel Mac)
  * darwin/arm64 (Apple Silicon)
  * linux/amd64
  * linux/arm64
- **Rationale**: Kubernetes runs on Linux servers

## 2025-07-06: Memory Management Strategy
- **Decision**: Short-term → Long-term transition rules
- **Criteria**:
  * Task completion → Move to decisions/learnings
  * Pattern discovery → Document in patterns.md
  * Milestone reached → Create checkpoint
- **Rationale**: Keep active memory clean, preserve important knowledge

## 2025-07-06: TUI Framework Migration
- **Decision**: Survey → Bubble Tea
- **Rationale**:
  * Survey는 프롬프트 방식으로 k9s 같은 풀스크린 TUI 불가
  * Bubble Tea는 Elm 아키텍처로 복잡한 상태 관리 가능
  * 실시간 업데이트, 애니메이션 지원
  * Bubbles 컴포넌트 라이브러리 활용 가능
- **Trade-offs**:
  * 학습 곡선 있음
  * 의존성 증가
  * 하지만 전문적인 TUI 구현 가능

## 2025-07-06: Command Pattern Architecture
- **Decision**: `cli-restore [action] [target] [options]` 패턴 확립
- **Rationale**:
  * 일관된 사용자 경험
  * 확장 가능한 구조
  * CLI와 TUI 통합 용이
- **Examples**:
  * `cli-restore backup pod nginx /data`
  * `cli-restore restore mongodb dump.gz mongo-primary`
  * `cli-restore verify minio backup-bucket`

## 2025-07-06: TUI Layout System
- **Decision**: Header/Main/Preview/Footer 표준 레이아웃
- **Components**:
  * Header: 상태 정보 + 네비게이션 경로
  * Main: 리스트 기반 콘텐츠
  * Preview: 실시간 명령어 생성
  * Footer: 컨텍스트별 단축키
- **Rationale**: 
  * 모든 화면 일관성
  * 명령어 투명성
  * k9s 스타일 전문성

## 2025-07-06: Bitnami Chart Support Strategy
- **Decision**: Chart-aware backup strategies
- **Key Findings**:
  * MongoDB: mongodump included ✓
  * MinIO: mc NOT included ✗
  * PostgreSQL: pg_dump included ✓
- **Implications**:
  * MinIO requires local mc or binary injection
  * Other services can use pod-internal tools
  * Storage capacity check critical

## 2025-07-06: Storage-Aware Backup Strategy
- **Decision**: Automatic strategy selection based on data size
- **Thresholds**:
  * < 10GB: Pod internal if space > 2x data
  * 10-100GB: Always streaming
  * > 100GB: Parallel/incremental required
- **Rationale**: Prevent pod storage exhaustion
- **Default**: External streaming for safety

## 2025-07-06: Offline Environment Support
- **Decision**: Embed essential binaries using Go embed
- **Strategy**:
  * Minimal build: mc only (~35MB)
  * Standard build: mc all platforms (~80MB)
  * Full offline: all tools (~300MB)
- **Security**:
  * SHA256 verification
  * Temporary injection only
  * Guaranteed cleanup
- **Rationale**: Support air-gapped environments
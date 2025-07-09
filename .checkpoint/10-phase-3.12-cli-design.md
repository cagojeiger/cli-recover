# Checkpoint: Phase 3.12 CLI 디자인 문서

## 날짜: 2025-01-09
## 상태: 문서화 완료, 부분 구현 (플래그 충돌 해결)

## 완료된 작업

### CLI 디자인 가이드 작성
- `/docs/cli-design-guide/` 디렉토리 생성
- 6개 문서 작성 완료:
  1. `00-overview.md`: 전체 아키텍처와 시스템 흐름
  2. `01-command-patterns.md`: 명령어 구조와 패턴
  3. `02-flag-management.md`: 플래그 충돌 분석 및 해결책
  4. `03-user-experience.md`: UX 디자인과 피드백 전략
  5. `04-implementation-guide.md`: 구현 템플릿과 패턴
  6. `05-testing-strategy.md`: 테스트 전략과 체크리스트

### 주요 디자인 결정

#### 1. 플래그 충돌 해결 ✅
**문제점 발견**:
- `-o`: backup(--output) vs restore(--overwrite)
- `-c`: backup(--compression) vs restore(--container)
- `-t`: backup(--totals) vs restore(--target-path)

**구현 완료** (2025-01-09):
- backup: `-t` → `-T` (--totals) ✅
- restore: `-o` → `-f` (--force) ✅
- restore: `-c` → `-C` (--container) ✅

#### 2. 플래그 레지스트리 설계
```go
var Registry = struct {
    Namespace   string
    Output      string
    Force       string
    Container   string
    // ...
}{
    Namespace: "n",
    Output:    "o",
    Force:     "f",
    Container: "C",
}
```

#### 3. 하이브리드 인자 처리
- Positional args 우선: `backup filesystem <pod> <path>`
- Flag 오버라이드 가능: `--pod`, `--path`
- kubectl/docker 스타일 준수

#### 4. 에러 처리 철학
- What: 무엇이 잘못됐나
- Why: 왜 발생했나
- How: 어떻게 해결하나
- See: 추가 문서 링크

### 시각적 다이어그램
- Mermaid 다이어그램으로 복잡한 관계 시각화
- 시스템 흐름, 명령 계층, 플래그 관리 등
- 사용자 인터랙션 플로우

## Phase 3.12 구현 현황

### 완료된 작업 ✅
1. **플래그 충돌 해결** (2025-01-09)
   - backup.go: `-t` → `-T`
   - restore.go: `-o` → `-f`, `-c` → `-C`
   - restore_logic.go: GetBool("overwrite") → GetBool("force")

### 미구현 항목 ❌
1. **플래그 레지스트리** (flags/registry.go)
2. **하이브리드 인자** (옵션 빌더 개선)
3. **CLIError 타입** (구조화된 에러)
4. **진행률 통합** (모든 provider)
5. **문서 업데이트** (CHANGELOG, README)

### 예상 복잡도: 30/100
- 단순한 리팩토링 위주
- 기능 추가 최소화
- 테스트 커버리지 유지

## 프로젝트 현황

### 완료된 Phase
- Phase 1-3: CLI 핵심 기능 ✅
- Phase 3.9: 아키텍처 단순화 ✅
- Phase 3.10: 백업 무결성 ✅
- Phase 3.11: 진행률 보고 시스템 ✅
- Phase 3.12 문서: CLI 디자인 가이드 ✅

### 진행 예정
- Phase 3.12 구현: CLI 사용성 개선 (준비 완료)
- Phase 3.13: 도구 자동 다운로드
- Phase 4: TUI 구현

## 코드베이스 상태
- 테스트 커버리지: 50.7%
- 모든 테스트 통과
- 복잡도: ~30/100
- 문서화 수준: 높음

## 참고사항
- CLAUDE.md 원칙 준수
- Occam's Razor 적용
- TDD 방식 개발 예정

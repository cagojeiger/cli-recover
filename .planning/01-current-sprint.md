# 현재 스프린트: 테스트 커버리지 향상 및 플래그 충돌 해결

## 스프린트 정보
- **시작**: 2025-01-09
- **목표**: 테스트 커버리지 90% 달성 및 플래그 충돌 해결
- **상태**: 🚀 진행중
- **우선순위**: CLAUDE.md RULE_04 준수

## Phase 3.* 통합 현황
- ✅ Phase 3.9: 아키텍처 단순화 (완료)
- ✅ Phase 3.10: 백업 무결성 (완료)
- ✅ Phase 3.11: 진행률 보고 (완료)
- 📝 Phase 3.12: CLI 사용성 (문서만, 0%)
- 📋 Phase 3.13: 도구 다운로드 (계획만)

## 주요 작업 항목

### 1. 테스트 커버리지 향상 (우선순위 1) ⏳
- [ ] restore_logic_test.go 작성
- [ ] list_logic_test.go 작성
- [ ] logs_test.go 작성
- [ ] 현재: 50.7% → 목표: 90%

### 2. 플래그 충돌 해결 (우선순위 2) ⏳
- [ ] backup: `-t` → `-T` (--totals)
- [ ] restore: `-o` → `-f` (--force)
- [ ] restore: `-c` → `-C` (--container)
- [ ] 복잡도: ~10/100 (간단한 변경)

### 3. Phase 3.12 최소 구현 (우선순위 3) ⏳
- [ ] 플래그 충돌 해결 구현
- [ ] 복잡한 레지스트리 제외
- [ ] CLIError는 나중에

## 발견된 실제 문제
- **플래그 충돌 미해결**:
  - `-o`: backup(--output) vs restore(--overwrite)
  - `-c`: backup(--compression) vs restore(--container)
  - `-t`: backup(--totals) vs restore(--target-path)
- **낮은 테스트 커버리지**:
  - cmd/cli-recover: 33.8%
  - tui: 0.0%
  - domain/log: 45.0%

## 진행 상황
```
[✓] Phase 3.* 통합 문서 작성 (100%)
[→] 메모리 파일 업데이트 (50%)
[ ] 테스트 커버리지 향상 (0%)
[ ] 플래그 충돌 해결 (0%)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
전체: 25%
```

## 성공 지표
- ⏳ 테스트 커버리지 90% 달성
- ⏳ 플래그 충돌 해결
- ✅ 복잡도 최소화
- ✅ CLAUDE.md 규칙 준수

## 참고 자료
- [Phase 3.* 통합 현황](.checkpoint/12-phase-3-integration.md)
- [CLI 디자인 가이드](../docs/cli-design-guide/)
- [CLAUDE.md](../CLAUDE.md)
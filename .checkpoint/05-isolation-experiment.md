# Phase 4 재설계 - Provider 격리 실험

## 날짜: 2025-01-10
## 상태: 🧪 실험적 접근 준비
## 브랜치: feature/tui-backup
## 철학: "격리성 > 재사용성"

## 이전 상황 요약

### 완료된 작업
- ✅ Phase 1-3: 기본 기능 완성
- ✅ Phase 3.9: 아키텍처 단순화 (3계층 → 2계층)
- ✅ Phase 3.10: 백업 무결성 (원자적 쓰기)
- ✅ Phase 3.11: 진행률 보고 시스템
- ✅ Phase 3.12: CLI 사용성 개선
- ✅ Phase 3-1: Restore 긴급 수정

### 현재 상태
- 테스트 커버리지: 52.9%
- 복잡도: ~30/100
- Provider: Filesystem만 구현
- 모든 기본 기능 동작

## 중요한 전환점

### 발견된 문제
1. **restore 사용성 이슈**
   - "File exists" 에러 시 가이드 부족
   - 사용자 혼란 야기

2. **아키텍처 논의**
   - "CLI → TUI 자동 생성" 아이디어
   - 공통 인터페이스의 한계 인식

3. **핵심 통찰**
   - 사용자: "격리성이 재사용성보다 중요"
   - "코드량이 늘어나더라도 제대로 격리만 된다면 서로 영향을 주지 않으니까"

### 새로운 철학 채택
> "Duplication is cheaper than the wrong abstraction"  
> — Sandi Metz

> "Isolation with minimal coordination"  
> — 우리의 결론

### 리팩토링 트라우마
- 과거: backup.go, backup_old.go, backup_v2.go 지옥
- 해결: experimental/ 디렉토리 전략
- 원칙: 기존 코드 절대 수정 금지

## 현재 결정사항

### 1. experimental/ 접근법
```
cli-recover/                    # 기존 코드 (변경 없음)
experimental/                   # 새로운 실험 공간
└── providers/
    └── filesystem_v2/         # 명확한 버전 표시
        └── README.md          # "EXPERIMENTAL" 경고
```

### 2. Phase 4 재설계
- 복잡도 목표: 35 → 20으로 하향
- 기간: 5일 단계별 실험
- 방식: TDD + 점진적 접근

### 3. 안전장치
- 환경변수로 전환: `USE_EXPERIMENTAL=true`
- 언제든 롤백 가능 (폴더 삭제)
- A/B 테스트 인프라

## 문서 대청소 완료

### 삭제된 레거시 (15개)
- 완료된 Phase 문서들
- 구식 Provider 분석
- 오래된 checkpoint들
- 빈 README 파일들

### 남은 핵심 문서 (18개)
- 격리 철학 문서들
- 리팩토링 전략
- 핵심 교훈들
- 현재 계획

## 다음 행동

### Phase 4-1: Experimental 구조 생성 (1일)
- [ ] experimental/providers/ 디렉토리 생성
- [ ] README.md 경고문 작성
- [ ] .gitignore 설정 추가

### Phase 4-2: 최소 기능 구현 (2일)
- [ ] EstimateSize 격리 구현
- [ ] TDD로 처음부터 작성
- [ ] 100% 테스트 커버리지

### Phase 4-3: A/B 테스트 (1일)
- [ ] 환경변수 전환 구현
- [ ] 성능 비교 측정
- [ ] 사용 추적 로깅

### Phase 4-4: 평가 (1일)
- [ ] 성공/실패 판단
- [ ] 다음 단계 결정

## 성공 기준

### 실험 성공
- 코드 복잡도 감소
- Provider 완전 격리
- 테스트 독립성 향상

### 실험 실패
- experimental/ 삭제
- 다른 접근 모색

### 부분 성공
- 좋은 아이디어만 적용
- 실용적 타협

## 위험 관리
- 기존 코드 영향: 0% (완전 격리)
- 롤백 난이도: 매우 쉬움 (폴더 삭제)
- 팀원 혼란: 최소화 (명확한 경계)

## 핵심 원칙
1. **Clear Boundaries**: 명확한 경계 유지
2. **No Big Bang**: 점진적 접근
3. **Prove First**: 증명 후 통합
4. **Easy Rollback**: 쉬운 롤백

## 참고 문서
- [94-design-philosophy.md](../.planning/94-design-philosophy.md) - 격리 철학
- [07-refactoring-strategy.md](../.context/07-refactoring-strategy.md) - 안전한 리팩토링
- [99-vision-draft.md](../.planning/99-vision-draft.md) - 구현 비전
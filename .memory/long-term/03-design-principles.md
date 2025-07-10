# Design Principles - 설계 원칙 기록

## 핵심 철학

### 격리성 우선 설계
**채택일**: 2025-01-10
**계기**: Phase 3-1 restore hanging 이슈 해결 과정에서 깨달음

> "Duplication is cheaper than the wrong abstraction"  
> — Sandi Metz

> "Isolation with minimal coordination"  
> — 우리의 결론

### 원칙

#### 1. Complete Isolation (완전 격리)
- 각 Provider는 독립적인 모듈
- 다른 Provider 문제가 전파되지 않음
- 독립적 개발, 테스트, 배포 가능

#### 2. Intentional Duplication (의도적 중복)
- 코드 중복을 두려워하지 않음
- 각 Provider에 최적화된 구현 우선
- 잘못된 추상화보다 중복이 나음

#### 3. Minimal Coordination (최소 조율)
- 공유는 정말 필요한 것만
  - Logger (로그 일관성)
  - Base Error (사용자 경험)
- 그 외는 모두 독립

## 과거의 실수

### Wrong Abstraction 사례
```go
// 처음엔 좋아 보였던 공통 인터페이스
type Provider interface {
    Execute(ctx context.Context, opts Options) error
    EstimateSize(opts Options) (int64, error)
    StreamProgress() <-chan Progress
}

// 결과
// 1. Options 구조체 비대화
// 2. if provider == "mongodb" 분기 난무
// 3. 억지로 인터페이스 맞추기
```

## 현재의 선택

### Provider 격리 구조
```
providers/
├── filesystem/      # 100% 독립
│   ├── backup.go
│   ├── restore.go
│   ├── tui.go      # 전용 TUI
│   └── tests/      # 독립 테스트
└── mongodb/         # 100% 독립
    └── (별도 구현)
```

### Trade-offs (의도적 선택)

#### 수용하는 것
1. 코드 중복 (격리의 대가)
2. 바이너리 크기 증가 (무시 가능)
3. 수동 일관성 유지 (자동화보다 실용적)

#### 얻는 것
1. 완벽한 격리 (버그 전파 차단)
2. 단순함 (전체 이해 불필요)
3. 최적화 자유도 (타협 없음)

## 적용 지침

### 언제 공유할 것인가
- 정말로 100% 동일한 로직
- 모든 Provider가 함께 변경되어야 함
- 사용자 경험의 일관성 필요

### 언제 격리할 것인가
- Provider별로 다른 특성
- 미래에 달라질 가능성
- 최적화 방법이 다른 경우

## 영향

### 개발자 경험
- 새 개발자는 하나의 Provider만 이해하면 됨
- Provider 담당자는 전체 시스템 몰라도 됨
- 빠른 온보딩과 기여 가능

### 운영 안정성
- 한 Provider 장애가 격리됨
- 독립적 롤백 가능
- Provider별 다른 배포 주기

## 미래 전망

### 6개월 후
- 각 Provider가 독립적으로 성숙
- 서로 다른 속도로 발전
- 충돌 없는 공존

### 1년 후
- Provider별 전문가 등장
- 각자의 최적화 극대화
- 안정적인 운영 체계

## 결론

```
코드 중복을 두려워하지 마라.
잘못된 추상화를 두려워하라.

격리된 단순함이
통합된 복잡함보다 낫다.
```

이것이 CLI-Recover의 설계 철학이다.

## 참고 문서
- [94-design-philosophy.md](../../.planning/94-design-philosophy.md)
- [99-vision-draft.md](../../.planning/99-vision-draft.md)
- [98-architecture-insights.md](../../.planning/98-architecture-insights.md)
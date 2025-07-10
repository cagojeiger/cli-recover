# Design Philosophy - 격리성 우선 설계

## 핵심 명언

> "Duplication is cheaper than the wrong abstraction"  
> — Sandi Metz

> "Perfect is the enemy of good"  
> — Voltaire

> "격리성 > 재사용성"  
> — 우리의 깨달음

> "Isolation with minimal coordination"  
> — 최종 결론

> "Keep it simple, make it work"  
> — 실용주의 원칙

## 설계 철학의 전환

### 기존 사고방식 (교과서적)
- DRY (Don't Repeat Yourself) 맹신
- 공통 인터페이스로 모든 것 통합
- 재사용 가능한 컴포넌트 우선
- 중복 코드는 악

### 새로운 사고방식 (현실적)
- WET (Write Everything Twice) 허용
- Provider별 완전 격리
- 독립성과 안정성 우선
- 중복은 잘못된 추상화보다 나음

## 격리성 우선 원칙

### 1. Complete Isolation
```
각 Provider는 독립적인 섬이다.
다른 Provider가 침몰해도 영향받지 않는다.
```

### 2. Minimal Coordination
```
공유는 최소한으로:
- Logger (로그 일관성)
- Base Error (사용자 경험)
그 외는 모두 독립
```

### 3. Intentional Duplication
```
중복 코드를 두려워하지 않는다.
각 Provider에 맞는 최적 구현이 더 중요하다.
```

## 실제 적용

### 디렉토리 구조
```
cli-recover/
├── providers/
│   ├── filesystem/      # 100% 독립
│   │   ├── backup.go    # tar 전문
│   │   ├── restore.go   # 복원 전문
│   │   ├── tui.go       # filesystem 특화 TUI
│   │   └── tests/       # 독립 테스트
│   │
│   └── mongodb/         # 100% 독립
│       ├── dump.go      # mongodump 전문
│       ├── restore.go   # mongorestore 전문
│       ├── tui.go       # MongoDB 특화 TUI
│       └── tests/       # 독립 테스트
│
└── internal/shared/     # 정말 최소한만
    ├── logger/
    └── errors/
```

### 코드 예시
```go
// ❌ 나쁜 예: 강제된 인터페이스
type Provider interface {
    Backup(opts Options) error
    Restore(opts Options) error
    EstimateSize() int64
}

// ✅ 좋은 예: Provider별 특화
// filesystem/backup.go
func TarBackup(pod, path string, compression string) error {
    // tar에 최적화된 구현
}

// mongodb/dump.go  
func MongoDump(uri string, includeOplog bool) error {
    // mongodump에 최적화된 구현
}
```

## Trade-offs (의도적 선택)

### 수용하는 것
1. **코드 중복**
   - 각 Provider에 비슷한 코드 존재
   - 하지만 미묘하게 다름 (그게 중요)

2. **바이너리 크기 증가**
   - 2023년 기준 무시 가능한 수준
   - 안정성이 10MB보다 중요

3. **수동 일관성 유지**
   - 자동화의 복잡성보다 나음
   - 각 Provider 담당자의 책임

### 얻는 것
1. **완벽한 격리**
   - 한 Provider 버그가 다른 곳 영향 없음
   - 독립적 배포/롤백 가능

2. **단순함**
   - 새 개발자가 하나의 Provider만 이해하면 됨
   - 전체 시스템 이해 불필요

3. **최적화 자유도**
   - 각 Provider에 맞는 최선의 구현
   - 타협 없는 성능

## 실제 사례

### Wrong Abstraction의 예
```go
// 처음엔 좋아 보였던...
type BackupProvider interface {
    Execute(ctx context.Context, opts Options) error
}

// 현실
// - Filesystem은 tar 스트리밍
// - MongoDB는 BSON 덤프 + oplog
// - MinIO는 HTTP multipart
// → Options가 계속 비대해짐
// → if provider == "mongodb" 분기 난무
```

### 격리의 이점
```go
// filesystem 개발자
"MongoDB? 몰라도 됨. 내 tar만 잘 돌면 됨"

// mongodb 개발자  
"Filesystem 터졌다고? 내 거는 멀쩡함"

// 사용자
"cli-recover backup filesystem 여전히 잘 됨"
```

## 언제 공유할 것인가?

### 공유해야 할 때
1. 정말로 100% 동일한 로직
2. 변경 시 모든 Provider가 함께 변경되어야 함
3. 사용자 경험의 일관성 (에러 메시지 등)

### 공유하지 말아야 할 때
1. "비슷해 보이는" 코드
2. 지금은 같지만 미래에 달라질 수 있는 것
3. Provider 특성에 따라 최적화가 다른 것

## 미래 전망

### 6개월 후
- 각 Provider가 독립적으로 진화
- 서로 다른 속도로 발전
- 충돌 없는 평화로운 공존

### 1년 후
- MongoDB Provider는 샤딩 지원
- Filesystem은 증분 백업 추가
- MinIO는 멀티파트 최적화
- **서로 영향 없이 각자 발전**

## 리팩토링 철학

### "Clear Boundaries over Clean Code"
깨끗한 코드보다 명확한 경계가 우선이다.

> "혼란스러운 깨끗한 코드보다
> 명확한 더러운 코드가 낫다"

### 실험적 접근
1. **experimental/ 디렉토리로 격리**
   - 기존 코드와 완전 분리
   - 실패해도 영향 없음
   - 명확한 경고 표시

2. **기존 코드 절대 건드리지 않음**
   - backup.go, backup_old.go 지옥 방지
   - Git history 보존
   - 팀원 혼란 제거

3. **성공 증명 후에만 통합**
   - A/B 테스트로 검증
   - 성능 비교 측정
   - 점진적 마이그레이션

4. **실패는 폴더 삭제로 끝**
   - 깨끗한 롤백
   - 흔적 없는 실패
   - 다시 시작 가능

### 마이그레이션 원칙
- **No Big Bang**: 한 번에 모든 것 금지
- **Feature Toggle**: 기능 전환 가능
- **Parallel Run**: 병렬 실행 기간
- **Clear Rollback**: 명확한 롤백 경로

### 리팩토링 경험의 교훈
```
과거: "이 파일이 최신인가?"
현재: "experimental/은 실험, 나머지는 프로덕션"
미래: "성공한 실험만 통합"
```

## 결론

```
"코드 중복을 두려워하지 마라.
잘못된 추상화를 두려워하라.

격리된 단순함이
통합된 복잡함보다 낫다.

명확한 경계가
깨끗한 코드보다 중요하다."
```

이것이 CLI-Recover 2.0의 철학이다.
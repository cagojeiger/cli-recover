# Architecture Evolution - 아키텍처 진화 과정

## 현재 아키텍처: 2계층 헥사고날 (2025-01-08)

### 구조
```
internal/
├── domain/              # 비즈니스 로직 & 인터페이스 (순수 도메인)
│   ├── backup/         # 백업 인터페이스
│   ├── restore/        # 복원 인터페이스
│   ├── metadata/       # 메타데이터 인터페이스
│   ├── logger/         # 로거 인터페이스
│   ├── log/            # 작업 이력 도메인
│   └── progress/       # 진행률 타입
└── infrastructure/      # 외부 시스템 연동 & 구현체
    ├── config/         # 설정 관리
    ├── filesystem/     # 파일시스템 provider 구현
    ├── kubernetes/     # K8s 클라이언트
    ├── logger/         # 로거 구현체
    ├── progress/       # 진행률 구현체
    └── metadata/       # 메타데이터 파일 저장소
```

### 책임 분리
```
Domain Layer (What & Why)
├── 비즈니스 규칙
├── 도메인 모델
├── 인터페이스 정의
└── 외부 의존성 없음

Infrastructure Layer (How)
├── Domain 인터페이스 구현
├── 외부 시스템 통합
├── 파일/네트워크 I/O
└── 실제 저장소 구현
```

### 의존성 방향
- CMD → Infrastructure → Domain
- Domain은 순수함 (no imports)
- Infrastructure가 Domain에 의존

## 아키텍처 전환 과정

### Phase 1-3: 초기 3계층 구조
```
cmd/
├── adapter/        # Application 레이어
├── domain/         # Domain 레이어
└── infrastructure/ # Infrastructure 레이어
```

### Phase 3.9: 2계층으로 단순화 ✅
**문제점 발견**:
- Application 레이어가 단순 전달만 수행
- Registry 패턴이 provider 1개만 관리
- 과도한 추상화로 복잡도 증가

**해결책**:
- Application 레이어 제거
- Registry → Factory 함수로 교체
- 직접적인 호출 구조

**결과**:
- 복잡도: 75 → 30
- 파일 수: -40%
- 코드 라인: -35%

## 발견된 한계점

### 1. 공통 인터페이스의 억압
```go
// 현재: 모든 Provider가 같은 인터페이스 강제
type Provider interface {
    Execute(ctx context.Context, opts Options) error
    EstimateSize(opts Options) (int64, error)
    StreamProgress() <-chan Progress
}
```
- Filesystem: `du -s`로 간단
- MongoDB: dump 후에나 알 수 있음
- MinIO: 메타데이터로 즉시 계산

### 2. Options 구조체 비대화
```go
type Options struct {
    // 공통 필드
    Namespace   string
    Pod         string
    
    // Backup 전용
    SourcePath  string
    Compression string
    
    // Restore 전용
    BackupFile  string
    Overwrite   bool
    
    // MongoDB 전용 (미래)
    Collections []string
    OpLog      bool
    
    // MinIO 전용 (미래)
    Bucket     string
    Prefix     string
}
```

### 3. Provider별 특성 무시
- Filesystem: tar 기반
- MongoDB: BSON dump
- MinIO: S3 API
- 각각 완전히 다른 패러다임

## 미래 방향: Provider 격리 아키텍처

### 철학
> "Duplication is cheaper than the wrong abstraction"  
> — Sandi Metz

> "Isolation with minimal coordination"  
> — 우리의 결론

### 목표 구조
```
cli-recover/
├── providers/
│   ├── filesystem/      # 100% 독립
│   │   ├── backup.go    # tar 특화 구현
│   │   ├── restore.go   # 복원 특화 구현
│   │   ├── tui.go       # filesystem 전용 TUI
│   │   └── tests/       # 독립 테스트
│   │
│   └── mongodb/         # 100% 독립
│       ├── dump.go      # mongodump 특화
│       ├── restore.go   # mongorestore 특화
│       ├── tui.go       # MongoDB 전용 TUI
│       └── tests/       # 독립 테스트
│
└── internal/shared/     # 최소 공유만
    ├── logger/         # 로그 일관성
    └── errors/         # 사용자 경험
```

### 장점
1. **완벽한 격리**
   - 한 Provider 버그가 다른 곳 영향 없음
   - 독립적 배포/롤백 가능

2. **Provider별 최적화**
   - 각자에게 맞는 최선의 구현
   - 타협 없는 성능

3. **단순함**
   - 새 개발자가 하나의 Provider만 이해하면 됨
   - 전체 시스템 이해 불필요

## 마이그레이션 전략

### experimental/ 디렉토리 활용
```
cli-recover/                    # 기존 구조 (유지)
experimental/                   # 새로운 실험
└── providers/
    └── filesystem_v2/         # 격리된 구현
```

### Feature Toggle
```go
if os.Getenv("USE_EXPERIMENTAL") == "true" {
    // 새로운 격리된 구현
} else {
    // 기존 구현 (변경 없음)
}
```

### 성공 판단 기준
- 코드 복잡도 감소
- 테스트 독립성 향상
- Provider별 자유도 증가
- 버그 격리 확인

## 교훈

### 아키텍처 진화의 교훈
1. **과도한 미래 대비는 독**
   - 필요할 때 추가하는 것이 쉬움
   - 불필요한 복잡도 제거가 어려움

2. **단순함이 최선**
   - 이해하기 쉬운 코드
   - 유지보수 용이
   - 버그 감소

3. **격리가 통합보다 나을 수 있다**
   - 잘못된 추상화의 비용
   - 중복 허용의 실용성
   - Provider별 독립성의 가치

### 리팩토링의 교훈
- 같은 디렉토리에서 리팩토링 → 혼란
- experimental/ 격리 → 명확한 경계
- 성공 증명 후 통합 → 안전한 진화

## 현재 상태 및 다음 단계

### 완료됨
- ✅ 2계층 아키텍처 구현
- ✅ 기본 기능 모두 동작
- ✅ 테스트 커버리지 52.9%

### 진행 중
- 🧪 Provider 격리 실험 준비
- 📝 experimental/ 전략 수립

### 미래
- Provider별 독립 구조
- 각자의 속도로 진화
- 실제 필요 기반 확장
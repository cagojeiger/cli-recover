# CLI-Recover 2.0 Vision Draft (Revised)

## Purpose
이 문서는 CLI-Recover의 새로운 아키텍처 비전을 담은 임시 문서입니다.
격리성 우선 설계 철학을 바탕으로 현실적인 접근을 추구합니다.

## 핵심 통찰
- **Filesystem restore ≈ cp**: tar 복원은 본질적으로 파일 복사
- **Provider 다양성**: MongoDB, MinIO, PostgreSQL은 완전히 다른 패러다임
- **공통 인터페이스의 한계**: 강제된 추상화는 오히려 복잡도 증가
- **격리성 > 재사용성**: 중복을 허용하되 완벽한 격리 추구

## 새로운 비전

### 1. Provider 격리 아키텍처
```
cli-recover/
├── providers/
│   ├── filesystem/              # 완전 독립 모듈
│   │   ├── backup.go           # tar 특화 구현
│   │   ├── restore.go          # 복원 특화 구현
│   │   ├── tui.go              # filesystem 전용 TUI
│   │   ├── cmd.go              # cobra 명령 정의
│   │   └── tests/              # 독립 테스트
│   │
│   └── mongodb/                 # 완전 독립 모듈 (향후)
│       ├── dump.go             # mongodump 특화
│       ├── restore.go          # mongorestore 특화
│       ├── tui.go              # MongoDB 전용 TUI
│       └── tests/              # 독립 테스트
│
└── internal/shared/             # 최소 공유
    ├── logger/                  # 로그 일관성
    └── errors/                  # 에러 구조
```

### 2. Provider별 독립 TUI

#### 자동화 대신 특화
```go
// filesystem/tui.go
type FilesystemTUI struct {
    // tar 특화 옵션들
    compressionSelector *tview.DropDown
    incrementalCheckbox *tview.Checkbox
    excludePatterns     *tview.InputField
}

// mongodb/tui.go  
type MongoDBTUI struct {
    // MongoDB 특화 옵션들
    oplogCheckbox      *tview.Checkbox
    collectionSelector *tview.List
    authForm          *tview.Form
}
```

각 Provider가 자신만의 TUI를 가짐:
- 공통 UI 컴포넌트 강제 없음
- Provider 특성에 맞는 최적 UX
- 독립적 진화 가능

### 3. 최소 조율 원칙

#### 공유는 정말 필요한 것만
```go
// internal/shared/logger/logger.go
type Logger interface {
    Info(msg string, fields ...Field)
    Error(msg string, err error)
}

// internal/shared/errors/base.go
type UserError struct {
    Message string
    Fix     string
}
```

#### Provider는 독립적으로
```go
// providers/filesystem/backup.go
func Backup(pod, path string, opts TarOptions) error {
    // Filesystem에 최적화된 구현
    // 다른 Provider 신경 안 씀
}

// providers/mongodb/dump.go
func Dump(uri string, opts DumpOptions) error {
    // MongoDB에 최적화된 구현
    // Filesystem과 완전 무관
}
```

### 4. 진행률 표시 통합
```go
// 모든 Provider가 공통으로 사용하는 진행률 시스템
type ProgressReporter interface {
    Start(operation string, total int64)
    Update(current int64, message string)
    Complete()
    Error(err error)
}

// CLI와 TUI 모두에서 동작
type UnifiedProgressReporter struct {
    cliReporter *CLIProgressBar
    tuiReporter *TUIProgressWidget
}
```

## 현실적 마이그레이션 전략

### Phase 4: Provider 격리 구조 (복잡도: 35/100)
1. Filesystem Provider 완전 독립화
2. 공유 코드는 logger, error만
3. 기존 인터페이스 제거

### Phase 5: 테스트 커버리지 90%
1. 격리된 구조로 테스트 단순화
2. Provider별 독립 테스트
3. 통합 테스트는 최소화

### Phase 6: 실제 수요 기반 확장
1. MongoDB Provider (요청 시)
2. MinIO Provider (필요 시)
3. 각각 완전 독립 구현

## 실용적 장기 전망

### 1. 안정적 운영
- 각 Provider가 독립적으로 안정화
- 버그가 격리되어 영향 최소화
- 담당자별 전문성 확보

### 2. 점진적 개선
- 실제 사용 피드백 기반 개선
- Provider별 다른 속도로 진화
- 필요한 기능만 추가

### 3. 유지보수 용이성
- 새 개발자 온보딩 간단
- Provider 하나만 이해하면 기여 가능
- 명확한 경계와 책임

## Trade-offs (의도적 선택)

### 수용하는 것
1. **코드 중복**
   - 각 Provider에 유사 코드 존재
   - 격리의 이점이 더 큼

2. **수동 일관성**
   - 자동화하지 않고 수동 유지
   - 복잡한 자동화보다 실용적

3. **바이너리 크기**
   - 약간의 증가 (무시 가능)
   - 안정성이 더 중요

### 얻는 것
1. **완벽한 격리**
   - 한 Provider 문제가 다른 곳 영향 없음
   - 독립적 배포/롤백

2. **단순함**
   - 전체 시스템 이해 불필요
   - Provider 하나만 알면 됨

3. **최적화 자유도**
   - 각 Provider별 최선의 구현
   - 타협 없는 성능

## 결론

> "Duplication is cheaper than the wrong abstraction"  
> — Sandi Metz

이 비전은 CLI-Recover를 과도한 추상화에서 벗어나
각 Provider가 최선의 성능을 발휘하는 도구로 만듭니다.

**격리성 우선, 중복 허용, 현실적 접근**

이것이 우리의 새로운 방향입니다.
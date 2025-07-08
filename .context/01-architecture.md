# 시스템 아키텍처

## Phase 3.9: 2계층 아키텍처 (2025-01-08)

### 아키텍처 단순화
- **이전**: 3계층 (Domain → Application → Infrastructure)
- **현재**: 2계층 (Domain ↔ Infrastructure)
- **이유**: Application 레이어가 단순 전달만 수행
- **효과**: 복잡도 75 → 30 목표

### 단순화된 디렉토리 구조
```
internal/
├── domain/              # 비즈니스 로직 & 인터페이스
│   ├── operation/      # backup/restore 통합
│   ├── metadata/       # 메타데이터
│   ├── logger/         # 로거 인터페이스
│   └── log/            # 작업 이력
└── infrastructure/      # 외부 시스템 연동 & 구현체
    ├── config/         # 설정 관리 (application에서 이동)
    ├── filesystem/     # 파일시스템 provider
    ├── kubernetes/     # K8s 클라이언트
    └── logger/         # 로거 구현체

cmd/cli-recover/        # CLI 진입점 (adapter 역할 포함)
├── backup.go          # 백업 명령 + adapter 로직
├── restore.go         # 복원 명령 + adapter 로직
├── list.go            # 목록 명령 + adapter 로직
└── tui/               # TUI 인터페이스
```

## 설계 원칙

### Occam's Razor 적용
1. **필요한 것만 구현**: filesystem provider만 완전 구현
2. **과도한 추상화 제거**: Registry 패턴 제거
3. **직접적인 호출**: 불필요한 중간 레이어 제거
4. **명확한 책임**: 각 패키지의 역할 명확화

### 2계층 책임 분리
```
Domain Layer (What & Why)
├── 비즈니스 규칙
├── 도메인 모델
├── 인터페이스 정의
└── 유스케이스

Infrastructure Layer (How)
├── 외부 시스템 통합
├── 구현체 제공
├── 설정 관리
└── 파일/네트워크 I/O
```

## 핵심 인터페이스

### 통합된 Operation Provider
```go
// domain/operation/provider.go
type Provider interface {
    // 공통 정보
    Name() string
    Type() OperationType  // Backup or Restore
    
    // 실행
    Execute(ctx context.Context, opts Options) error
    
    // 진행률
    StreamProgress() <-chan Progress
    
    // 검증
    Validate(opts Options) error
}

type OperationType int
const (
    BackupOperation OperationType = iota
    RestoreOperation
)
```

### 단순화된 Options
```go
// domain/operation/types.go
type Options struct {
    // 공통 필드
    Namespace   string
    Pod         string
    Container   string
    
    // Backup 전용
    SourcePath  string
    OutputFile  string
    Compression string
    
    // Restore 전용
    BackupFile  string
    TargetPath  string
    Overwrite   bool
}
```

## CLI 직접 통합

### Before (복잡)
```go
// cmd → adapter → registry → provider → infrastructure
cmd.Execute() → adapter.ExecuteBackup() → registry.Get() → provider.Execute()
```

### After (단순)
```go
// cmd → infrastructure
cmd.Execute() → filesystem.NewProvider().Execute()
```

## 로그 시스템 (Phase 3 완료)

### 구현 내용
- 작업별 로그 파일 자동 생성
- 메타데이터 JSON 저장
- CLI 명령어로 조회/관리
- 파일 기반 영구 저장

### 디렉토리 구조
```
~/.cli-recover/
├── logs/
│   ├── metadata/    # JSON 메타데이터
│   └── files/       # 실제 로그 파일
├── metadata/        # 백업 메타데이터
└── config.yaml      # 설정 파일
```

## TUI 계획 (Phase 4 예정)

### tview 기반 구현 계획
- CLI 명령어 래핑 방식
- 실시간 진행률 표시
- 메뉴 기반 네비게이션
- 목표: ~800줄 이내

### 예정 구조
```
cmd/cli-recover/tui/
├── app.go      # TUI 메인
├── menu.go     # 메인 메뉴
├── backup.go   # 백업 워크플로우
├── restore.go  # 복원 워크플로우
├── list.go     # 백업 목록
├── logs.go     # 로그 뷰어
└── progress.go # 진행률 표시
```

## 성공 지표

### Phase 3.9 목표
- [ ] 복잡도: 75 → 30
- [ ] 파일 수: 40% 감소
- [ ] 코드 라인: 35% 감소
- [ ] 디렉토리 깊이: 5 → 3단계

### 현재 달성
- ✅ Filesystem provider 완성
- ✅ 로그 시스템 구현
- ✅ 테스트 커버리지 53%
- ⏳ TUI 구현 예정 (Phase 4)

## 향후 계획

### 실사용 피드백 후
1. 필요시 MinIO provider 추가
2. 필요시 MongoDB provider 추가
3. 성능 최적화
4. 사용성 개선

### 유지 원칙
- YAGNI (You Aren't Gonna Need It)
- KISS (Keep It Simple, Stupid)
- DRY (Don't Repeat Yourself)
- 실사용자 피드백 기반 개발
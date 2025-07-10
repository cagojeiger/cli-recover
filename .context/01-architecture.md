# 시스템 아키텍처

> **주의**: 이 문서는 더 상세한 아키텍처 진화 과정을 담은 [01-architecture-evolution.md](01-architecture-evolution.md)로 대체되었습니다.
> 아래는 현재 아키텍처의 요약입니다.

## 현재: 2계층 아키텍처 (2025-01-08)

### 아키텍처 단순화
- **이전**: 3계층 (Domain → Application → Infrastructure)
- **현재**: 2계층 (Domain ↔ Infrastructure)
- **이유**: Application 레이어가 단순 전달만 수행
- **효과**: 복잡도 75 → 30 목표

### 최종 디렉토리 구조 (2025-01-09 업데이트)
```
internal/
├── domain/              # 비즈니스 로직 & 인터페이스 (순수 도메인)
│   ├── backup/         # 백업 인터페이스
│   ├── restore/        # 복원 인터페이스
│   ├── operation/      # 통합 어댑터 (미사용 - 제거 예정)
│   ├── metadata/       # 메타데이터 인터페이스만
│   ├── logger/         # 로거 인터페이스
│   ├── log/            # 작업 이력 도메인
│   └── progress/       # 진행률 타입
└── infrastructure/      # 외부 시스템 연동 & 구현체
    ├── config/         # 설정 관리
    ├── filesystem/     # 파일시스템 provider 구현
    ├── kubernetes/     # K8s 클라이언트
    ├── logger/         # 로거 구현체
    ├── progress/       # 진행률 구현체
    ├── metadata/       # 메타데이터 파일 저장소
    ├── log/           
    │   └── storage/   # 로그 파일 저장소
    ├── provider_factory.go      # Provider 팩토리
    └── provider_factory_test.go # 팩토리 테스트

cmd/cli-recover/        # CLI 진입점 (adapter 로직 통합)
├── main.go            # 메인 진입점
├── backup.go          # 백업 명령
├── backup_logic.go    # 백업 로직 (adapter 통합)
├── restore.go         # 복원 명령
├── restore_logic.go   # 복원 로직 (adapter 통합)
├── list.go            # 목록 명령
├── list_logic.go      # 목록 로직 (adapter 통합)
├── init.go            # 초기화 명령
├── logs.go            # 로그 명령
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
├── 순수 타입 정의
└── 외부 의존성 없음

Infrastructure Layer (How)
├── Domain 인터페이스 구현
├── 외부 시스템 통합
├── 파일/네트워크 I/O
├── 설정 관리
└── 실제 저장소 구현
```

### 의존성 방향 (헥사고날 아키텍처)
- CMD → Infrastructure → Domain
- Domain은 순수함 (no imports)
- Infrastructure가 Domain에 의존

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
// cmd → factory → provider
cmd.Execute() → infrastructure.CreateBackupProvider("filesystem") → provider.Execute()
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

### Phase 3.9 달성 (완료)
- ✅ 복잡도: 75 → ~30 달성
- ✅ 파일 수: ~40% 감소
- ✅ 코드 라인: ~35% 감소
- ✅ 디렉토리 깊이: 5 → 3단계
- ✅ Application 레이어 완전 제거
- ✅ Registry 패턴 → Factory 함수
- ✅ 모든 테스트 통과

### 전체 진행 상황
- ✅ Phase 1-3: 기본 기능 구현 완료
- ✅ Phase 3.9: 아키텍처 단순화 완료
- ✅ Filesystem provider 완성
- ✅ 로그 시스템 구현
- ✅ 테스트 커버리지 유지
- ⏳ Phase 4: TUI 구현 예정

## CLI 사용성 개선 (Phase 3.12 계획)

### 플래그 관리 중앙화
```go
// flags/registry.go
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

### 충돌 해결 전략
- **backup**: `-t` → `-T` (--totals)
- **restore**: `-o` → `-f` (--force), `-c` → `-C` (--container)
- **원칙**: POSIX/GNU 표준 준수

### 에러 처리 개선
```
❌ Error: Backup file not found
   Reason: The file 'backup.tar' does not exist
   Fix: Check the filename or use 'cli-recover list backups'
   See: https://docs.cli-recover.io/errors/file-not-found
```

## 향후 계획
### Phase 3.11: 진행률 보고 시스템 (완료) ✅
1. 3초 규칙 적용
2. 다중 환경 지원
3. 크기 추정 통합
4. 테스트 수정

### Phase 3.12: CLI 사용성 개선 (진행 예정)
1. 플래그 충돌 해결
2. 하이브리드 인자 처리
3. 에러 메시지 개선
4. 진행률 표시 통합

### Phase 3.13: 도구 자동 다운로드 (계획)
1. kubectl 자동 설치
2. 플랫폼별 바이너리 선택
3. 캐싱 메커니즘

### Phase 4: TUI 구현
1. CLI 명령 래핑
2. 실시간 진행률
3. 메뉴 네비게이션

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

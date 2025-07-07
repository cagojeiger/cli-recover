# 시스템 아키텍처

## 현재 아키텍처 분석

### 문제점
- **God Object 안티패턴**: Model struct가 115개 이상의 필드 보유
- **강한 결합**: UI와 비즈니스 로직이 혼재
- **중복 코드**: 각 화면이 비슷한 레이아웃을 반복 구현
- **메모리 누수**: BackupJob.Output이 무제한 증가
- **테스트 불가**: 구체적 구현에 의존하여 모킹 불가
- **확장성 부족**: 새 백업 타입 추가 시 다수 파일 수정 필요

### 현재 구조
```
Model (모든 것을 포함)
├── UI 상태 (selected, screen, etc.)
├── 비즈니스 상태 (jobs, backupOptions, etc.)
├── 데이터 (namespaces, pods, etc.)
└── 의존성 (runner, jobManager, etc.)
```

## 목표 아키텍처

### 레이어 분리
```
Presentation Layer (UI)
├── Components (재사용 가능한 UI 요소)
├── Screens (화면별 조합)
└── Layout Manager (레이아웃 관리)
    ↓ (인터페이스)
Domain Layer (비즈니스)
├── Services (비즈니스 로직)
├── Entities (도메인 모델)
└── Interfaces (포트)
    ↓ (인터페이스)
Infrastructure Layer (인프라)
├── KubernetesClient (kubectl 래핑)
├── FileSystem (파일 시스템 접근)
└── ProcessExecutor (명령 실행)
```

### 핵심 인터페이스
```go
// 도메인 레이어
type BackupService interface {
    CreateBackup(ctx context.Context, req BackupRequest) (*BackupJob, error)
    ListJobs() []*BackupJob
    GetJob(id string) (*BackupJob, error)
}

type BackupType interface {
    Name() string
    ValidateOptions(opts map[string]interface{}) error
    BuildCommand(target BackupTarget, opts map[string]interface{}) string
}

// 인프라 레이어
type KubernetesClient interface {
    GetNamespaces() ([]string, error)
    GetPods(namespace string) ([]Pod, error)
    GetContainers(namespace, pod string) ([]string, error)
    ExecCommand(namespace, pod, container string, cmd []string) error
}
```

### 컴포넌트 기반 UI
```go
type Component interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
    Focus() tea.Cmd
    Blur() tea.Cmd
}

// 재사용 가능한 컴포넌트
- ListComponent[T any] (제네릭 리스트)
- FormComponent (폼 입력)
- TableComponent (테이블 표시)
- JobListComponent (작업 목록)
```

## 데이터 플로우

### 백업 생성 플로우
```
User Input → Screen → Service → BackupType → KubernetesClient
                ↓                     ↓
            JobManager ← BackupJob ← CommandBuilder
```

### 상태 관리
```
Event (tea.Msg) → Update → New State → View
         ↓
    Side Effect (tea.Cmd)
```

## 메모리 관리 전략
- Ring Buffer로 출력 라인 수 제한 (최대 1000줄)
- 로그는 파일로 스트리밍 (~/.cli-recover/logs/)
- UI는 tail 방식으로 최근 라인만 표시

## 확장 포인트
- BackupType 인터페이스로 새 백업 타입 추가
- Component 인터페이스로 새 UI 요소 추가
- Middleware 패턴으로 공통 기능 추가 (로깅, 모니터링)

## 실용적 구현 전략

### Phase별 접근
1. **Phase 1 (2주)**: 긴급 문제 해결
   - Ring Buffer로 메모리 누수 해결
   - 기본적인 레이어 분리
   - 핵심 인터페이스 정의

2. **Phase 2 (3주)**: 컴포넌트화
   - UI 컴포넌트 추출
   - 재사용 가능한 패턴 확립
   - 테스트 커버리지 향상

3. **Phase 3 (4주)**: 플러그인 시스템
   - BackupType 플러그인화
   - 새 백업 타입 추가
   - 확장성 검증

4. **Phase 4 (2주)**: 최적화
   - 성능 개선
   - UX 향상
   - 문서화 완성

### 핵심 원칙
- **점진적 개선**: 한 번에 모든 것을 바꾸지 않음
- **하위 호환성**: 기존 사용자 영향 최소화
- **실용적 접근**: 과도한 추상화 지양
- **측정 가능**: 각 단계별 성공 지표 설정

## CLI-First 아키텍처 (전략 전환)

### 전환 배경
- 이미 동작하는 CLI 백업 기능 존재
- TUI보다 CLI가 더 핵심적인 가치 제공
- 테스트와 자동화에 유리한 구조

### CLI 레이어 구조
```
CLI Layer (사용자 인터페이스)
├── Commands (backup, restore, list, status)
├── Handlers (명령별 처리 로직)
└── Output (진행률, 에러 표시)
    ↓
Application Layer (비즈니스 로직)
├── Services (BackupService, RestoreService)
├── UseCases (명령별 유스케이스)
└── DTOs (데이터 전송 객체)
    ↓
Domain Layer (핵심 비즈니스)
├── Entities (Job, Backup, Progress)
├── Interfaces (BackupProvider, Storage)
└── ValueObjects (BackupOptions, Metadata)
    ↓
Infrastructure Layer (외부 연동)
├── Kubernetes (kubectl 추상화)
├── Providers (filesystem, minio, mongodb)
└── Storage (로컬 파일, 메타데이터)
```

### BackupProvider 플러그인 인터페이스
```go
type BackupProvider interface {
    // 기본 정보
    Name() string
    Description() string
    
    // 백업 실행
    Execute(ctx context.Context, opts BackupOptions) error
    
    // 크기 추정
    EstimateSize(opts BackupOptions) (int64, error)
    
    // 진행률 스트림
    StreamProgress() <-chan Progress
    
    // 옵션 검증
    ValidateOptions(opts BackupOptions) error
    
    // 복원 지원
    SupportsRestore() bool
}
```

### CLI 명령 체계
```bash
# 백업 명령
cli-recover backup <type> <pod> <path> [options]
  --namespace, -n    # Kubernetes 네임스페이스
  --output, -o       # 출력 파일명
  --compress, -c     # 압축 옵션
  --exclude          # 제외 패턴

# 복원 명령  
cli-recover restore <type> <pod> <backup-file> [options]
  --namespace, -n    # Kubernetes 네임스페이스
  --target, -t       # 복원 대상 경로

# 목록 조회
cli-recover list backups
  --format, -f       # 출력 포맷 (table, json)
  --filter           # 필터 조건

# 상태 확인
cli-recover status <job-id>
  --watch, -w        # 실시간 모니터링
```

### TUI 통합 전략
```
TUI Layer (Bubble Tea)
    ↓ (exec)
CLI Commands
    ↓
Business Logic
```

- TUI는 CLI 명령을 내부적으로 실행
- 진행률과 출력을 파싱하여 UI에 표시
- CLI와 TUI가 동일한 비즈니스 로직 공유

### 개발 우선순위
1. **CLI 완성**: 모든 기능을 CLI로 구현
2. **테스트 작성**: CLI 기반 통합 테스트
3. **문서화**: 명령어 사용법 문서
4. **TUI 래핑**: CLI 위의 시각적 레이어

### 성공 지표
- CLI로 모든 백업 작업 수행 가능
- 스크립트 자동화 지원
- CI/CD 파이프라인 통합 가능
- 테스트 커버리지 80% 이상

## 2025-01-07 현재 구현 상태

### 완료된 아키텍처 요소
- **Domain Layer**:
  - backup/restore provider 인터페이스
  - Registry 패턴 구현
  - Metadata store 인터페이스
  - Logger 인터페이스
- **Infrastructure Layer**:
  - Kubernetes client 추상화
  - Command executor 패턴
  - Filesystem provider 구현
  - Logger 구현체들 (file, console)
- **Application Layer**:
  - BackupAdapter (CLI → Domain)
  - RestoreAdapter (CLI → Domain)
  - ListAdapter (메타데이터 조회)

### 현재 디렉토리 구조 (문제점 포함)
```
cli-recover/
├── cmd/cli-recover/
│   ├── adapters/          # Application layer ✅
│   │   ├── backup_adapter.go
│   │   ├── restore_adapter.go
│   │   └── list_adapter.go
│   ├── backup_new.go      # CLI commands
│   ├── restore_new.go
│   └── list_new.go
├── internal/
│   ├── domain/           # Domain layer ✅
│   │   ├── backup/
│   │   ├── restore/
│   │   ├── metadata/
│   │   └── logger/
│   ├── infrastructure/   # Infrastructure layer ✅
│   │   ├── kubernetes/
│   │   ├── logger/
│   │   └── providers/
│   │       └── filesystem/
│   ├── backup/          # ❌ 중복 (삭제 필요)
│   ├── kubernetes/      # ❌ 중복 (삭제 필요)
│   ├── providers/       # ❌ 잘못된 위치 (infrastructure로 이동)
│   ├── runner/          # ❌ 잘못된 위치 (infrastructure로 이동)
│   ├── config/          # ⚠️ application layer로 이동 고려
│   └── presentation/    # ❌ 빈 디렉토리 (삭제)
└── .memory/             # AI memory system
    ├── short-term/
    └── long-term/
```

### 아키텍처 위반 사항
1. **중복 패키지**:
   - internal/backup/ vs internal/domain/backup/
   - internal/kubernetes/ vs internal/infrastructure/kubernetes/
   - internal/providers/ vs internal/infrastructure/providers/

2. **잘못된 위치**:
   - internal/runner/ → internal/infrastructure/runner/
   - internal/config/ → internal/application/config/

3. **빈 디렉토리**:
   - internal/presentation/ (TUI 삭제로 불필요)

### 아키텍처 준수 평가
- ✅ 레이어 분리 완료
- ✅ 인터페이스 기반 설계
- ✅ 의존성 역전 원칙
- ✅ Provider 플러그인 패턴
- ✅ TUI 레이어 완전 제거
- ❌ 레거시 중복 코드 존재

## 향후 구조 (Job 도메인 추가)
```
internal/
├── domain/              # 비즈니스 로직
│   ├── backup/
│   ├── restore/
│   ├── metadata/
│   ├── logger/
│   └── job/            # 새로 추가
├── infrastructure/      # 외부 시스템 연동
│   ├── kubernetes/
│   ├── logger/
│   ├── providers/
│   ├── process/        # 새로 추가
│   ├── storage/        # 새로 추가
│   └── runner/         # 이동
└── application/        # 애플리케이션 서비스
    ├── config/         # 이동
    └── service/        # 새로 추가
```
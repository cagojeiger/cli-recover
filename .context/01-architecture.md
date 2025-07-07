# 시스템 아키텍처

## 현재 아키텍처 (2025-01-07 현재)

### 완료된 개선
- **TUI 완전 제거**: God Object 안티패턴 해결
- **헥사고날 아키텍처**: 명확한 레이어 분리
- **Provider 패턴**: 확장 가능한 구조
- **인터페이스 기반**: 테스트 용이
- **의존성 주입**: Mock 가능

### 현재 디렉토리 구조
```
internal/
├── domain/              # 비즈니스 로직
│   ├── backup/         # 백업 도메인
│   ├── restore/        # 복원 도메인
│   ├── metadata/       # 메타데이터
│   ├── logger/         # 로거 인터페이스
│   └── log/            # 작업 이력 도메인
├── infrastructure/      # 외부 시스템 연동
│   ├── kubernetes/     # K8s 클라이언트
│   ├── logger/         # 로거 구현체
│   ├── providers/      # 백업 프로바이더
│   └── runner/         # 명령 실행기
└── application/        # 애플리케이션 서비스
    ├── adapters/       # CLI 어댑터
    └── config/         # 설정 관리
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

## Phase 3 완료: 로그 파일 시스템

### 구현 내용 (복잡도 30/100)
1. **Log 도메인 모델**
   - internal/domain/log/ 패키지
   - Log 엔티티 (작업 이력)
   - LogRepository 인터페이스
   - 파일 기반 저장소

2. **CLI 명령어**
   - logs list: 작업 이력 조회
   - logs show: 로그 상세 보기
   - logs tail: 로그 마지막 부분 보기
   - logs clean: 오래된 로그 정리

3. **통합 기능**
   - 백업/복구 시 자동 로그 생성
   - 작업별 고유 ID 부여
   - 상태 추적 (running, completed, failed)

4. **파일 구조**
   ```
   ~/.cli-recover/logs/
   ├── metadata/    # JSON 메타데이터
   └── files/       # 실제 로그 파일
   ```

### Phase 4: TUI 재구현 (2주)
- CLI 래퍼 방식
- 단순한 UI로만
- 비즈니스 로직 분리

### Phase 5: Provider 확장 (4주)
- MinIO Provider
- MongoDB Provider  
- PostgreSQL Provider
- MySQL Provider

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

## 주요 인터페이스

### BackupProvider
```go
type BackupProvider interface {
    Name() string
    Description() string
    Execute(ctx context.Context, opts BackupOptions) error
    EstimateSize(opts BackupOptions) (int64, error)
    StreamProgress() <-chan Progress
    ValidateOptions(opts BackupOptions) error
}
```

### RestoreProvider
```go
type RestoreProvider interface {
    Name() string
    Description() string
    Execute(ctx context.Context, opts RestoreOptions) error
    ValidateBackup(metadata *Metadata) error
    StreamProgress() <-chan Progress
}
```

### LogRepository (Phase 3 구현됨)
```go
type LogRepository interface {
    Save(log *Log) error
    Get(id string) (*Log, error)
    List(filter ListFilter) ([]*Log, error)
    Update(log *Log) error
    Delete(id string) error
    GetLatest(filter ListFilter) (*Log, error)
}
```

## 성공 지표

### 현재 달성
- ✅ 헥사고날 아키텍처 100% 준수
- ✅ 모든 의존성 인터페이스화
- ✅ 새 provider 추가 < 200 LOC
- ✅ 테스트 커버리지 53.0%

### Phase 3 달성
- ✅ 작업 이력 영구 보관
- ✅ 로그 파일 자동 생성
- ✅ 오래된 로그 정리 기능
- ✅ 복잡도 30/100 유지

### Phase 4 달성
- ✅ TUI 구현 (tview 사용)
- ✅ CLI 래퍼 방식
- ✅ 복잡도 40/100 유지
- ✅ God Object 회피 성공
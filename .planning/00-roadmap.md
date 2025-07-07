# CLI-First 개발 로드맵

## 전략 전환
- TUI 중심 → CLI 우선 개발
- 이유: 동작하는 CLI 백업이 이미 구현되어 있고, 더 실용적
- 원칙: "Make it work → Make it right → Make it pretty"
- **전략 수정 (2025-01-07)**: MinIO/MongoDB Provider를 나중으로 미루고 핵심 기능에 집중

## 현재 상태 (95% 완료)
- ✅ Filesystem 백업 Provider 완전 구현
- ✅ Filesystem 복원 Provider 구현 완료
- ✅ 메타데이터 저장 시스템 구현
- ✅ List 명령 구현 (백업 목록 조회)
- ✅ 아키텍처 기반 구축 (Hexagonal + Plugin)
- ✅ CLI 프레임워크 통합 완료
- ✅ TDD 방식으로 높은 테스트 커버리지
- ✅ kubectl exec + tar 통합 완료
- ✅ 진행률 모니터링 구현
- ✅ 크기 추정 & ETA 계산 가능
- ✅ 코드 정리 및 중복 제거 완료
- ✅ 테스트 커버리지 58.4% 달성

## Phase 1: CLI 핵심 완성 (95% 완료)
**복잡도**: 25/100 ✅
**목표**: 완전한 백업/복원 도구로 만들기
**진행률**: 95%

### 작업 항목
- [x] CLI 명령 체계 표준화
- [x] Filesystem backup provider
- [x] 공통 진행률 처리
- [x] 에러 처리 통일
- [x] **Restore 명령 구현** ✅
- [x] **List 명령 (백업 목록)** ✅
- [x] **메타데이터 저장 시스템** ✅
- [x] **코드 정리 및 중복 제거** ✅
- [x] **테스트 커버리지 개선** ✅ (58.4%)
- [ ] **Status 명령 (Phase 4로 이동)**

### 명령 구조
```bash
# Backup
cli-recover backup filesystem <pod> <path> [options]

# Restore (구현 완료)
cli-recover restore filesystem <pod> <backup-file> [options]

# List (구현 완료)
cli-recover list backups [--namespace <ns>]

# Status (미구현)
cli-recover status <job-id>
```

### 성공 지표
- ✅ 백업/복원 전체 사이클 동작
- ✅ 백업 목록 조회 가능
- ✅ 코드베이스 정리 완료
- ✅ 테스트 커버리지 목표 근접 (58.4%/60%)
- ⏳ 작업 상태 추적 (Phase 4에서 구현)

## Phase 2: 아키텍처 고도화 (95% 완료)
**복잡도**: 30/100 ✅
**목표**: Restore 기능을 위한 아키텍처 확장

### 작업 항목
- [x] BackupProvider 인터페이스 정의
- [x] Provider 레지스트리 구현
- [x] 도메인 레이어 분리
- [x] Infrastructure 레이어 구성
- [x] **RestoreProvider 인터페이스** ✅
- [x] **메타데이터 도메인 모델** ✅
- [ ] **작업 추적 시스템** (Status와 함께 Phase 4로)

### 구현된 인터페이스
```go
// Backup Provider
type BackupProvider interface {
    Name() string
    Execute(ctx context.Context, opts BackupOptions) error
    EstimateSize(opts BackupOptions) (int64, error)
    StreamProgress() <-chan Progress
}

// Restore Provider
type RestoreProvider interface {
    Name() string
    Execute(ctx context.Context, opts RestoreOptions) error
    ValidateBackup(backupFile string) error
    StreamProgress() <-chan Progress
}
```

### 메타데이터 구조 (구현 완료)
```go
type BackupMetadata struct {
    ID          string
    Type        string
    Namespace   string
    PodName     string
    SourcePath  string
    BackupFile  string
    Size        int64
    CreatedAt   time.Time
    CompletedAt time.Time
    Status      string
}
```

### 성공 지표
- ✅ 새 provider 추가 < 200 LOC
- ✅ 테스트 커버리지 근접 (58.4%)
- ✅ 모든 의존성 인터페이스화
- ✅ 메타데이터 시스템 완성
- ⏳ 작업 추적 시스템 (Phase 4에서)

## Phase 3: CLI 고도화 (다음 단계)
**복잡도**: 40/100 ⚠️
**목표**: 프로덕션 레벨 기능 추가

### 작업 항목
- [ ] status 명령 (작업 상태)
- [ ] 설정 파일 지원 (~/.cli-recover/config.yaml)
- [ ] 구조화된 로깅 시스템
- [ ] 백업 압축 옵션 개선
- [ ] 백업 검증 기능
- [ ] 재시도 로직
- [ ] 백업 만료 정책

### 추가 기능
- 백업 암호화 지원
- 대역폭 제한
- 병렬 처리
- 프로그레스 바 개선

### 성공 지표
- 설정 파일로 관리 가능
- 로그 레벨 조절 가능
- 에러 시 자동 재시도
- 백업 무결성 검증

## Phase 4: TUI 통합 (계획)
**복잡도**: 45/100 ⚠️
**목표**: CLI 위에 사용자 친화적 TUI 구축

### 작업 항목
- [ ] CLI 명령 래핑 레이어
- [ ] 백업 타입 선택 화면
- [ ] 파드/경로 선택 UI
- [ ] 실시간 진행률 뷰
- [ ] 백업 목록 브라우저
- [ ] 복원 마법사

### TUI 화면 구성
```
1. Main Menu
   - Backup
   - Restore
   - List Backups
   - Settings

2. Backup Flow
   - Select Type (filesystem)
   - Select Pod
   - Select Path
   - Options
   - Execute with Progress

3. Restore Flow
   - Select Backup
   - Select Target Pod
   - Confirm
   - Execute with Progress
```

### 성공 지표
- TUI에서 모든 CLI 기능 사용 가능
- 부드러운 UI 전환
- 실시간 상태 업데이트
- 키보드 단축키 지원

## Phase 5: Provider 확장 (계획)
**복잡도**: 60/100 ⚠️⚠️
**목표**: 다양한 백업 타입 지원

### 작업 항목
- [ ] MinIO Provider
  - S3 프로토콜 지원
  - 버킷 백업/복원
- [ ] MongoDB Provider
  - mongodump/mongorestore
  - 컬렉션 선택
- [ ] PostgreSQL Provider
  - pg_dump/pg_restore
  - 스키마/데이터 옵션
- [ ] MySQL Provider
  - mysqldump/mysql
  - 데이터베이스 선택

### Provider 추가 시
- CLI와 TUI 자동 지원
- 통일된 진행률/에러 처리
- 메타데이터 자동 관리

## 일정 요약 (수정됨)
- **Phase 1 완료**: 1월 2주 (85% 완료)
- **Phase 2 완료**: 1월 3주 초 (90% 완료)
- **Phase 3**: 1월 3-4주
- **Phase 4**: 2월 1-2주
- **Phase 5**: 2월 3주 이후
- **총 기간**: 6-7주

## 주요 변경사항
1. **MinIO/MongoDB를 Phase 5로 이동**
   - 복잡도 감소
   - 빠른 핵심 기능 완성
   - TUI 완성 후 한번에 지원

2. **Restore/List/Status 우선 구현**
   - 실용적 가치 제공
   - 완전한 백업 도구로 만들기
   - 사용자 피드백 조기 확보

3. **메타데이터 시스템 추가**
   - 백업 추적 가능
   - 복원 시 검증
   - 향후 고급 기능 기반

## 위험 관리
- **Restore 복잡도**: 단계적 구현으로 관리
- **메타데이터 저장**: 로컬 파일 시스템 사용
- **Provider 확장성**: 인터페이스로 보장됨

## 성공 지표
- v0.3.0: Restore 기능 포함 (1월 말)
- v0.4.0: CLI 고도화 완료 (2월 초)
- v1.0.0: TUI 통합 완성 (2월 중)
- v1.1.0: 추가 Provider 지원 (2월 말)
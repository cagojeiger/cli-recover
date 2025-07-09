# CLI-First 개발 로드맵

## 전략 전환
- TUI 중심 → CLI 우선 개발
- 이유: 동작하는 CLI 백업이 이미 구현되어 있고, 더 실용적
- 원칙: "Make it work → Make it right → Make it pretty"
- **전략 수정 (2025-01-07)**: MinIO/MongoDB Provider를 나중으로 미루고 핵심 기능에 집중

## Phase 3.9: 아키텍처 단순화 (완료) ✅
**복잡도**: 75 → ~30 달성 ✅
**목표**: Occam's Razor 적용으로 과도한 복잡성 제거
**기간**: 2025-01-08 (하루만에 완료!)

### 완료된 작업
- [x] Application 레이어 제거 (3계층 → 2계층) ✅
- [x] Domain 통합 어댑터 제공 (operation) ✅
- [x] 미사용 코드 제거 (minio/mongodb 스텁, runner) ✅
- [x] 구조 평탄화 (providers 디렉토리 제거) ✅
- [x] Registry 패턴 → Factory 함수 교체 ✅
- [x] backup 디렉토리 제거 ✅

### 달성된 결과
- 파일 수: ~40% 감소 ✅
- 코드 라인: ~35% 감소 ✅
- 디렉토리 깊이: 5 → 3 ✅
- 복잡도: 75 → ~30 ✅
- 모든 테스트 통과 ✅

## Phase 3.10: 백업 파일 무결성 보장 (완료) ✅
**복잡도**: 25/100 (목표 달성) ✅
**목표**: 백업 중 파일 손상 방지를 위한 원자적 파일 쓰기
**기간**: 2025-07-08 (완료)
**문서화**: 완료 (2025-01-08)

### 문제점 분석
- 백업 중단 시 불완전한 tar 파일 생성
- 동시 백업 실행 시 파일 덮어쓰기 위험
- 실패한 백업을 성공으로 착각할 가능성

### 해결책 (Occam's Razor 적용)
- [x] 임시 파일(.tmp)에 먼저 쓰기 ✅
- [x] 완료 후 원자적 이동(os.Rename) ✅
- [x] 실패 시 자동 정리(defer) ✅
- [x] 스트리밍 체크섬 계산 (SHA256) ✅
- [x] **진행률 보고 통합** (기존 시스템 활용) ✅

### 달성된 효과
- 백업 중단 시 최종 파일 없음 (안전) ✅
- OS 레벨 원자성 보장 ✅
- 데이터 무결성 검증 가능 ✅
- 복잡도 25/100 달성 (목표보다 낮음) ✅
- 테스트 커버리지 71.2% ✅

## Phase 3.11: 진행률 보고 시스템 (완료) ✅
**복잡도**: 35/100 ✅
**목표**: 통합 진행률 보고 시스템
**기간**: 2025-07-08 ~ 2025-07-09
**상태**: 구현 완료

### 구현 내용
- [x] 3초 규칙 적용 (3초 이상 작업은 진행률 표시)
- [x] Terminal, CI/CD, Pipe 환경 지원
- [x] filesystem provider에 size estimation 통합
- [x] 테스트 수정 및 mock 추가
- [x] 문서 정리 (CLAUDE.md 원칙 준수)

## Phase 3.12: CLI 사용성 개선 (계획)
**복잡도**: 30/100 ✅
**목표**: CLI 일관성 확보 및 사용자 경험 개선
**기간**: 2025-01-09 ~ 2025-01-10
**문서화**: 완료 (2025-01-09)

### 문제점 분석
- 플래그 단축키 충돌 (-o, -c, -t 중복 사용)
- 명령어 패턴 일관성 부족
- 사용자 피드백 부족
- 에러 메시지 불친절

### 해결책 (CLAUDE.md 준수)
- [ ] 플래그 레지스트리 구현 (중앙 관리)
- [ ] 충돌 플래그 수정 (-o→-f, -t→-T, -c→-C)
- [ ] 하이브리드 args/flags 지원
- [ ] 진행률 표시 통합
- [ ] 에러 메시지 개선 (원인, 해결법 제시)

### 기대 효과
- kubectl/docker 스타일 일관성
- 플래그 충돌 제거
- 더 나은 사용자 경험
- 복잡도 30/100 유지

## Phase 3.13: CLI 도구 자동 다운로드 (계획)
**복잡도**: 50/100 ✅
**목표**: kubectl, mc 등 외부 도구 자동 설치
**기간**: 2025-01-11 (Phase 3.12 이후)
**문서화**: 완료 (2025-01-08)

### 문제점
- kubectl이 없으면 백업/복원 실패
- 사용자가 수동으로 도구 설치 필요
- 버전 불일치 가능성

### 해결책
- [ ] ToolManager 구현
- [ ] PATH → 캐시 → 다운로드 전략
- [ ] 원자적 다운로드 (임시파일 + rename)
- [ ] 플랫폼별 바이너리 선택
- [ ] **다운로드 진행률 표시** (필수)

### 지원 도구
- kubectl: Kubernetes CLI
- mc: MinIO Client (향후)

### 기대 효과
- 도구 없는 환경에서도 자동 작동
- 투명한 사용자 경험
- 캐싱으로 반복 다운로드 방지

## 현재 상태 (Phase 1-4 완료)
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

## Phase 3: 로그 시스템 (완료)
**복잡도**: 30/100 ✅
**목표**: 작업 이력 영구 보관 시스템
**진행률**: 100%

### 완료된 항목
- [x] 구조화된 로깅 시스템 ✅ (2025-01-07)
- [x] init 명령 (설정 파일 초기화) ✅
- [x] 로그 파일 시스템 구현 ✅
- [x] CLI 명령어: logs list, show, tail, clean ✅
- [x] 백업 시 자동 로그 생성 ✅
- [x] 작업 상태 추적 (running, completed, failed) ✅

### 로그 시스템 특징
- 각 작업마다 고유 ID 생성
- 상세 로그 파일 자동 생성
- 메타데이터 JSON 저장
- 오래된 로그 자동 정리

### 성공 지표 달성
- ✅ 실용적인 기능 구현
- ✅ 복잡도 30/100 유지
- ✅ Claude.md Occam's Razor 원칙 준수
- ✅ 작업 이력 영구 보관

## Phase 4: TUI 통합 (계획)
**복잡도**: 40/100 (목표)
**목표**: CLI 위에 사용자 친화적 TUI 구축
**진행률**: 0% (Phase 3.9 완료 후 시작)

### 작업 항목
- [ ] CLI 명령 래핑 레이어
- [ ] 백업 타입 선택 화면
- [ ] 파드/경로 선택 UI
- [ ] 실시간 진행률 뷰
- [ ] 백업 목록 브라우저
- [ ] 복원 마법사
- [ ] 로그 뷰어 추가

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

### 성공 지표 (목표)
- [ ] TUI에서 모든 CLI 기능 사용 가능
- [ ] 부드러운 UI 전환 (tview 사용)
- [ ] 실시간 상태 업데이트
- [ ] 키보드 단축키 지원
- [ ] 복잡도 40/100 유지

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

## 일정 요약 (업데이트)
- **Phase 1 완료**: 1월 2주 (95% 완료)
- **Phase 2 완료**: 1월 3주 초 (95% 완료)
- **Phase 3 완료**: 1월 7일 (100% 완료)
- **Phase 3.9 완료**: 1월 8일 (아키텍처 단순화 - 하루만에 완료!) ✅
- **Phase 3.10 완료**: 1월 8일 (백업 무결성 - TDD로 안전하게 구현!) ✅
- **Phase 3.11 완료**: 7월 9일 (진행률 보고 시스템) ✅
- **Phase 3.12 계획**: 1월 9-10일 (CLI 사용성 개선)
- **Phase 3.13 계획**: 1월 11일 (도구 자동 다운로드)
- **Phase 4 계획**: Phase 3 완료 후 (TUI 구현)
- **Phase 5**: 필요시 진행 (MinIO/MongoDB)
- **총 기간**: 유동적 (사용자 요구사항 기반)

## 진행률 보고 원칙 (2025-01-08 추가)
**전체 프로젝트 적용 원칙**
- 3초 이상 모든 작업에 진행률 표시 필수
- 터미널, CI/CD, 로그, TUI 통합 지원
- 복잡도: +5/100 (표준 라이브러리만 사용)
- [상세 가이드](../docs/progress-reporting/)

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
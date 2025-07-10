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

## Phase 3.12: CLI 사용성 개선 (완료) ✅
**복잡도**: 30/100 ✅
**목표**: CLI 일관성 확보 및 사용자 경험 개선
**기간**: 2025-01-09 (당일 완료!)
**문서화**: 완료 (2025-01-09)

### 문제점 분석
- 플래그 단축키 충돌 (-o, -c, -t 중복 사용)
- 명령어 패턴 일관성 부족
- 사용자 피드백 부족
- 에러 메시지 불친절

### 해결책 (CLAUDE.md 준수)
- [x] 플래그 레지스트리 구현 (중앙 관리) ✅
- [x] 충돌 플래그 수정 (-o→-f, -t→-T, -c→-C) ✅
- [x] 에러 메시지 개선 (CLIError 타입) ✅
- [x] 진행률 표시 통합 ✅
- [ ] 하이브리드 args/flags 지원 → Phase 6로 이동

### 달성된 효과
- 플래그 충돌 완전 제거 ✅
- 컴파일 타임 충돌 검증 ✅
- 사용자 친화적 에러 메시지 ✅
- 복잡도 30/100 달성 ✅

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

## Phase 4: Provider 격리 구조 - 안전한 실험 (수정됨)
**복잡도**: 20/100 (낮춤) ✅
**목표**: 명확한 경계를 유지하며 Provider 격리 실험
**전략**: Experimental 디렉토리 활용
**진행률**: 0%

### 새로운 접근 방식

#### Phase 4-1: Experimental 구조 생성 (1일)
- [ ] experimental/providers/ 디렉토리 생성
- [ ] 명확한 README 작성 ("THIS IS EXPERIMENTAL")
- [ ] 기존 코드 전혀 수정 안 함
- [ ] .gitignore에 실험 설정 추가

#### Phase 4-2: 최소 기능 구현 (2일)
- [ ] EstimateSize 같은 단순 기능부터
- [ ] TDD로 새로 작성 (기존 코드 참고만)
- [ ] 완전히 독립적 구현
- [ ] 테스트 커버리지 100% 목표

#### Phase 4-3: A/B 테스트 인프라 (1일)
- [ ] 환경변수 기반 전환 (USE_EXPERIMENTAL)
- [ ] 로깅으로 사용 추적
- [ ] 성능 비교 측정
- [ ] 안전한 폴백 메커니즘

#### Phase 4-4: 평가 및 결정 (1일)
- [ ] 실험 성공/실패 판단
- [ ] 코드 복잡도 비교
- [ ] 성능 메트릭 분석
- [ ] 계속 진행 or 롤백 결정

### 디렉토리 구조 (실험적)
```
cli-recover/                    # 기존 구조 (변경 없음)
├── internal/
│   └── infrastructure/
│       └── filesystem/        # 현재 프로덕션 코드
│
experimental/                   # 새로운 실험 공간
└── providers/
    └── filesystem_v2/         # 명확한 버전 표시
        ├── README.md          # "EXPERIMENTAL" 경고
        ├── backup.go
        ├── restore.go
        └── tests/
```

### 성공 지표
- [ ] experimental/ 디렉토리 명확히 구분
- [ ] 기존 코드 영향 0%
- [ ] 언제든 롤백 가능 (폴더 삭제만으로)
- [ ] 코드 위치 100% 명확
- [ ] 팀원 혼란 0%

### 리스크 관리
- 실험 실패 시: experimental/ 폴더 삭제
- 부분 성공 시: 좋은 아이디어만 선택적 적용
- 완전 성공 시: 점진적 마이그레이션 계획 수립

### 참고
- [07-refactoring-strategy.md](../.context/07-refactoring-strategy.md) - 안전한 리팩토링 전략

## Phase 5: TUI 구현 (계획)
**복잡도**: 40/100 (목표)
**목표**: Provider별 독립 TUI 구현
**철학**: 각 Provider에 최적화된 UI
**진행률**: 0%

### 작업 항목
- [ ] Filesystem TUI 구현
- [ ] MongoDB TUI 구현 (필요 시)
- [ ] Provider별 독립적 UI/UX
- [ ] 공통 UI 컴포넌트 최소화
- [ ] CLI 명령과 연동

### 설계 원칙
- 각 Provider가 자신의 TUI 소유
- 공통 UI 강제 없음
- Provider 특성에 맞는 UX

## Phase 6: 테스트 커버리지 90% (계획)
**복잡도**: 20/100 ✅
**목표**: CLAUDE.md RULE_04 준수
**현재**: 52.9% → 목표: 90%

### 작업 항목
- [ ] Provider별 독립 테스트 작성
- [ ] 격리된 구조로 테스트 단순화
- [ ] 통합 테스트는 최소화
- [ ] 커버리지 90% 달성

### 격리된 테스트의 이점
- 한 Provider 테스트가 다른 곳 영향 없음
- 병렬 테스트 실행 가능
- 빠른 피드백 사이클
- Provider별 특화 테스트

## Phase 7: Provider 확장 (실제 수요 기반)
**복잡도**: 35/100 (Provider당) ✅
**목표**: 필요 시 독립 Provider 추가
**철학**: 각 Provider는 완전히 독립적

### 작업 항목 (실제 요청 시)
- [ ] MongoDB Provider (사용자 요청 시)
  - mongodump/mongorestore 래핑
  - MongoDB 전용 TUI
  - 완전 독립 구현
- [ ] MinIO Provider (필요 시)
  - mc (MinIO Client) 통합
  - S3 API 활용
  - MinIO 전용 TUI
- [ ] 각 Provider는 독립 모듈

## 일정 요약 (업데이트 - 철학 반영)
- **Phase 1-3 완료**: 기본 기능 구현 ✅
- **Phase 3.9-3.12 완료**: 아키텍처 개선 ✅
- **Phase 3-1 완료**: restore 긴급 수정 ✅
- **Phase 4 계획**: Provider 격리 구조 (복잡도: 35/100)
- **Phase 5 계획**: TUI 구현 (Provider별 독립)
- **Phase 6 계획**: 테스트 커버리지 90%
- **Phase 7 계획**: 실제 수요 기반 Provider 확장
- **총 기간**: 유동적 (격리성 우선)

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

## 성공 지표 (수정)
- v0.4.0: Provider 격리 구조 완성
- v0.5.0: Filesystem TUI 추가
- v0.6.0: 테스트 커버리지 90%
- v1.0.0: 안정적 운영 가능
- v2.0.0: MongoDB/MinIO 추가 (수요 기반)

## 설계 철학
> "Duplication is cheaper than the wrong abstraction"  
> — Sandi Metz

> "Isolation with minimal coordination"  
> — 우리의 결론

자세한 내용은 [94-design-philosophy.md](94-design-philosophy.md) 참조.

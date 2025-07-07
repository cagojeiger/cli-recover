# 주요 설계 결정사항

## 프레임워크 선택

### Bubble Tea 유지 (2024-01-07)
**결정**: 기존 Bubble Tea 프레임워크 유지
**이유**:
- 마이그레이션 비용 > 기대 이익
- 이미 작동하는 코드베이스 존재
- Charm 생태계 활용 가능 (lipgloss, bubbles)
- teatest로 테스트 가능

**대안 검토**:
- tview: k9s가 사용하지만 전면 재작성 필요
- 직접 구현: 복잡도 너무 높음

## 아키텍처 패턴

### 헥사고날 아키텍처 채택
**결정**: Ports & Adapters 패턴 적용
**이유**:
- 비즈니스 로직과 인프라 분리
- 테스트 용이성
- 확장성 (새 백업 타입 추가 쉬움)

### 점진적 리팩토링
**결정**: Big Bang 대신 단계별 접근
**이유**:
- 위험 최소화
- 기능 유지하면서 개선
- 각 단계별 검증 가능

## 기술적 결정

### Ring Buffer 도입
**결정**: 무제한 배열 대신 순환 버퍼 사용
**이유**:
- 메모리 사용량 예측 가능
- 대용량 백업 시 OOM 방지
- 최근 N개 라인만 유지로 충분

### TDD 우선순위
**결정**: 비-UI 로직부터 TDD 적용
**순서**:
1. Ring Buffer
2. 비즈니스 서비스
3. Repository
4. UI 컴포넌트 (teatest)

**이유**:
- UI 테스트는 상대적으로 어려움
- 핵심 로직 안정성 우선

### 인터페이스 기반 설계
**결정**: 구체 타입 대신 인터페이스 의존
**예시**:
```go
// Before
type Model struct {
    runner runner.Runner // 구체 타입
}

// After  
type Model struct {
    kubeClient KubernetesClient // 인터페이스
}
```
**이유**:
- 테스트 시 모킹 가능
- 구현 교체 용이
- 의존성 역전 원칙

## 프로세스 결정

### 백업 타입별 독립성
**결정**: 각 백업 타입은 독립적인 프로세스
**구현**:
```go
type BackupType interface {
    Name() string
    ValidateOptions(opts) error
    BuildCommand(target) string
}
```
**이유**:
- filesystem, minio, mongodb 각각 다른 요구사항
- 플러그인 방식으로 확장 가능
- 기존 타입 영향 없이 새 타입 추가

### 설정 관리
**결정**: ~/.cli-recover/config.yaml 사용
**이유**:
- 사용자 커스터마이징 가능
- 환경별 설정 분리
- viper로 쉬운 관리

## UI/UX 결정

### 프로그레스바 제거
**결정**: 복잡한 파싱 대신 원본 출력 표시
**이유**:
- 백업 타입마다 출력 형식 다름
- 파싱 로직 복잡도 높음
- 사용자가 원본 정보 선호

### 컴포넌트 기반 UI
**결정**: 재사용 가능한 컴포넌트 추출
**컴포넌트**:
- ListComponent
- FormComponent
- TableComponent
**이유**:
- 코드 중복 제거
- 일관된 UX
- 유지보수 용이

## 2025-01-07 추가 결정사항

### CLI-First 전략 채택
**결정**: TUI-first에서 CLI-first로 전환
**이유**:
- TUI Model이 God Object화 (115+ 필드)
- 테스트 어려움
- 복잡도 관리 실패
**구현**:
- Provider 기반 CLI 명령 구조
- TUI는 CLI 래퍼로만 사용
- MinIO/MongoDB는 Phase 5로 연기

### TUI 테스트 제외
**결정**: TUI 레이어 테스트 작성 안함
**이유**:
- CLI가 테스트 가능하므로 중복
- TUI는 단순 래퍼
- 투자 대비 효과 낮음
**영향**:
- Makefile에서 TUI 커버리지 제외
- 비즈니스 로직 테스트에 집중

### 메타데이터 저장소
**결정**: 로컬 파일 기반 저장
**구현**:
- ~/.cli-recover/metadata/
- JSON 형식
- SHA256 체크섬 포함
**이유**:
- 단순함 우선
- 외부 의존성 없음
- 향후 확장 가능

### Status 명령 연기
**결정**: Phase 1에서 Phase 4로 이동
**이유**:
- Job 히스토리와 함께 구현이 효율적
- 필수 기능이 아님
- Phase 1을 빠르게 마무리
**영향**:
- Phase 1이 95% 완료로 마무리
- 핵심 백업/복원 기능에 집중

### 코드 정리 전략
**결정**: TDD 방식으로 레거시 코드 제거
**절차**:
1. 호환성 테스트 작성
2. 기능 동일성 검증
3. 안전한 제거
**결과**:
- backup_filesystem.go 안전하게 제거
- 테스트 커버리지 58.4% 달성
- 깨끗한 코드베이스 확보

## 2025-01-07 Phase 3 진행

### 로그 시스템 구현
**발견**: 로그 시스템이 이미 완전히 구현되어 있음
**위치**:
- internal/domain/logger: 인터페이스 정의
- internal/infrastructure/logger: 구현체들
**기능**:
- 파일/콘솔 로거
- 로그 로테이션
- JSON/Text 포맷
- 글로벌 로거 지원
**작업**:
- 기존 fmt.Printf를 로거로 교체
- CLI 플래그 추가 (--log-level, --log-file, --log-format)
- 테스트에 NoOpLogger 추가

### Phase 3 빠른 완료
**이유**: 이미 구현된 기능 발견
**결과**:
- 예상 2일 → 실제 1시간
- 복잡도 30/100 유지
- 모든 테스트 통과

## 2025-01-07 TUI 완전 삭제

### TUI 삭제 결정
**결정**: TUI 레이어 완전 제거
**이유**:
- 헥사고날 아키텍처 심각한 위반
- God Object 안티패턴 (Model 115+ 필드)
- 비즈니스 로직이 UI 레이어에 혼재
- 테스트 불가능한 구조
**백업**: backup/legacy-tui-20250107/에 보관
**영향**:
- 코드베이스 대폭 단순화
- 테스트 가능한 구조로 전환
- Phase 4에서 깨끗한 재시작 가능

### 백그라운드 실행 모드 설계
**결정**: Job 도메인 모델 도입
**구현 방식**:
- exec.Command 자기 재실행 패턴 (Go는 fork() 불가)
- PID 파일 관리 (~/.cli-recover/jobs/)
- Job 상태 추적 (pending → running → completed/failed)
**아키텍처**:
- Domain: Job entity, JobRepository interface
- Infrastructure: FileJobRepository, ProcessExecutor
- Application: JobService

### 파일 관리 시스템 설계
**결정**: ~/.cli-recover/ 하위 통합 관리
**구조**:
```
~/.cli-recover/
├── config.yaml    # 설정 파일
├── logs/         # 로그 파일 (로테이션)
├── metadata/     # 백업 메타데이터
├── jobs/         # Job 상태 및 PID
└── tmp/          # 임시 파일
```
**Cleanup 명령**:
- 오래된 파일 자동 정리
- 타입별 선택적 정리
- Dry-run 모드 지원
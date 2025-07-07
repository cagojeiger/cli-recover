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
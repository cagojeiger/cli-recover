# 아키텍처 결정사항

## 1. Hexagonal Architecture 선택 이유

### 배경
- 현재 Model struct가 115개 이상의 필드를 가진 God Object
- kubectl 의존성이 코드 전반에 산재
- 테스트가 거의 불가능한 구조

### 결정
**Hexagonal Architecture (Ports & Adapters) 패턴 채택**

### 이유
1. **테스트 용이성**: 모든 외부 의존성을 인터페이스로 격리
2. **kubectl 독립성**: adapter 교체로 다른 컨테이너 런타임 지원 가능
3. **비즈니스 로직 보호**: 도메인 로직이 인프라 세부사항에 의존하지 않음
4. **확장성**: 새로운 adapter 추가가 용이

### 대안 검토
- **MVC**: UI 중심 패턴으로 비즈니스 로직 분리에 부적합
- **Clean Architecture**: 너무 많은 레이어로 과도한 복잡성 (복잡도 70+)
- **단순 레이어드**: 의존성 역전이 어려워 테스트 곤란

## 2. Plugin Pattern 선택 이유

### 배경
- Filesystem, MinIO, MongoDB 등 다양한 백업 타입 지원 필요
- 향후 PostgreSQL, Elasticsearch 등 추가 예정
- 각 백업 타입마다 다른 옵션과 검증 로직

### 결정
**BackupType 인터페이스 기반 Plugin Pattern**

### 이유
1. **Open/Closed 원칙**: 새 백업 타입 추가 시 기존 코드 수정 불필요
2. **독립적 개발**: 각 백업 타입을 별도로 개발/테스트 가능
3. **동적 등록**: 런타임에 플러그인 추가/제거 가능
4. **일관된 인터페이스**: 모든 백업 타입이 동일한 계약 준수

### 구현 방식
```go
type BackupType interface {
    Name() string
    BuildCommand(target Target, options Options) []string
    ValidateOptions(options Options) error
}
```

## 3. 복잡도 관리 전략

### 원칙
**오캄의 면도날**: 필요한 것만 구현

### 복잡도 평가 기준
- **Phase 1**: 25/100 (단순 분리)
- **Phase 2**: 40/100 (컴포넌트화)
- **Phase 3**: 50/100 (플러그인 시스템)
- **Phase 4**: 35/100 (최적화)

### 복잡도 제한 방법
1. **점진적 리팩토링**: 한 번에 하나씩
2. **실용적 추상화**: 3개 이상 구현체가 있을 때만 인터페이스 생성
3. **YAGNI 원칙**: 미래를 위한 과도한 설계 지양
4. **측정 가능한 개선**: 각 단계별 명확한 지표

## 4. kubectl 의존성 격리

### 문제
- 직접적인 kubectl 명령 실행이 곳곳에 산재
- 다른 런타임(docker, podman) 지원 불가
- 테스트 시 실제 클러스터 필요

### 해결
**KubernetesPort 인터페이스와 KubectlAdapter**

### 장점
1. **테스트 가능**: Mock 구현으로 단위 테스트
2. **런타임 독립성**: docker exec, podman exec 등으로 교체 가능
3. **에러 처리 일원화**: adapter에서 모든 kubectl 에러 처리
4. **성능 최적화**: 캐싱, 배치 처리 등 가능

## 5. 메모리 관리 결정

### 문제
- BackupJob.Output []string이 무제한 증가
- 대용량 백업 시 OOM 발생

### 해결
**Ring Buffer 패턴 (최대 1000줄)**

### 이유
1. **예측 가능한 메모리**: 최대 사용량이 고정됨
2. **전체 로그 보존**: 파일로 스트리밍하여 이력 유지
3. **UI 성능**: 최근 N줄만 표시하여 렌더링 최적화
4. **간단한 구현**: 복잡한 메모리 관리 불필요

## 6. 이벤트 기반 Job 관리

### 배경
- Job 상태 변경을 폴링으로 감지
- UI와 비즈니스 로직이 강하게 결합

### 결정
**Event Bus 패턴**

### 장점
1. **느슨한 결합**: Publisher와 Subscriber가 서로 모름
2. **실시간 업데이트**: 상태 변경 즉시 알림
3. **확장성**: 새로운 이벤트 핸들러 추가 용이
4. **디버깅**: 모든 이벤트를 중앙에서 추적 가능

## 7. TDD 접근 방식

### 원칙
**UI가 아닌 부분은 최대한 TDD**

### 적용 범위
1. **Domain Layer**: 100% 테스트 커버리지 목표
2. **Infrastructure Adapters**: 인터페이스 계약 테스트
3. **UI Components**: 주요 시나리오만 통합 테스트
4. **Plugins**: 각 백업 타입별 독립적 테스트

### 테스트 전략
- 단위 테스트: 빠른 피드백
- 통합 테스트: 주요 플로우 검증
- E2E 테스트: 사용자 시나리오 (최소한)
# 전략적 결정사항

## CLI-First 접근 전환 (2025-01-07)

### 배경
- TUI 중심 개발로 시작했으나 구조적 한계 발견
- God Object 안티패턴 (Model 115+ 필드)
- 현재 작동하는 CLI 백업 기능 존재
  ```bash
  ./cli-recover backup filesystem <pod> <path> --namespace <ns>
  ```
- 진행률 모니터링, 크기 추정 등 핵심 기능 이미 구현

### 결정
- **CLI 우선 개발**: 핵심 기능을 CLI로 먼저 완성
- **TUI는 나중에**: CLI를 감싸는 얇은 레이어로 구현
- **점진적 접근**: 동작하는 것부터 시작하여 개선

### 이유
1. **즉각적 가치 제공**
   - CLI는 바로 사용 가능
   - 스크립트 자동화 지원
   - CI/CD 파이프라인 통합 용이

2. **테스트 용이성**
   - CLI 명령은 자동화된 테스트 작성 쉬움
   - 입력/출력이 명확하여 검증 간단
   - Mock 없이도 통합 테스트 가능

3. **명확한 관심사 분리**
   - CLI = 비즈니스 로직
   - TUI = 프레젠테이션 레이어
   - 각 레이어 독립적 개발/테스트 가능

4. **기존 자산 활용**
   - 이미 동작하는 filesystem 백업 코드
   - kubectl exec 통합 경험
   - 진행률/ETA 계산 로직

### 원칙
```
"Make it work → Make it right → Make it pretty"
```
- **Make it work**: CLI로 모든 기능 동작
- **Make it right**: 아키텍처 정리, 플러그인 패턴
- **Make it pretty**: TUI로 사용성 개선

### 영향
- 기존 TUI 코드는 레퍼런스로 보존
- 새로운 구조로 처음부터 구축
- 하지만 검증된 로직은 재사용

## Hexagonal Architecture 유지

### 이전 결정 재확인
- Hexagonal Architecture + Plugin Pattern은 여전히 유효
- 단, CLI 레이어부터 시작하여 구축

### 적용 방식
```
CLI Commands
    ↓
Application Services (Use Cases)
    ↓
Domain (BackupProvider Interface)
    ↓
Infrastructure (kubectl, filesystem)
```

### 플러그인 시스템
- BackupProvider 인터페이스로 확장성 확보
- filesystem, minio, mongodb 등 provider 구현
- 새로운 백업 타입 추가 용이

## 개발 우선순위

### Phase 1: CLI 핵심 (1주)
1. 명령 체계 정리
2. 3가지 백업 타입 구현
3. 공통 기능 (진행률, 에러 처리)

### Phase 2: 아키텍처 (1주)
1. 도메인 레이어 분리
2. 플러그인 패턴 적용
3. 테스트 커버리지 확보

### Phase 3: CLI 고도화 (2주)
1. restore 명령
2. list/status 명령
3. 설정 관리

### Phase 4: TUI 통합 (2주)
1. CLI 명령 래핑
2. 인터랙티브 UI
3. 실시간 모니터링

## 위험 관리

### 식별된 위험
- 기존 TUI 사용자 혼란 → 명확한 마이그레이션 가이드 제공
- CLI 복잡도 증가 → 직관적인 명령 체계 설계
- 중복 구현 → 공통 로직 최대한 재사용

### 완화 전략
- 단계별 릴리즈로 점진적 전환
- 명확한 문서화
- 기존 기능 호환성 유지
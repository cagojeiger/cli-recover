# v2.0 Fresh Start Checkpoint

## 날짜: 2025-01-11
## 상태: 🚀 Zero to Hero Implementation
## 브랜치: feature/v2-fresh-start
## 결정: 완전히 새로운 시작

## 왜 다시 시작하는가?

### v1의 문제점
1. **과도한 복잡성**
   - 3계층 아키텍처가 단순한 도구에 과함
   - Provider 인터페이스가 각 provider의 특성을 제한
   - 불필요한 추상화로 코드 이해도 저하

2. **리팩토링 지옥**
   - backup.go, backup_old.go, backup_v2.go 혼재
   - 어떤 코드가 production인지 불명확
   - Git history 의미 상실

3. **잘못된 우선순위**
   - 재사용성을 위해 격리성 희생
   - 미래를 위한 설계로 현재 복잡도 증가
   - 실제 필요보다 많은 기능 구현

### v2의 접근법
> "Start from scratch with lessons learned"

1. **격리 우선 설계**
   - Provider 간 의존성 없음
   - 코드 중복 허용
   - 명확한 경계 유지

2. **단순함 추구**
   - 최소한의 구조로 시작
   - 필요할 때만 확장
   - 추상화 금지 (v2.0에서)

3. **TDD 철저히 적용**
   - 테스트 없는 코드 없음
   - Red → Green → Refactor
   - 90% 이상 커버리지 목표

## 무엇을 버리는가?

### 코드 (100% 폐기)
- ❌ internal/ 디렉토리 전체
- ❌ cmd/ 디렉토리 전체
- ❌ experimental/ 디렉토리
- ❌ 모든 Go 코드

### 아키텍처
- ❌ Domain/Application/Infrastructure 3계층
- ❌ Provider 인터페이스
- ❌ Registry 패턴
- ❌ 플러그인 시스템

### 복잡한 기능
- ❌ 다중 Provider 지원
- ❌ TUI 인터페이스
- ❌ 설정 파일 시스템
- ❌ 고급 메타데이터

## 무엇을 가져가는가?

### 문서와 지식
- ✅ CLAUDE.md (작업 규칙)
- ✅ .context/* (프로젝트 컨텍스트)
- ✅ .planning/* (계획 문서)
- ✅ .checkpoint/* (이정표)
- ✅ .memory/* (학습 내용)

### 핵심 교훈
1. **기술적 교훈**
   - 원자적 파일 쓰기 (temp + rename)
   - 스트리밍 처리 중요성
   - 진행률 보고 필수성
   - kubectl exec 사용법

2. **프로세스 교훈**
   - TDD의 실제 효과
   - 구조/행동 변경 분리
   - 계획 우선 구현
   - 복잡도 관리 방법

3. **철학적 교훈**
   - Isolation > Reusability
   - Duplication > Wrong abstraction
   - Simple > Clever
   - Working > Perfect

## v2.0 목표

### 단기 목표 (2주)
- 단일 Provider (filesystem) 완성
- CLI 인터페이스만 제공
- 테스트 커버리지 90%+
- 복잡도 < 20/100

### 코드 구조 (예상)
```
cli-recover/
├── main.go           # 진입점
├── backup.go         # 백업 로직
├── backup_test.go    # 백업 테스트
├── restore.go        # 복원 로직
├── restore_test.go   # 복원 테스트
├── progress.go       # 진행률 보고
├── progress_test.go  # 진행률 테스트
└── Makefile         # 빌드 스크립트
```

### 성공 기준
- 코드 라인 < 1000
- 파일 개수 < 10
- 외부 의존성 = 0
- 빌드 시간 < 10초
- 바이너리 크기 < 30MB

## 위험과 대응

### 위험 요소
1. **단순함이 제한이 될 수 있음**
   - 대응: v2.1에서 필요시 확장

2. **코드 중복 증가**
   - 대응: 격리의 이점이 더 큼

3. **확장성 부족**
   - 대응: YAGNI (You Ain't Gonna Need It)

### 실패 조건
- 복잡도 > 30/100
- 테스트 커버리지 < 80%
- 사용자 만족도 저하

## 다음 단계

### 즉시 실행
1. go.mod 초기화
2. main.go 생성 (version 명령만)
3. Makefile 작성
4. 첫 커밋

### Phase 1 목표
- 최소 실행 가능한 바이너리
- CI/CD 파이프라인
- 기본 문서

## 마무리 생각

v1은 실패가 아니라 학습이었습니다. 우리는 이제 무엇이 중요하고 무엇이 불필요한지 알고 있습니다. v2는 그 지식을 바탕으로 한 의도적인 단순함입니다.

> "Perfection is achieved not when there is nothing more to add, but when there is nothing left to take away."  
> — Antoine de Saint-Exupéry
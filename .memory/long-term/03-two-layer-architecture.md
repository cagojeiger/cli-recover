# 2계층 아키텍처 결정

## 날짜
- 2025-01-08

## 결정 배경
- Phase 3.9에서 3계층 아키텍처를 2계층으로 단순화하기로 결정

## 문제점
### Application 레이어의 역할 부재
- 단순히 CLI에서 Domain으로 전달만 수행
- 비즈니스 로직이 없음
- 불필요한 중간 단계

### 과도한 추상화
- Registry 패턴이 provider 1개만 관리
- 미래를 위한 과도한 준비
- YAGNI 원칙 위반

### 코드 복잡도
- 함수 호출 체인이 너무 김
- 디버깅 어려움
- 새 개발자 온보딩 시간 증가

## 해결책: 2계층 아키텍처

### 구조
- CMD Layer: 진입점 + 조율
- Domain Layer: 비즈니스 로직
- Infrastructure Layer: 외부 시스템
- 양방향 통신: Domain ↔ Infrastructure

### 장점
#### 명확한 책임 분리
- Domain: What & Why
- Infrastructure: How
- CMD: 사용자 인터페이스

#### 직접적인 호출
- 불필요한 중간 단계 제거
- 코드 추적 용이
- 디버깅 간소화

#### Go 생태계 표준
- 대부분의 Go 프로젝트가 2계층 사용
- 커뮤니티 베스트 프랙티스
- 친숙한 구조

## 구현 상세

### Before (3계층)
- cmd.Execute() → adapter.ExecuteBackup()
- → registry.Get("filesystem")
- → provider.Execute()
- → kubernetes.Exec()

### After (2계층)
- cmd.Execute()
- → filesystem.NewBackupProvider().Execute()
- → kubernetes.Exec()

## 교훈
### 과도한 미래 대비는 독
- 필요할 때 추가하는 것이 쉬움
- 불필요한 복잡도 제거가 어려움

### 단순함이 최선
- 이해하기 쉬운 코드
- 유지보수 용이
- 버그 감소

### 실용적 접근
- 이론보다 실제 사용
- 패턴을 위한 패턴 지양
- 문제 해결에 집중

## 측정 지표
- 복잡도: 75 → 30 (목표)
- 파일 수: -40%
- 코드 라인: -35%
- 함수 호출 깊이: 5 → 3
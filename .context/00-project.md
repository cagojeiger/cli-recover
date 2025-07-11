# CLI-Recover Project v2.0

## 프로젝트 개요
- **이름**: cli-recover
- **목적**: Kubernetes 통합 백업/복원 도구
- **설명**: Kubernetes 파드 파일시스템 백업/복원 (격리 우선 설계)
- **주 사용자**: Kubernetes 클러스터 관리 DevOps 엔지니어
- **현재 버전**: v2.0.0-alpha (fresh start)

## 현재 상태 (2025-01-11)
- **단계**: Fresh Start - Zero to Hero Implementation
- **브랜치**: feature/v2-fresh-start
- **철학**: "Isolation > Reusability"
- **접근법**: 완전히 새로운 구현, 배운 교훈만 적용

## v1에서 배운 핵심 교훈
- 과도한 추상화는 복잡성만 증가시킴
- Provider 격리가 재사용성보다 중요
- CLI-first 접근이 더 실용적
- TDD가 안전한 리팩토링 보장
- 명확한 경계가 혼란 방지

## 목표 및 제약사항

### 주요 목표
- Kubernetes 워크로드의 안정적인 백업/복원 제공
- 단일 Provider (filesystem)에 집중
- CLI 인터페이스만 제공 (TUI는 나중에)
- 100% 격리된 Provider 구현
- 단순하고 이해하기 쉬운 코드베이스

### 기술적 제약
- 표준 Kubernetes API만 사용 (kubectl exec)
- 클러스터 레벨 권한 불필요
- 외부 도구 의존성 최소화 (kubectl만)
- 크로스 플랫폼 호환성 (Linux, macOS, Windows)
- 단일 바이너리 배포

### 설계 제약
- No shared interfaces initially
- Direct implementation without abstractions
- Copy code rather than share
- TDD from the beginning
- Progress reporting built-in

## 성공 지표
- 테스트 커버리지 >90%
- 바이너리 크기 <30MB
- 백업/복원 작업 성공률 99%+
- 코드 복잡도 <20/100
- Zero external dependencies (except kubectl)

## 비목표
- Multiple providers in v2.0
- TUI interface in v2.0
- Shared abstractions
- Plugin architecture
- Cloud storage integration
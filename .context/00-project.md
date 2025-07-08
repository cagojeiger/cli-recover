# CLI-Recover Project

## 프로젝트 개요
- **이름**: cli-recover
- **목적**: Kubernetes 통합 백업/복원 도구
- **설명**: Kubernetes 파드 파일시스템, 데이터베이스, 오브젝트 스토리지 백업/복원
- **주 사용자**: Kubernetes 클러스터 관리 DevOps 엔지니어
- **현재 버전**: v0.2.0 (dev)

## 현재 상태 (2025-01-08)
- **단계**: Phase 3.9 코드 단순화 진행중
- **브랜치**: feature/simplification
- **진행률**: Phase 1-3 완료, Phase 3.9 진행중
- **완료 항목**:
  - Filesystem 백업/복원 provider ✅
  - 메타데이터 저장 시스템 ✅
  - List 명령 (백업 목록) ✅
  - Provider registry 패턴 ✅
  - 로그 파일 시스템 ✅
- **진행중**:
  - Phase 3.9: 아키텍처 단순화 (Occam's Razor)
  - 3계층 → 2계층 구조 변경
  - 미사용 코드 제거
  - 복잡도 75 → 30 목표
- **계획됨**:
  - Phase 4: TUI 구현 (tview)

## 목표 및 제약사항

### 주요 목표
- Kubernetes 워크로드의 안정적인 백업/복원 제공
- 핵심 백업 타입 지원 (filesystem 우선)
- CLI와 TUI 인터페이스 모두 제공
- 높은 테스트 커버리지 유지 (비즈니스 로직 >80%)
- 단순하고 이해하기 쉬운 코드베이스 (Occam's Razor)

### 기술적 제약
- 표준 Kubernetes API만 사용
- 클러스터 레벨 권한 불필요
- 외부 도구 의존성 최소화
- 크로스 플랫폼 호환성 (Linux, macOS, Windows)
- 단일 바이너리 배포

### 설계 제약
- CLI-first 접근 (TUI는 래퍼)
- 2계층 아키텍처 (Domain/Infrastructure)
- 단순성 우선 설계 원칙
- 동기식 실행 모델
- 로컬 메타데이터 저장

## 성공 지표
- 테스트 커버리지 >80% (TUI 제외)
- 바이너리 크기 <50MB
- 백업/복원 작업 성공률 99%+
- 명확한 에러 메시지와 복구 경로
- 모든 공개 API 문서화

## 비목표
- 실시간 연속 백업
- 다중 클러스터 관리
- GUI 인터페이스
- 클라우드 스토리지 직접 통합
- 백업 스케줄링 (Kubernetes CronJob 사용)
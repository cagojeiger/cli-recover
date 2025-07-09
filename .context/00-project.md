# CLI-Recover Project

## 프로젝트 개요
- **이름**: cli-recover
- **목적**: Kubernetes 통합 백업/복원 도구
- **설명**: Kubernetes 파드 파일시스템, 데이터베이스, 오브젝트 스토리지 백업/복원
- **주 사용자**: Kubernetes 클러스터 관리 DevOps 엔지니어
- **현재 버전**: v0.2.0 (dev)

## 현재 상태 (2025-01-09)
- **단계**: Phase 3.12 CLI 사용성 개선 (구현 준비) ⏳
- **브랜치**: feature/tui-backup
- **완료된 Phase**:
  - Phase 1-3: CLI 핵심 기능 ✅
  - Phase 3.9: 아키텍처 단순화 (2계층) ✅
  - Phase 3.10: 백업 무결성 ✅
  - Phase 3.11: 진행률 보고 시스템 ✅
  - Phase 3.12 문서: CLI 디자인 가이드 완료 ✅
  - Phase 4: TUI 구현 (tview) ✅
- **계획된 Phase**:
  - Phase 3.12: CLI 사용성 개선 (플래그 레지스트리, 충돌 해결)
  - Phase 3.13: 도구 자동 다운로드
  - Phase 5: Provider 확장 (MinIO/MongoDB)

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
- 테스트 커버리지 >58% (현재 달성)
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
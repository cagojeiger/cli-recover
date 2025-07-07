# 현재 작업: Phase 3 - 백그라운드 모드 구현

## 작업 시작
- 날짜: 2025-01-07
- 목표: Phase 3 백그라운드 실행 모드 및 파일 관리 시스템
- 이전 작업: 헥사고날 아키텍처 정리 완료

## 완료된 작업
### 1. 헥사고날 아키텍처 정리 ✅
- TUI 완전 삭제 (backup/legacy-tui-20250107/)
- 중복 디렉토리 제거
- 패키지 올바른 레이어로 이동
- _new suffix 파일명 정리

### 2. 로거 시스템 확인 ✅
- 이미 완전히 구현되어 있음
- CLI 플래그 통합 완료
- 테스트 커버리지 유지

### 3. 테스트 상태 ✅
- 모든 테스트 통과
- 커버리지: 53.0% (TUI 제거로 감소)
- 실제 비즈니스 로직 커버리지는 유지

## 진행 예정
### 1. Job 도메인 모델
- [ ] internal/domain/job/ 패키지 생성
- [ ] Job 엔티티 (ID, PID, Status, StartTime, EndTime)
- [ ] JobRepository 인터페이스
- [ ] JobStatus enum (pending, running, completed, failed)

### 2. 백그라운드 실행 인프라
- [ ] internal/infrastructure/process/ 패키지
- [ ] ProcessExecutor 구현
- [ ] PID 파일 관리 (~/.cli-recover/jobs/)
- [ ] 시그널 핸들링

### 3. JobService 애플리케이션 레이어
- [ ] internal/application/service/ 패키지
- [ ] JobService 구현
- [ ] FileJobRepository 구현

### 4. CLI 통합
- [ ] backup 명령에 --background 플래그
- [ ] status 명령 구현
- [ ] --watch 옵션

### 5. 파일 관리 시스템
- [ ] cleanup 명령
- [ ] 보관 정책 설정
- [ ] Dry-run 모드

## 목표
- 백그라운드 실행 가능
- Job 상태 추적
- 파일 자동 정리
- 테스트 커버리지 80%
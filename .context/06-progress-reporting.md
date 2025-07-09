# 진행률 보고 원칙

## 프로젝트 내 위치
- 모든 장시간 I/O 작업의 필수 요소
- 사용자 경험의 핵심 구성 요소
- CLI, 로그, TUI 전체에 걸친 통합 시스템

## 핵심 원칙

### 3초 규칙
**3초 이상 걸리는 모든 작업은 진행률을 보고해야 한다.**

적용 대상:
- 백업/복원 작업
- 네트워크 다운로드
- 대용량 파일 처리
- 외부 명령 실행

### 다중 환경 지원
하나의 구현으로 모든 환경 대응:
```
터미널 → \r을 사용한 실시간 업데이트
CI/CD → 주기적 로그 라인 출력
로그 파일 → 구조화된 진행률 데이터
TUI → 채널을 통한 실시간 전송
```

### 복잡도 관리
- 추가 라이브러리 없음
- 표준 라이브러리만 사용
- 복잡도 증가: +5/100
- Occam's Razor 원칙 준수

## 표준 인터페이스

```go
type ProgressReporter interface {
    Update(current, total int64, message string)
    Complete(message string)
    Error(err error, message string)
}
```

## 구현 요구사항

### 필수 기능
1. **환경 감지**: `term.IsTerminal()` 사용
2. **주기적 로깅**: 10초 간격 기본값
3. **원자적 업데이트**: 진행 상태 일관성
4. **에러 시 정리**: 진행률 표시 정리

### 성능 요구사항
- 업데이트 간격: 200ms ~ 500ms (터미널)
- 로그 간격: 10초 (설정 가능)
- CPU 오버헤드: < 1%
- 메모리: 상수 사용량 (스트리밍)

## 패턴 활용

### io.TeeReader 패턴
```go
pw := &progressWriter{reporter: reporter}
reader := io.TeeReader(source, pw)
io.Copy(dest, reader)
```

### 크기 추정
- 가능한 경우 항상 전체 크기 제공
- 불가능한 경우 처리된 양만 표시
- 예상 시간(ETA)은 신뢰할 수 있을 때만

## 아키텍처 통합

### Domain 레이어
- ProgressReporter 인터페이스 정의
- Provider들이 진행률 보고 책임

### Infrastructure 레이어
- 구체적인 Reporter 구현
- 터미널/로그 출력 처리

### Presentation 레이어
- CLI: 터미널 진행바
- TUI: 그래픽 진행 표시
- API: JSON 진행률 스트림

## 테스트 전략

### 단위 테스트
- Mock Writer로 출력 검증
- 시간 기반 로직 Mock
- 환경 감지 Override

### 통합 테스트
- 실제 파일 작업 진행률
- 네트워크 다운로드 시뮬레이션
- 다양한 터미널 환경

## 사용자 경험 가이드라인

### DO ✅
- 명확한 진행 상태 표시
- 정확한 크기와 시간 정보
- 인터럽트 시 깔끔한 정리
- 에러 시 명확한 메시지

### DON'T ❌
- 너무 빠른 업데이트 (깜빡임)
- 부정확한 예측 시간
- 터미널 잔상 남기기
- 로그 스팸

## Phase 별 적용

### Phase 3.10 (백업 무결성)
- 백업 파일 쓰기 진행률
- 임시 파일 사용 중 표시
- 체크섬 계산 진행률

### Phase 3.13 (도구 다운로드)
- HTTP Content-Length 활용
- 다운로드 속도 표시
- 재시도 시 진행률 리셋

### Phase 4 (TUI)
- 별도 진행률 채널
- 그래픽 진행바
- 다중 작업 동시 표시

## 확장 가능성

### 향후 고려사항
- 병렬 작업 진행률
- 하위 작업 진행률
- 진행률 히스토리
- 성능 메트릭 수집

### 플러그인 지원
- Custom Reporter 구현 가능
- 진행률 포맷 커스터마이징
- 외부 모니터링 통합

## 구현 체크리스트

- [ ] ProgressReporter 인터페이스 정의
- [ ] 표준 Reporter 구현
- [ ] 환경 감지 로직
- [ ] io.TeeReader 통합
- [ ] 테스트 작성
- [ ] 문서화

## 참고 자료

- [구현 가이드](../docs/progress-reporting/01-implementation-guide.md)
- [예제 모음](../docs/progress-reporting/02-examples.md)
- Go 표준 라이브러리: io, time, os
- golang.org/x/term 패키지
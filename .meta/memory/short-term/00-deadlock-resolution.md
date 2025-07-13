# 데드락 해결 과정

## 문제 발생
- 날짜: 2025-01-13
- 증상: hello-world.yaml 실행 시 "all goroutines are asleep - deadlock!"
- 원인: io.Pipe의 동기식 특성 (버퍼 없음)

## 시도한 해결책들

### 1차 시도: 동시 실행
- 모든 스텝을 고루틴으로 실행
- 결과: 단순 케이스는 해결, 복잡한 케이스는 여전히 데드락
- 문제점: 타이밍 의존성, 복잡한 동기화

### 2차 시도: MultiWriter 제거
- 로깅을 위한 MultiWriter가 추가 블로킹 유발
- 결과: 일부 개선되었으나 근본 해결 안됨

### 3차 시도: 하이브리드 접근
- Unix pipe와 Go streams 선택적 사용
- 결과: 성공! 데드락 완전 해결

## 핵심 깨달음

### io.Pipe vs Unix Pipe
- io.Pipe: 버퍼 없음, 동기식, Go 프로세스 내부
- Unix Pipe: 64KB 버퍼, 비동기식, 커널 관리

### Unix Philosophy
- "Do one thing and do it well"
- 복잡한 파이프 메커니즘을 재발명할 필요 없음
- 이미 검증된 Unix 도구 활용

## 최종 해결책

### 하이브리드 실행 전략
```go
if isSimpleLinear(pipeline) {
    return executeAsUnixPipe()  // 대부분의 케이스
}
return executeWithGoStreams()   // 진행률 필요시
```

### 로깅 전략
- 추적성은 필수 요구사항
- Unix pipe에서도 tee로 로깅 보장
- 로그 디렉토리에 각 스텝 출력 저장

## 남은 과제
- [ ] ShellPipeExecutor 구현
- [ ] 진행률 표시를 위한 ProgressReader
- [ ] 복잡한 파이프라인 최적화

## 교훈
- 단순함이 최선 (Occam's Razor)
- 기존 도구의 장점 활용
- 과도한 추상화 피하기
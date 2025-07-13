# 학습 내용

## io.Pipe vs Unix Pipe

### io.Pipe의 한계
- 버퍼가 없어 동기적 동작
- Writer가 쓰면 Reader가 즉시 읽어야 함
- 블로킹 명령어(grep, sed)와 데드락 발생

### Unix Pipe의 장점
- 커널 버퍼(64KB) 제공
- 비동기적 동작
- 프로세스 간 자연스러운 통신

### 해결책
- 단순한 경우 Unix pipe 직접 사용
- 복잡한 경우만 Go stream 사용
- 전략 패턴으로 유연하게 대응

## 동시성 관리

### 잘못된 접근
- 모든 Step을 동시에 시작
- 입력이 준비되지 않은 상태에서 실행
- 데드락 불가피

### 올바른 접근
- 의존성 기반 그룹화
- 각 그룹을 순차적으로 실행
- 그룹 내에서만 동시 실행

### 구현 패턴
```go
// 의존성 분석
producedOutputs := map[string]bool{}
for _, step := range pipeline.Steps {
    if canExecute(step, producedOutputs) {
        // 실행 가능한 그룹에 추가
    }
}
```

## 테스트 전략

### 유닛 테스트
- 각 전략별 독립적 테스트
- 모의 파이프라인으로 검증
- 에러 케이스 충분히 커버

### 통합 테스트
- 실제 Unix 명령어 실행
- 파일 생성/읽기 검증
- 로깅 출력 확인

## 디버깅 팁

### 데드락 추적
- goroutine 덤프 확인
- 어느 지점에서 블록되는지 파악
- io.Copy의 EOF 대기 주의

### 성능 분석
- Unix pipe가 대부분 더 빠름
- 메모리 사용량도 적음
- 복잡도가 낮을수록 Unix pipe 유리
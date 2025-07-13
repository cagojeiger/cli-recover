# 실행 전략 결정: 하이브리드 접근법

## 결정 일자: 2025-01-13

### 배경
- MVP 구현 중 io.Pipe 사용 시 데드락 발생
- 순차 실행에서 writer가 reader를 기다리며 블록
- 동시 실행으로 해결 시도했으나 복잡한 파이프라인에서 여전히 문제

### 문제 분석

#### io.Pipe의 한계
- 버퍼가 없는 동기식 파이프
- Write()는 Read()가 일어날 때까지 블록
- Unix pipe와 달리 커널 버퍼가 없음
- 타이밍에 민감한 "빡빡한 결합"

#### Unix Pipe의 장점
- 커널이 관리하는 64KB 버퍼
- 프로세스 간 느슨한 결합
- 자연스러운 병렬 실행
- 검증된 안정성

### 고려한 옵션들

#### 1. 순수 Go 구현 (io.Pipe + 동시성)
```go
// 모든 스텝을 고루틴으로 실행
for _, step := range pipeline.Steps {
    go executeStep(step)
}
```
- 장점: 세밀한 제어 가능
- 단점: 데드락 위험, 복잡한 동기화

#### 2. 순수 Unix Pipe
```bash
echo "data" | gzip | tee output.gz
```
- 장점: 단순함, 안정성, 성능
- 단점: 로깅/진행률 표시 어려움

#### 3. 하이브리드 접근법 (선택)
- 단순 파이프라인: Unix pipe 사용
- 복잡한 경우: Go streams 사용
- 로깅은 항상 활성화

### 최종 결정

**하이브리드 실행 전략**을 채택함.

#### 핵심 원칙
1. **단순함 우선**: 가능하면 Unix pipe 사용
2. **필수 로깅**: 모든 실행은 추적 가능해야 함
3. **점진적 복잡도**: 필요한 경우만 Go streams

#### 실행 모드 결정 로직
```go
if isSimpleLinear(pipeline) && !requiresProgress(pipeline) {
    return ShellPipeExecution  // Unix pipe
}
return GoStreamExecution  // 세밀한 제어 필요
```

### 로깅 전략

#### 필수 요구사항
- 프로젝트 비전: "모든 CLI 작업을 추적 가능하게"
- 각 스텝의 입출력 기록
- 실행 시간, 종료 코드 저장

#### Unix Pipe에서의 로깅
```bash
# tee를 사용한 로깅
tar cf - /data 2>step1.err | tee step1.out | gzip
```

### 진행률 표시 전략

#### 제약사항
- Unix 기본 도구만 사용 (pv 등 외부 도구 배제)
- 메모리 효율성 유지

#### 해결책
- 진행률이 필요한 경우 Go streams 사용
- ProgressReader로 바이트 카운트
- 100ms 간격으로 업데이트

### 복잡도 평가

- 기존 io.Pipe 전용: 20/100
- 하이브리드 접근법: 35/100 ✅
- 여전히 Occam's Razor 준수

### 영향

1. **아키텍처 변경**
   - ExecutePipeline에 전략 패턴 적용
   - ShellPipeExecutor 추가
   - StreamManager 역할 축소

2. **성능 향상**
   - Unix pipe의 커널 최적화 활용
   - 불필요한 Go 루틴 제거

3. **안정성 증가**
   - 데드락 위험 제거
   - 검증된 Unix 메커니즘 활용

### 향후 계획

1. **Phase 1**: 기본 하이브리드 구현
2. **Phase 2**: 모니터링 최적화
3. **Phase 3**: 성능 튜닝
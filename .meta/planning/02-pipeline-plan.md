# Pipeline (v0.2) 구현 계획

## 목표
- **기간**: 1주 (MVP 완료 후)
- **목적**: 실용적인 파이프라인 기능 추가
- **범위**: 변수, 병렬 실행, 에러 처리

## 추가 기능

### 1. 파라미터 시스템

#### YAML 정의
```yaml
name: parameterized-pipeline
params:
  source:
    type: string
    default: "/data"
    description: "Source directory"
  level:
    type: int
    default: 9
    min: 1
    max: 9

steps:
  - name: compress
    run: tar czf - {{.source}} | gzip -{{.level}}
    output: archive
```

#### 구현
```go
type Parameter struct {
    Type        string      `yaml:"type"`
    Default     interface{} `yaml:"default"`
    Required    bool        `yaml:"required"`
    Description string      `yaml:"description"`
}

func (p *Pipeline) BindParams(values map[string]string) error {
    // 템플릿 치환
    // 타입 검증
    // 기본값 적용
}
```

### 2. 고루틴 기반 병렬 실행

#### DAG 분석
```go
type DAG struct {
    nodes map[string]*Node
    edges map[string][]string
}

func BuildDAG(steps []Step) *DAG {
    // 입출력 관계 분석
    // 의존성 그래프 구성
    // 실행 순서 결정
}
```

#### 병렬 실행
```go
func (e *Executor) ExecuteParallel(p *Pipeline) error {
    dag := BuildDAG(p.Steps)
    levels := dag.TopologicalSort()
    
    for _, level := range levels {
        var wg sync.WaitGroup
        for _, step := range level {
            wg.Add(1)
            go func(s Step) {
                defer wg.Done()
                e.executeStep(s)
            }(step)
        }
        wg.Wait()
    }
}
```

### 3. 스트림 분기 (Hidden tee 전략)

#### YAML 정의 (사용자 관점 - 단순함 유지)
```yaml
steps:
  - name: source
    run: tar cf - /data
    output: archive
    
  - name: compress
    run: gzip -9
    input: archive
    output: file:backup.gz
    
  - name: checksum
    run: sha256sum
    input: archive  # 같은 archive 재사용
    output: file:backup.sha256
```

#### 핵심 전략: 자동 tee 삽입
```go
// 내부적으로 생성되는 명령어
// tar cf - /data | tee >(sha256sum > backup.sha256) | gzip -9 > backup.gz
```

#### 구현
```go
type StreamAnalyzer struct {
    usage map[string]*StreamUsage
}

type StreamUsage struct {
    producer  string   // output을 생성하는 step
    consumers []string // input으로 사용하는 steps
}

func (a *StreamAnalyzer) Analyze(p *Pipeline) {
    // 각 output이 몇 번 사용되는지 분석
    // 분기점 자동 감지
}

func BuildSmartCommand(p *Pipeline, usage map[string]*StreamUsage) string {
    // 분기가 필요한 곳에 자동으로 tee 삽입
    // 사용자는 복잡한 Unix 명령을 몰라도 됨
    
    for _, step := range p.Steps {
        if len(usage[step.Output].consumers) > 1 {
            // tee 명령 자동 생성
            cmd = buildTeeCommand(step, usage)
        } else {
            // 일반 파이프
            cmd = step.Run
        }
    }
}
```

#### 장점
- **사용자 단순성**: YAML 변경 없음 (복잡도 20/100)
- **내부 스마트함**: 자동 최적화 (복잡도 40/100)
- **Unix 네이티브**: tee는 POSIX 표준
- **메모리 효율**: 스트리밍 처리

### 4. 에러 처리

#### 전략 정의
```yaml
on_error: stop|continue|retry

steps:
  - name: risky-step
    run: curl https://api.example.com
    on_error: retry
    retry:
      count: 3
      delay: 5s
```

#### 구현
```go
type ErrorStrategy string

const (
    ErrorStop     ErrorStrategy = "stop"
    ErrorContinue ErrorStrategy = "continue"
    ErrorRetry    ErrorStrategy = "retry"
)

func (e *Executor) handleError(step Step, err error) error {
    switch step.OnError {
    case ErrorRetry:
        return e.retryStep(step)
    case ErrorContinue:
        e.logger.Warn("Step failed, continuing", err)
        return nil
    default:
        return err
    }
}
```

### 5. 진행률 표시

#### 통합 방법
```yaml
steps:
  - name: download
    run: curl -L {{.url}}
    progress: true
    
  - name: process
    run: pv -pterb | gzip -9
    input: download
```

#### 구현
```go
func (e *Executor) wrapWithProgress(cmd *exec.Cmd, step Step) {
    if step.Progress {
        // pv 명령어 주입
        // 또는 자체 진행률 추적
    }
}
```

## 구조 확장

```
internal/
├── pipeline/
│   ├── types.go
│   ├── parser.go
│   └── validator.go    # 새로 추가
├── executor/
│   ├── executor.go
│   ├── parallel.go     # 병렬 실행
│   └── stream.go       # 스트림 관리
├── template/
│   └── engine.go       # 변수 치환
└── dag/
    └── analyzer.go     # DAG 분석
```

## 구현 일정

### Week 1

#### Day 1-2: 파라미터 시스템
- [ ] Parameter 타입 정의
- [ ] 템플릿 엔진 구현
- [ ] CLI 플래그 파싱
- [ ] 검증 로직

#### Day 3-4: DAG 및 병렬 실행
- [ ] DAG 구조 설계
- [ ] 의존성 분석
- [ ] 병렬 실행 엔진
- [ ] 동기화 처리

#### Day 5-6: 스트림 분기
- [ ] StreamManager 구현
- [ ] 다중 reader 지원
- [ ] 메모리/파일 백엔드
- [ ] 리소스 관리

#### Day 7: 에러 처리 및 테스트
- [ ] 에러 전략 구현
- [ ] 재시도 로직
- [ ] 통합 테스트
- [ ] 성능 테스트

## 테스트 시나리오

### 1. 병렬 백업
```yaml
name: parallel-backup
steps:
  - name: backup-app
    run: tar czf app.tar.gz /app
    
  - name: backup-data
    run: tar czf data.tar.gz /data
    
  - name: backup-logs
    run: tar czf logs.tar.gz /logs
    
  - name: create-manifest
    run: |
      echo "app.tar.gz" > manifest.txt
      echo "data.tar.gz" >> manifest.txt
      echo "logs.tar.gz" >> manifest.txt
    depends_on: [backup-app, backup-data, backup-logs]
```

### 2. 스트림 분기
```yaml
name: process-with-verification
steps:
  - name: download
    run: curl -L {{.url}}
    output: data
    
  - name: save
    run: cat > {{.filename}}
    input: data
    
  - name: verify-md5
    run: md5sum
    input: data
    output: var:md5
    
  - name: verify-sha256
    run: sha256sum
    input: data
    output: var:sha256
    
  - name: report
    run: |
      echo "MD5: {{.md5}}"
      echo "SHA256: {{.sha256}}"
```

## 성능 목표

### 메모리 사용
- 스트림 버퍼: 64KB
- 최대 동시 고루틴: 10
- 전체 메모리: < 100MB

### 처리 속도
- 파이프라인 오버헤드: < 10ms
- 스트림 분기 오버헤드: < 5%
- DAG 분석: < 1ms

## 위험 관리

### 1. 고루틴 누수
- 대응: context 사용
- 모니터링: runtime.NumGoroutine()

### 2. 데드락
- 대응: 타임아웃 설정
- 테스트: 순환 의존성

### 3. 메모리 폭발
- 대응: 스트림 크기 제한
- 백업: 디스크 스필

## v0.3 준비

### 인터페이스 설계
- Builder를 위한 API
- 파이프라인 검증 API
- 실행 상태 조회 API

### 메타데이터 수집
- 실행 통계
- 자주 사용되는 패턴
- 에러 패턴
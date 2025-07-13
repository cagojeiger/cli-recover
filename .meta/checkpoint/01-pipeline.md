# Checkpoint: Pipeline (v0.2)

## 상태: 계획됨

### 목표
- 실용적인 파이프라인 기능 구현
- 변수 시스템과 병렬 실행
- 스트림 분기 지원

### 완료 기준
- [ ] 파라미터 정의 및 변수 치환
- [ ] DAG 기반 병렬 실행
- [ ] 스트림 분기 (하나의 출력을 여러 입력으로)
- [ ] 에러 처리 전략
- [ ] 고루틴 기반 동시 실행

### 핵심 기능

#### 1. 파라미터 시스템
```yaml
name: backup-pipeline
params:
  source:
    type: path
    default: /data
  level:
    type: int
    default: 9

steps:
  - name: compress
    run: tar czf - {{.source}} | gzip -{{.level}}
    output: archive
```

#### 2. 병렬 실행
```yaml
steps:
  # 이 세 단계는 동시 실행 가능
  - name: backup-app
    run: tar czf app.tar.gz /app
    
  - name: backup-data
    run: tar czf data.tar.gz /data
    
  - name: backup-logs
    run: tar czf logs.tar.gz /logs
    
  # 이 단계는 위 세 단계 완료 후 실행
  - name: create-manifest
    run: ls -la *.tar.gz > manifest.txt
    depends_on: [backup-app, backup-data, backup-logs]
```

#### 3. 스트림 분기
```yaml
steps:
  - name: create-archive
    run: tar cf - /data
    output: archive-stream
    
  - name: compress
    run: gzip -9 > backup.tar.gz
    input: archive-stream
    
  - name: checksum
    run: sha256sum > backup.sha256
    input: archive-stream  # 같은 스트림 재사용
```

#### 4. 에러 처리
```yaml
on_error: stop  # 전역 설정

steps:
  - name: download
    run: curl -f {{.url}}
    output: data
    on_error: retry
    retry:
      count: 3
      delay: 5s
```

### 아키텍처 변경

#### DAG 분석기
```go
type DAG struct {
    nodes map[string]*Node
    edges map[string][]string
}

func (d *DAG) TopologicalSort() [][]string
func (d *DAG) FindCycles() [][]string
```

#### 스트림 매니저
```go
type StreamManager struct {
    streams map[string]*ManagedStream
}

type ManagedStream struct {
    source  io.ReadCloser
    readers []io.Reader
    buffer  *bytes.Buffer
}
```

### 성능 목표
- DAG 분석: < 10ms (100 steps)
- 병렬 실행 오버헤드: < 5%
- 스트림 분기 오버헤드: < 10%
- 메모리 사용: < 100MB

### 새로운 의존성
- golang.org/x/sync/errgroup (병렬 실행)
- github.com/dustin/go-humanize (진행률)

### 테스트 시나리오
1. 10개 step 병렬 실행
2. 3-way 스트림 분기
3. 중첩된 변수 치환
4. 순환 의존성 감지

### 다음 단계
- v0.3: Builder UI
- v1.0: 안정화 및 최적화
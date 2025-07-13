# MVP (v0.1) 구현 계획

## 목표
- **기간**: 1주 (5-7일)
- **목적**: 작동하는 최소 파이프라인 실행 엔진
- **범위**: 입출력 명시된 YAML 파일을 읽어 실행

## 핵심 기능

### 1. YAML 파싱
```go
type Pipeline struct {
    Name  string `yaml:"name"`
    Steps []Step `yaml:"steps"`
}

type Step struct {
    Name   string `yaml:"name"`
    Run    string `yaml:"run"`
    Input  string `yaml:"input,omitempty"`
    Output string `yaml:"output,omitempty"`
}
```

### 2. 파이프 연결
- Step 간 입출력 매칭
- os/exec 사용한 프로세스 실행
- 파이프를 통한 데이터 전달

### 3. 순차 실행
- 단순 for 루프로 구현
- 에러 시 즉시 중단
- 기본적인 성공/실패 보고

### 4. 파일 로깅
```
~/.cli-pipe/
└── runs/
    └── 2024-01-15/
        └── run-123456/
            ├── pipeline.yaml
            ├── execution.log
            └── steps/
                ├── step1.out
                └── step1.err
```

## 구현 계획

### Day 1-2: 기본 구조
```
cli-pipe/
├── cmd/
│   └── cli-pipe/
│       └── main.go         # 엔트리 포인트
├── internal/
│   ├── pipeline/
│   │   ├── types.go        # 타입 정의
│   │   └── parser.go       # YAML 파서
│   ├── executor/
│   │   ├── executor.go     # 실행 엔진
│   │   └── pipe.go         # 파이프 연결
│   └── logger/
│       └── logger.go       # 파일 로깅
├── go.mod
└── go.sum
```

### Day 3-4: 핵심 로직

#### parser.go
```go
func Parse(content []byte) (*Pipeline, error) {
    var p Pipeline
    err := yaml.Unmarshal(content, &p)
    if err != nil {
        return nil, fmt.Errorf("parse yaml: %w", err)
    }
    return &p, p.Validate()
}
```

#### executor.go
```go
func Execute(p *Pipeline) error {
    streams := make(map[string]io.ReadCloser)
    
    for _, step := range p.Steps {
        // 입력 설정
        var input io.Reader
        if step.Input != "" {
            input = streams[step.Input]
        }
        
        // 명령 실행
        cmd := exec.Command("sh", "-c", step.Run)
        cmd.Stdin = input
        
        // 출력 캡처
        if step.Output != "" {
            output, err := cmd.StdoutPipe()
            if err != nil {
                return err
            }
            streams[step.Output] = output
        }
        
        // 실행
        if err := cmd.Start(); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Day 5-6: 로깅 및 테스트

#### 로깅 구현
- 각 step의 stdout/stderr 캡처
- 타임스탬프 추가
- 실행 시간 측정

#### 테스트 케이스
```yaml
# test/simple.yaml
name: test-pipeline
steps:
  - name: generate
    run: echo "Hello, World!"
    output: text
  - name: transform
    run: tr '[:lower:]' '[:upper:]'
    input: text
    output: result
  - name: display
    run: cat
    input: result
```

### Day 7: 문서화 및 릴리즈

#### README.md
- 설치 방법
- 사용 예제
- YAML 스키마 설명

#### 릴리즈
- v0.1.0 태그
- 바이너리 빌드
- GitHub Release

## 제외 사항 (v0.2로 연기)

1. **변수 시스템**: {{.param}} 지원 없음
2. **병렬 실행**: 순차 실행만
3. **에러 처리**: 기본적인 중단만
4. **진행률**: 단순 로그만
5. **고급 타입**: stream 타입만 지원

## 테스트 계획

### 단위 테스트
- [ ] YAML 파싱
- [ ] 입출력 매칭
- [ ] 명령 실행
- [ ] 로깅

### 통합 테스트
- [ ] 단순 파이프라인
- [ ] 다단계 파이프라인
- [ ] 에러 케이스
- [ ] 큰 데이터 처리

### 예제 파이프라인
```yaml
# examples/backup.yaml
name: simple-backup
steps:
  - name: list-files
    run: find . -type f -name "*.go"
    output: file-list
    
  - name: create-tar
    run: tar cf - -T -
    input: file-list
    output: archive
    
  - name: compress
    run: gzip -9
    input: archive
    output: compressed
    
  - name: save
    run: cat > backup.tar.gz
    input: compressed
```

## 성공 기준

### 기능적 요구사항
- [x] YAML 파일 읽기
- [x] Step 실행
- [x] 파이프 연결
- [x] 로그 저장

### 비기능적 요구사항
- [x] 200줄 이하의 핵심 코드
- [x] 10초 이내 빌드
- [x] 50MB 이하 메모리 사용
- [x] 90% 이상 테스트 커버리지

## 일정

| 일자 | 작업 | 산출물 |
|------|------|--------|
| Day 1 | 프로젝트 구조 | 디렉토리, go.mod |
| Day 2 | 타입 정의 및 파서 | types.go, parser.go |
| Day 3 | 실행 엔진 | executor.go |
| Day 4 | 파이프 연결 | pipe.go |
| Day 5 | 로깅 | logger.go |
| Day 6 | 테스트 | *_test.go |
| Day 7 | 문서화 | README.md |

## 위험 요소

1. **파이프 데드락**
   - 완화: 버퍼 사용
   - 모니터링: 타임아웃

2. **메모리 누수**
   - 완화: defer close
   - 테스트: 장시간 실행

3. **동시성 문제**
   - 완화: 순차 실행만
   - v0.2에서 해결
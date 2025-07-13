# Checkpoint: MVP (v0.1)

## 상태: 계획됨

### 목표
- 입출력이 명시된 YAML 파일을 파싱하여 실행
- Step 간 파이프 연결
- 기본적인 로깅

### 완료 기준
- [ ] YAML 파일 파싱 가능
- [ ] Step 순차 실행
- [ ] 입출력 매칭 및 파이프 연결
- [ ] 실행 로그 파일 저장
- [ ] 기본 에러 처리

### 핵심 기능

#### 1. YAML 구조
```yaml
name: example-pipeline
steps:
  - name: generate
    run: echo "Hello World"
    output: greeting
  - name: transform
    run: tr '[:lower:]' '[:upper:]'
    input: greeting
    output: result
  - name: save
    run: cat > output.txt
    input: result
```

#### 2. 실행 엔진
- 각 Step을 순차적으로 실행
- 명시된 입출력을 파이프로 연결
- 프로세스 간 데이터 전달

#### 3. 로깅
```
~/.cli-pipe/runs/
└── 2024-01-15-123456/
    ├── pipeline.yaml
    ├── execution.log
    └── steps/
        ├── generate.out
        ├── transform.out
        └── save.out
```

### 기술 스택
- Go 1.21+
- gopkg.in/yaml.v3
- 표준 라이브러리 (os/exec, io)

### 제외 사항
- 변수 치환 ({{.param}})
- 병렬 실행
- 고급 에러 처리
- 진행률 표시

### 예상 코드 구조
```
cli-pipe/
├── cmd/cli-pipe/
│   └── main.go
├── internal/
│   ├── pipeline/
│   │   ├── types.go
│   │   └── parser.go
│   ├── executor/
│   │   └── executor.go
│   └── logger/
│       └── logger.go
└── examples/
    └── simple.yaml
```

### 성공 지표
- 200줄 이하의 핵심 코드
- 5개 이상의 예제 파이프라인 실행 가능
- 90% 이상 테스트 커버리지

### 다음 단계
- v0.2: 변수 지원, 병렬 실행
- v0.3: Builder UI
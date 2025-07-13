# Phase 1: Pipeline - 완료 상태

## 목표
- YAML 파이프라인 정의 및 파싱
- 순차적 명령 실행
- 변수 바인딩 및 치환
- 에러 처리 및 중단

## 완성된 아키텍처

```
┌─────────────────────────────────────┐
│         YAML Pipeline               │
│      ~/.cli-pipe/pipelines/         │
│         backup.yaml                 │
└────────────────┬────────────────────┘
                 ▼
        ┌────────────────┐
        │  YAML Parser   │
        └────────┬───────┘
                 ▼
        ┌────────────────┐
        │Pipeline Engine │
        │  - Validator   │
        │  - Variables   │
        └────────┬───────┘
                 ▼
    ┌────────────┴────────────┐
    ▼                         ▼
Step 1: tar cf -         Step 2: gzip
    │                         │
    └────────────┬────────────┘
                 ▼
          ┌──────────────┐
          │   Logger     │
          │(Pipeline Log)│
          └──────────────┘
```

## 추가된 프로젝트 구조

```
cli-pipe/
├── internal/
│   ├── domain/
│   │   ├── pipeline.go     # Pipeline, Step 타입
│   │   └── variable.go     # Variable 정의
│   ├── application/
│   │   ├── parser.go       # YAML 파서 인터페이스
│   │   └── runner.go       # Pipeline 실행
│   └── infrastructure/
│       ├── yaml/
│       │   └── parser.go   # YAML 파싱 구현
│       └── pipeline/
│           └── engine.go   # 파이프라인 엔진
├── pipelines/              # 내장 파이프라인
│   └── examples/
│       ├── backup.yaml
│       └── deploy.yaml
└── tests/
    └── fixtures/
        └── pipelines/      # 테스트용 YAML
```

## YAML 파이프라인 정의

### 기본 구조
```yaml
# ~/.cli-pipe/pipelines/backup.yaml
pipeline:
  name: backup-files
  version: "1.0"
  description: "Backup files with compression"
  
  parameters:
    source:
      type: string
      required: true
      description: "Source directory"
    dest:
      type: string
      default: "./backup.tar.gz"
      
  steps:
    - name: create-archive
      command: "tar cf - {{.source}}"
      
    - name: compress
      command: "gzip -9"
      input: previous
      
    - name: save
      command: "cat > {{.dest}}"
      input: previous
      
    - name: verify
      command: "ls -la {{.dest}}"
```

## 동작하는 기능

### 1. 파이프라인 실행
```bash
$ cli-pipe run pipeline backup.yaml --source=/data --dest=backup.tar.gz
Operation ID: 2024-01-14-145632-pip
Pipeline: backup-files v1.0
Status: Running...

[1/4] create-archive: tar cf - /data
      Status: Success (2.3s)
      
[2/4] compress: gzip -9
      Status: Success (1.2s)
      
[3/4] save: cat > backup.tar.gz
      Status: Success (0.1s)
      
[4/4] verify: ls -la backup.tar.gz
      -rw-r--r-- 1 user staff 45678 Jan 14 14:56 backup.tar.gz
      Status: Success (0.01s)

Pipeline completed successfully
Total duration: 3.61s
```

### 2. 변수 지원
```bash
# 환경 변수
$ export SOURCE=/important/data
$ cli-pipe run pipeline backup.yaml --dest=daily-backup.tar.gz

# 파이프라인 내 변수
steps:
  - name: timestamp
    command: "date +%Y%m%d"
    output_to: timestamp
    
  - name: backup
    command: "tar czf backup-{{.timestamp}}.tar.gz {{.source}}"
```

### 3. 에러 처리
```bash
$ cli-pipe run pipeline backup.yaml --source=/nonexistent
Operation ID: 2024-01-14-150234-err
Pipeline: backup-files v1.0

[1/4] create-archive: tar cf - /nonexistent
      Error: tar: /nonexistent: Cannot stat: No such file or directory
      Exit code: 1
      Status: Failed

Pipeline aborted at step 1
Duration: 0.05s
```

### 4. 파이프라인 목록
```bash
$ cli-pipe pipeline list
Available pipelines:
- backup-files (1.0) - Backup files with compression
- deploy-app (2.1) - Deploy application to server
- db-dump (1.0) - Database backup pipeline

$ cli-pipe pipeline show backup-files
Pipeline: backup-files
Version: 1.0
Parameters:
  - source (string, required): Source directory
  - dest (string, default: ./backup.tar.gz): Destination file
Steps: 4
```

## 저장 형식 확장

### Pipeline Operation JSON
```json
{
  "id": "2024-01-14-145632-pip",
  "type": "pipeline",
  "pipeline": {
    "name": "backup-files",
    "version": "1.0",
    "source": "backup.yaml"
  },
  "parameters": {
    "source": "/data",
    "dest": "backup.tar.gz"
  },
  "steps": [
    {
      "name": "create-archive",
      "command": "tar cf - /data",
      "status": "success",
      "duration_ms": 2300,
      "started_at": "2024-01-14T14:56:32Z",
      "completed_at": "2024-01-14T14:56:34.3Z"
    }
  ],
  "status": "success",
  "total_duration_ms": 3610
}
```

## 핵심 인터페이스 추가

```go
// domain/pipeline.go
type Pipeline struct {
    Name        string
    Version     string
    Description string
    Parameters  []Parameter
    Steps       []Step
}

type Step struct {
    Name    string
    Command string
    Input   InputType
    Output  OutputType
}

// application/runner.go
type PipelineRunner interface {
    Run(ctx context.Context, pipeline *Pipeline, params map[string]string) (*Operation, error)
}

// application/parser.go
type PipelineParser interface {
    Parse(content []byte) (*Pipeline, error)
    LoadFile(path string) (*Pipeline, error)
}
```

## 테스트 커버리지
- YAML Parser: 100%
- Pipeline Engine: 95%
- Variable Binding: 100%
- Error Handling: 90%

## 다음 Phase로의 연결점
- 순차 실행 → 파이프 연결 (스트림)
- 단순 입출력 → 스트림 분기 (tee)
- 텍스트 출력만 → 진행률 표시
- 파일 저장만 → 체크섬 계산
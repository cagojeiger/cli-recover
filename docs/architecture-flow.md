# CLI-Pipe 동작 흐름도

## 1. 실제 구현 흐름

```
┌─────────────┐
│ YAML 파일   │
│pipeline.yaml│
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│         main.go                     │
│  - 명령어 파싱 (run, init, version) │
│  - initializeLogger() 로거 초기화   │
│  - runPipelineCmd() 호출            │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         parser.go                   │
│  - ParseFile(): YAML 읽기           │
│  - ParseBytes(): Go 구조체로 변환   │
│  - yaml.v3 라이브러리 사용          │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         pipeline.go                 │
│  - Validate(): 구조 검증            │
│  - IsLinear(): 선형 파이프라인 체크 │
│  - 중복 스텝명, 잘못된 input 검증   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         builder.go                  │
│  - BuildCommand(): 셸 명령어로 변환 │
│  - 파이프(|)로 명령어 연결          │
│  - 멀티라인 명령어는 괄호로 감싸기  │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         executor.go                 │
│  - Execute(): bash -c로 실행        │
│  - UnifiedMonitor로 메트릭 수집     │
│  - io.MultiWriter로 다중 출력       │
│  - 구조화된 로깅 (logger 사용)      │
│  - 백그라운드 로그 정리             │
└─────────────────────────────────────┘
```

## 2. 선형 파이프라인 실행 예시

### YAML 입력:
```yaml
name: word-count
description: 단어 수 세기
steps:
  - name: generate
    run: echo "hello world from cli-pipe"
    output: text
  - name: count
    run: wc -w
    input: text
```

### 변환 과정:
```
1. ParseFile()로 YAML 파싱
2. Validate()로 구조 검증
3. IsLinear()로 선형 확인 → true
4. BuildCommand()로 변환
```

### 생성된 셸 명령어:
```bash
echo "hello world from cli-pipe" | wc -w
```

### 실제 실행:
```bash
exec.Command("bash", "-c", "echo \"hello world from cli-pipe\" | wc -w")
```

## 3. UnifiedMonitor 구조

```
┌─────────────────────────────────────┐
│        UnifiedMonitor               │
├─────────────────────────────────────┤
│  - byteMonitor: 바이트 수 추적     │
│  - lineMonitor: 라인 수 추적       │
│  - timeMonitor: 실행 시간 측정     │
└─────────────────────────────────────┘
         │
         ▼ Write() 메서드
┌─────────────────────────────────────┐
│  1. 바이트 카운트 증가              │
│  2. '\n' 문자 카운트 (라인 수)      │
│  3. 시작/종료 시간 기록             │
└─────────────────────────────────────┘
```

## 4. io.MultiWriter를 통한 출력 분배

```
                    ┌─▶ os.Stdout (콘솔 출력)
                    │
stdout ──▶ MultiWriter ├─▶ UnifiedMonitor (메트릭 수집)
                    │
                    ├─▶ logFile (pipeline.log)
                    │
                    └─▶ outputFile (있는 경우)

stderr ──▶ MultiWriter ├─▶ os.Stderr (콘솔 에러)
                    │
                    └─▶ stderrFile (stderr.log)
```

## 5. 로그 디렉토리 구조

```
~/.cli-pipe/logs/
├── cli-pipe.log                    # 애플리케이션 로그 (logger 출력)
├── cli-pipe-20240114-150405.log.gz # 로테이션된 압축 로그
└── word-count_20240114_150405/     # 파이프라인 실행 로그
    ├── pipeline.log    # 표준 출력
    ├── stderr.log      # 표준 에러
    └── summary.txt     # 실행 요약
```

### summary.txt 내용:
```
Pipeline: word-count
Duration: 15ms
Bytes: 5 B
Lines: 1
Status: Success
```

## 6. 파일 출력 지원

### YAML:
```yaml
name: save-output
steps:
  - name: generate
    run: echo "데이터"
  - name: save
    run: cat
    output: file:output.txt
```

### 실행 흐름:
```
echo "데이터" | cat
       │
       ▼
MultiWriter ──┬─▶ output.txt (파일 저장)
              ├─▶ 콘솔 출력
              └─▶ 로그 파일
```

**제약사항**: 마지막 스텝의 output만 파일로 저장 가능

## 7. 설정 파일 구조

### 위치: ~/.cli-pipe/config.yaml
```yaml
version: 1
logs:
  directory: ~/.cli-pipe/logs
  retention_days: 30
logger:
  level: info          # debug, info, warn, error
  format: text         # text, json
  output: stderr       # stdout, stderr, file, both
  file_path: ~/.cli-pipe/logs/cli-pipe.log
  max_size: 10         # MB 단위
  max_backups: 3       # 보관할 이전 로그 파일 수
  max_age: 30          # 일 단위
```

### 초기화:
```bash
cli-pipe init  # 설정 파일과 로그 디렉토리 생성
```

## 8. 제약사항 및 특징

### 지원되는 기능:
- ✅ 선형 파이프라인 (단순 파이프 체인)
- ✅ 멀티라인 명령어
- ✅ 자동 모니터링 (바이트, 라인, 시간)
- ✅ 구조화된 로깅 (slog 기반, 레벨별/포맷별 설정)
- ✅ 로그 파일 자동 생성 및 로테이션
- ✅ 오래된 로그 자동 정리
- ✅ 파일 출력 (마지막 스텝만)

### 지원되지 않는 기능:
- ❌ 비선형 파이프라인 (분기, 병합)
- ❌ 스텝별 개별 모니터링 설정
- ❌ 체크섬 생성
- ❌ 향상된 실행 모드
- ❌ 실시간 진행률 표시

이렇게 cli-pipe는 Unix 철학에 충실한 단순하고 효과적인 파이프라인 실행기입니다.
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
│  - IsTree(): 트리 파이프라인 체크   │
│  - AnalyzeStructure(): 구조 분석    │
│  - 중복 스텝명, 잘못된 input 검증   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         builder.go                  │
│  - BuildUnifiedCommand(): 통합 빌더 │
│  - BuildCommand(): 선형 명령어 변환 │
│  - BuildTreeCommand(): 트리 명령어  │
│  - 파이프(|)와 tee로 명령어 연결   │
│  - 멀티라인 명령어는 괄호로 감싸기  │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         executor.go                 │
│  - Execute(): bash -c로 실행        │
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
3. AnalyzeStructure()로 구조 분석 → Linear 타입
4. BuildUnifiedCommand() → BuildCommand()로 변환
```

### 생성된 셸 명령어:
```bash
echo "hello world from cli-pipe" | wc -w | tee /tmp/logs/pipeline.out
```

### 실제 실행:
```bash
exec.Command("bash", "-c", "echo \"hello world from cli-pipe\" | wc -w | tee /tmp/logs/pipeline.out")
```

## 3. 트리 파이프라인 실행 예시

### YAML 입력:
```yaml
name: analyze-data
description: 데이터 분석 및 백업
steps:
  - name: fetch
    run: curl https://api.example.com/data
    output: raw_data
  - name: backup
    run: gzip > backup.gz
    input: raw_data
  - name: analyze
    run: jq '.users | length'
    input: raw_data
  - name: filter
    run: jq '.logs | map(select(.level == "ERROR"))'
    input: raw_data
```

### 변환 과정:
```
1. ParseFile()로 YAML 파싱
2. Validate()로 구조 검증
3. AnalyzeStructure()로 구조 분석 → Tree 타입
4. BuildUnifiedCommand() → BuildTreeCommand()로 변환
```

### 생성된 셸 명령어:
```bash
curl https://api.example.com/data | tee >(gzip > backup.gz) >(jq '.users | length') >(jq '.logs | map(select(.level == "ERROR"))') > /dev/null | tee /tmp/logs/pipeline.out
```

## 4. 구조 분석 (AnalyzeStructure)

```
┌─────────────────────────────────────┐
│        AnalyzeStructure             │
├─────────────────────────────────────┤
│  1. 의존성 그래프 구축              │
│  2. 분기점(branch points) 탐지     │
│  3. 파이프라인 타입 결정:          │
│     - Linear: 분기 없음            │
│     - Tree: 분기만 있음            │
│     - Graph: 병합 있음 (미지원)    │
└─────────────────────────────────────┘
```

### 분기점 탐지 알고리즘:
```go
// O(n) 시간 복잡도로 분기점 탐지
for _, step := range steps {
    if step.Input != "" {
        consumers[step.Input]++
    }
}
// consumers > 1인 output이 분기점
```

## 5. io.MultiWriter를 통한 출력 분배

```
                    ┌─▶ os.Stdout (콘솔 출력)
                    │
stdout ──▶ MultiWriter ├─▶ logFile (pipeline.out)
                    │
                    └─▶ outputFile (있는 경우)

stderr ──▶ MultiWriter ├─▶ os.Stderr (콘솔 에러)
                    │
                    └─▶ stderrFile (stderr.log)
```

## 6. 로그 디렉토리 구조

```
~/.cli-pipe/logs/
├── cli-pipe.log                    # 애플리케이션 로그 (logger 출력)
├── cli-pipe-20240114-150405.log.gz # 로테이션된 압축 로그
└── word-count_20240114_150405/     # 파이프라인 실행 로그
    ├── pipeline.out    # 표준 출력
    ├── stderr.log      # 표준 에러
    └── summary.txt     # 실행 요약
```

### summary.txt 내용:
```
Pipeline: word-count
Duration: 15ms
Status: Success
Command: echo "hello world from cli-pipe" | wc -w
```

## 7. 파일 출력 지원

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

## 8. 설정 파일 구조

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

## 9. 제약사항 및 특징

### 지원되는 기능:
- ✅ 선형 파이프라인 (단순 파이프 체인)
- ✅ 트리 구조 파이프라인 (분기 지원)
- ✅ 멀티라인 명령어
- ✅ 자동 메트릭 수집 (실행 시간, 상태)
- ✅ 구조화된 로깅 (slog 기반, 레벨별/포맷별 설정)
- ✅ 로그 파일 자동 생성 및 로테이션
- ✅ 오래된 로그 자동 정리
- ✅ 파일 출력 (마지막 스텝만)
- ✅ O(n) 시간 복잡도의 효율적인 구조 분석

### 지원되지 않는 기능:
- ❌ 그래프 파이프라인 (병합)
- ❌ 스텝별 개별 모니터링 설정
- ❌ 체크섬 생성
- ❌ 향상된 실행 모드
- ❌ 실시간 진행률 표시

## 10. 트리 파이프라인 고급 예시

### 복잡한 다단계 트리:
```yaml
name: multi-level-tree
steps:
  - name: api_fetch
    run: curl -s https://api.github.com/repos/owner/repo
    output: api_response
  
  # 첫 번째 레벨 분기
  - name: extract_commits
    run: jq '.commits'
    input: api_response
    output: commits_data
  
  - name: extract_issues
    run: jq '.issues'
    input: api_response
    output: issues_data
  
  # 두 번째 레벨 분기
  - name: recent_commits
    run: jq 'map(select(.date > "2024-01-01"))'
    input: commits_data
  
  - name: commit_authors
    run: jq 'map(.author) | unique'
    input: commits_data
  
  - name: open_issues
    run: jq 'map(select(.state == "open"))'
    input: issues_data
```

### 생성된 명령어:
```bash
curl -s https://api.github.com/repos/owner/repo | \
  tee >(jq '.commits' | \
    tee >(jq 'map(select(.date > "2024-01-01"))') \
        >(jq 'map(.author) | unique') > /dev/null) \
      >(jq '.issues' | jq 'map(select(.state == "open"))') > /dev/null | \
  tee /tmp/logs/pipeline.out
```

이렇게 cli-pipe는 Unix 철학에 충실하면서도 트리 구조의 복잡한 파이프라인을 지원하는 강력한 도구입니다.
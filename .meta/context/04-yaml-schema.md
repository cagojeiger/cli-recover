# CLI-PIPE YAML 스키마 정의

## 개요
- **목적**: 파이프라인 정의를 위한 명확하고 일관된 YAML 스키마
- **원칙**: 입출력을 명시하여 데이터 흐름을 명확하게
- **버전**: v1.0 (초기 버전)

## 기본 구조

### 최소 파이프라인
```yaml
name: simple-pipeline
steps:
  - name: step1
    run: echo "Hello World"
    output: greeting
    
  - name: step2
    run: tr '[:lower:]' '[:upper:]'
    input: greeting
    output: result
```

### 전체 구조
```yaml
# 파이프라인 식별자
name: pipeline-name
description: "Optional description"
version: "1.0"

# 파라미터 정의 (선택적)
params:
  param1:
    type: string
    default: "value"
    description: "Parameter description"

# 실행 단계
steps:
  - name: step-name
    run: command to execute
    input: input-stream-name
    output: output-stream-name
    
# 에러 처리 (선택적)
on_error: stop|continue
finally:
  - name: cleanup
    run: cleanup command
```

## Step 정의

### 필수 필드
- `name`: Step의 고유 식별자
- `run`: 실행할 명령어
- `output`: 출력 스트림 이름 (마지막 step 제외)

### 선택적 필드
- `input`: 입력 스트림 이름 (첫 번째 step 제외)
- `inputs`: 복수 입력 (배열)
- `output_type`: 출력 타입 (stream|file|var)
- `on_error`: 에러 처리 (stop|continue|retry)
- `timeout`: 타임아웃 설정

## 입출력 규칙

### 1. 스트림 이름
- 영문자, 숫자, 하이픈, 언더스코어 사용
- 소문자 권장
- 예: `tar-stream`, `compressed_data`, `output1`

### 2. 입력 규칙
- 단일 입력: `input: stream-name`
- 복수 입력: `inputs: [stream1, stream2]`
- 입력 없음: 첫 번째 step이거나 stdin 사용

### 3. 출력 규칙
- 기본: `output: stream-name` (다음 step으로 전달)
- 파일: `output: file:filename.ext` (파일로 저장)
- 변수: `output: var:variable-name` (변수로 캡처)

### 4. 스트림 재사용
```yaml
steps:
  - name: source
    run: tar cf - /data
    output: tar-data
    
  - name: compress
    run: gzip -9
    input: tar-data
    output: compressed
    
  - name: checksum
    run: sha256sum
    input: tar-data  # tar-data 재사용
    output: file:checksum.txt
```

## 파라미터 시스템

### 정의
```yaml
params:
  source:
    type: string      # string, int, bool, path
    required: true
    default: "/data"
    description: "Source directory"
    validation:
      pattern: "^/.*"
```

### 사용
```yaml
steps:
  - name: backup
    run: tar cf - {{.source}}
    output: tar-stream
```

## 특수 Step 타입

### 1. Tee (스트림 분기)
```yaml
steps:
  - name: split
    type: tee
    input: source-stream
    outputs:
      - branch1
      - branch2
      - branch3
```

### 2. Parallel (병렬 실행)
```yaml
steps:
  - name: parallel-backup
    type: parallel
    tasks:
      - name: backup1
        run: tar cf - /app
        output: app-tar
      - name: backup2
        run: tar cf - /data
        output: data-tar
```

## 실행 모델

### 1. 순차 실행 (기본)
- Step이 순서대로 실행
- 이전 step의 output이 다음 step의 input으로

### 2. 병렬 실행
- 의존성이 없는 step들은 동시 실행 가능
- 입출력 관계로 자동 DAG 구성

### 3. 고루틴 파이프라인
- 각 step이 고루틴으로 실행
- io.Pipe로 실시간 스트리밍

## 검증 규칙

### 1. 구조 검증
- 모든 input은 정의된 output을 참조해야 함
- 순환 참조 금지
- 고아 스트림 금지

### 2. 타입 검증
- 파라미터 타입 확인
- output_type 일치성

### 3. 실행 검증
- 명령어 존재 여부
- 권한 확인

## 예제

### 1. 백업 파이프라인
```yaml
name: backup-with-progress
description: "Backup with compression and progress"

params:
  source:
    type: path
    default: /data
  dest:
    type: string
    default: backup.tar.gz

steps:
  - name: estimate-size
    run: du -sb {{.source}} | cut -f1
    output: var:size
    
  - name: create-tar
    run: tar cf - {{.source}}
    output: tar-stream
    
  - name: show-progress
    run: pv -s {{.size}}
    input: tar-stream
    output: progress-stream
    
  - name: compress
    run: gzip -9
    input: progress-stream
    output: compressed
    
  - name: save
    run: cat > {{.dest}}
    input: compressed
```

### 2. 분기 처리
```yaml
name: backup-with-verification
steps:
  - name: create-archive
    run: tar cf - /app
    output: archive
    
  - name: split
    type: tee
    input: archive
    outputs:
      - to-compress
      - to-checksum
      
  - name: compress
    run: gzip -9 > app.tar.gz
    input: to-compress
    
  - name: verify
    run: sha256sum > app.tar.gz.sha256
    input: to-checksum
```

## 버전 진화 계획

### v1.0 (현재)
- 명시적 입출력
- 기본 파라미터
- 순차 실행

### v1.1 (계획)
- 자동 입출력 추론
- 조건부 실행
- 템플릿 지원

### v2.0 (미래)
- CRD 스타일 옵션
- 고급 검증
- 플러그인 시스템
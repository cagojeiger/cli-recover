# CLI-Pipe 동작 흐름도

## 1. 기본 파이프라인 실행 흐름

```
┌─────────────┐
│ YAML 파일   │
│hello.yaml   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Parser    │ ← YAML을 Go 구조체로 변환
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Validator   │ ← 파이프라인 구조 검증
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Builder    │ ← 셸 명령어로 변환
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Executor   │ ← 명령어 실행
└─────────────┘
```

## 2. 선형 파이프라인 변환 예시

### YAML 입력:
```yaml
name: word-count
steps:
  - name: generate
    run: echo "hello world"
    output: text
  - name: count
    run: wc -w
    input: text
```

### 변환된 셸 명령어:
```bash
echo "hello world" | wc -w
```

### 실행 흐름:
```
[generate: echo]  ──pipe──▶  [count: wc -w]  ──▶  출력
```

## 3. 향상된 실행 모드 (--enhanced)

```
┌─────────────────┐
│   YAML 파일     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Enhanced        │
│ Executor        │
└────────┬────────┘
         │
    ┌────┴────┬────────┬──────────┐
    ▼         ▼        ▼          ▼
[Monitor] [Logger] [Checksum] [Progress]
    │         │        │          │
    └─────────┴────────┴──────────┘
                   │
                   ▼
            ┌──────────┐
            │ TeeWriter│ ← 데이터를 여러 곳으로 분배
            └──────────┘
```

## 4. 모니터링 옵션별 동작

### ByteMonitor (monitor.type: bytes)
```
데이터 흐름 ──▶ [ByteMonitor] ──▶ 다음 단계
                      │
                      ▼
                "1024 bytes processed"
```

### LineMonitor (monitor.type: lines)
```
데이터 흐름 ──▶ [LineMonitor] ──▶ 다음 단계
                      │
                      ▼
                "42 lines processed"
```

### ChecksumWriter
```
데이터 흐름 ──▶ [ChecksumWriter] ──▶ 다음 단계
                      │
                      ▼
              SHA256: abc123...
              (파일로 저장)
```

## 5. 복잡한 파이프라인 예시

### YAML:
```yaml
name: complex-pipeline
steps:
  - name: generate
    run: cat data.txt
    output: raw-data
    monitor:
      type: bytes
    
  - name: process
    run: grep "ERROR"
    input: raw-data
    output: errors
    monitor:
      type: lines
    log: errors.log
    
  - name: save
    run: tee
    input: errors
    output: file:errors.txt
    checksum: [sha256]
```

### 실행 흐름:
```
┌──────────┐     ┌──────────┐     ┌──────────┐
│ generate │────▶│ process  │────▶│   save   │
└──────────┘     └──────────┘     └──────────┘
     │                │                 │
     ▼                ▼                 ▼
[ByteMonitor]   [LineMonitor]     [ChecksumWriter]
                      │                 │
                      ▼                 ▼
                 errors.log        errors.txt
                                  errors.txt.sha256
```

## 6. TeeWriter의 비동기 처리

```
                    ┌─▶ [Monitor] ──▶ 출력
                    │
입력 ──▶ [TeeWriter]├─▶ [Logger] ──▶ 파일
                    │
                    └─▶ [Checksum] ──▶ 해시값

* 각 출력은 독립적인 고루틴에서 처리
* 하나가 막혀도 다른 것들은 계속 진행
```

## 7. 실행 모드 비교

### 기본 모드:
```
Step1 | Step2 | Step3
```

### 향상된 모드:
```
┌────────────────────────────────────┐
│        Enhanced Executor           │
├────────────────────────────────────┤
│ Step1 ──Monitor──▶ Step2 ──▶ Step3│
│   ↓                  ↓         ↓   │
│ [Log]              [Log]    [File] │
└────────────────────────────────────┘
```

이렇게 cli-pipe는 단순한 파이프 실행부터 복잡한 모니터링과 로깅까지 지원하는 유연한 구조를 가지고 있습니다.
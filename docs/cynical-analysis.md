# cli-pipe: 과도한 엔지니어링의 전형적 사례

## 요약

cli-pipe는 "단순한 Unix 파이프라인 실행기"로 시작했지만, 실제로는 Jenkins, GitHub Actions, Make가 이미 해결한 문제들을 Go로 재발명하는 프로젝트다. "복잡도 30/100"을 주장하면서도 실제로는 불필요한 추상화와 미래를 위한 과도한 준비로 가득 차 있다.

## 1. "단순함"이라는 거짓말

### 주장 vs 현실

**프로젝트의 주장:**
- "복잡도 30/100"
- "Unix 철학 준수"
- "Occam's Razor 적용"
- "단순한 선형 파이프라인 실행기"

**실제 코드:**
```go
// pipeline.go의 IsLinear() 메서드만 130줄
func (p *Pipeline) IsLinear() bool {
    // 130줄의 복잡한 로직...
}

// 순환 의존성 검사를 위한 DFS 구현
func (p *Pipeline) IsTree() bool {
    // DFS 알고리즘으로 사이클 검증
    // 대학교 자료구조 수업 과제 같은 코드
}
```

단순한 파이프 연결을 위해 그래프 이론까지 동원하는 아이러니.

## 2. Unix 철학의 배신

### Unix가 하는 방식
```bash
# Unix: 단순하고 직관적
tar cf - /data | gzip -9 | tee backup.gz | sha256sum > hash.txt
```

### cli-pipe가 하는 방식
```yaml
# 1. YAML 정의
name: backup
steps:
  - name: archive
    run: tar cf - /data
    output: data
  - name: compress
    run: gzip -9
    input: data
    output: compressed
  - name: hash
    run: sha256sum
    input: data
```

```go
// 2. Go로 파싱
type Pipeline struct {
    Name  string
    Steps []Step
}

// 3. DAG 분석
func BuildDAG(steps []Step) *DAG { }

// 4. 의존성 검증
func (p *Pipeline) Validate() error { }

// 5. 쉘 명령어 재생성
func BuildCommand(p *Pipeline) string { }

// 6. 결국 bash -c로 실행
exec.Command("bash", "-c", cmd).Run()
```

**5단계를 거쳐서 원래 Unix 명령어로 돌아오는 여정.**

## 3. 과도한 추상화의 향연

### 불필요한 인터페이스
```go
// logger 패키지
type Logger interface {
    Debug(msg string, attrs ...any)
    Info(msg string, attrs ...any)
    Warn(msg string, attrs ...any)
    Error(msg string, attrs ...any)
}

// 단순히 slog를 감싸기 위한 인터페이스
// Go 표준 라이브러리가 부족해서?
```

### 미구현 기능들의 무덤
```go
// PipelineType - Graph는 "not yet supported"
const (
    Linear PipelineType = iota
    Tree
    Graph  // 언젠가는... 아마도... 필요하면...
)

// WithContext - 구현은 나중에
func (l *logger) WithContext(ctx context.Context) Logger {
    // TODO: 언젠가 구현할 예정
    return l
}
```

## 4. 해결책을 찾아 문제를 만들기

### OutputLimiter: 바퀴의 재발명
```go
// 50줄 출력 제한을 위한 복잡한 구현
type OutputLimiter struct {
    maxLines    int
    currentLine int
    truncated   bool
}
```

**Unix 방식:**
```bash
command | head -50
```

### 진행률 표시: 고루틴의 낭비
```go
// 1초마다 경과 시간 표시
go func() {
    ticker := time.NewTicker(1 * time.Second)
    for {
        select {
        case <-done:
            return
        case <-ticker.C:
            e.log("\r⏱️  Running: %v", elapsed)
        }
    }
}()
```

**Unix는 그냥 조용히 일한다.**

## 5. 플랫폼 함정

### 현재 → 미래 로드맵
1. **Phase 0 (MVP)**: "단순한" 실행기
2. **Phase 1**: 파라미터, 병렬 실행, DAG
3. **Phase 2**: TUI 빌더, 템플릿 시스템
4. **Phase 3**: 원격 실행, 웹 UI, 분산 처리

**이미 존재하는 것들:**
- Jenkins (2011)
- GitHub Actions (2018)
- GitLab CI (2012)
- CircleCI (2011)
- Ansible (2012)
- Make (1976!)

## 6. 진짜 문제점

### 순환 의존성
```
logger ←→ config
```
config는 logger 설정을 포함하고, logger는 config를 사용한다.

### YAGNI 위반 사례들
- Tree 구조 지원 (실제로 누가 쓸까?)
- 미사용 config.Version 필드
- RetentionDays (로그 정리 코드는 없음)
- 복잡한 의존성 그래프 분석

### 실용성의 부재
```yaml
# 이런 복잡한 구조를 실제로 쓸 사람이?
steps:
  - name: source
    output: data
  - name: process1
    input: data
    output: result1
  - name: process2
    input: data
    output: result2
  - name: merge
    input: result1,result2  # 아직 미지원
```

## 7. 아키텍처의 모순

### "헥사고날 제거" 후에도 남은 복잡성
```
internal/
├── pipeline/     # 여전히 복잡한 구조
│   ├── pipeline.go (200줄)
│   ├── builder.go (243줄)
│   ├── executor.go (360줄)
│   └── parser.go
├── logger/       # 과도한 추상화
└── config/       # 순환 의존성
```

### ASCII 아트의 허영
```
┌─────────┐     ┌──────────┐
│ extract │ ──> │ compress │
└─────────┘     └──────────┘
                      │
                      v
                ┌──────────┐
                │ checksum │
                └──────────┘
```
**이걸 보고 "아, 이해했어!"라고 할 사람이 있을까?**

## 8. 결론

cli-pipe는 "모든 문제는 하나 더 추가된 추상화 레이어로 해결할 수 있다"는 사고방식의 전형적인 예시다. 단순한 bash 스크립트나 Makefile로 충분한 작업을 위해:

- YAML 파싱
- DAG 분석
- 의존성 검증
- Go 런타임
- 수백 줄의 보일러플레이트

를 추가했다.

**진짜 Unix 철학이라면:**
```bash
#!/bin/bash
# pipeline.sh
tar cf - /data | gzip -9 | tee backup.gz | sha256sum > hash.txt
```

**끝.**

## 9. 교훈

1. **단순한 문제는 단순하게 해결하라**
2. **미래를 위한 코드를 미리 쓰지 마라**
3. **이미 존재하는 도구를 재발명하지 마라**
4. **추상화는 비용이다**
5. **"플랫폼"이라는 단어를 조심하라**

> "Perfection is achieved, not when there is nothing more to add, but when there is nothing left to take away."  
> — Antoine de Saint-Exupéry

cli-pipe는 이 격언과 정반대 방향으로 가고 있다.
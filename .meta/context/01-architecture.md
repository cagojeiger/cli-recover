# cli-pipe 아키텍처 설계

## 시스템 개요

```
┌─────────────────────────────────────────────────┐
│                User Interface                    │
│          (CLI Commands / TUI / API)             │
├─────────────────────────────────────────────────┤
│            YAML Pipeline Definition              │
│         (pipelines/*.yaml 파일들)                │
├─────────────────────────────────────────────────┤
│              Go Pipeline Engine                  │
│  (Parser → Validator → Executor → Logger)       │
├─────────────────────────────────────────────────┤
│           Unix Command Execution                 │
│        (sh, tar, tee, kubectl, etc.)           │
└─────────────────────────────────────────────────┘
```

## YAML과 Go 코드의 관계

### YAML 정의 → Go 구조체
- YAML 파이프라인 정의가 Go 구조체로 매핑
- Pipeline, Step, Context 등의 구조체 정의
- 타입 안전성과 검증 로직 포함

### 실행 플로우
- YAML 로드
- 파싱 & 검증
- 변수 바인딩
- Step 실행
- 로깅 & 스트림 처리
- 결과 저장

### 핵심 컴포넌트 관계
- YAML Definition → Pipeline Engine → Operation
- Pipeline Engine ↔ Context (실행 환경)
- Context → Local/SSH/K8s Executor

## 입출력 기반 아키텍처

### 핵심 설계 원칙
- 모든 Step은 명시적인 입력과 출력을 가짐
- 입출력 이름을 통한 Step 간 연결
- 스트림 재사용을 통한 분기 처리

### Step 입출력 모델
```
Step {
    Name: "compress"
    Run: "gzip -9"
    Input: "tar-stream"      // 이전 step의 output 참조
    Output: "compressed"     // 다음 step에서 참조 가능
}
```

### 스트림 관리 (하이브리드)

#### Unix Pipe 모드
```bash
# 스트림은 쉘이 관리
# 이름은 변수로 추적만 함
STREAM_greeting="pipe:1"
STREAM_compressed="pipe:2"
```

#### Go Stream 모드
```go
type StreamManager struct {
    streams map[string]io.ReadCloser  // Go 모드에서만 사용
    mode    ExecutionMode            // Unix/Go 모드 구분
}

// 모드에 따라 다른 동작
func (sm *StreamManager) Connect(step Step) (io.Reader, io.Writer) {
    if sm.mode == UnixPipeMode {
        return nil, nil  // 쉘이 직접 처리
    }
    // Go 모드: 기존 방식
    input := sm.GetInput(step.Input)
    output := sm.CreateOutput(step.Output)
    return input, output
}
```

## 실행 전략 아키텍처 (하이브리드)

### 실행 모드 결정
```go
type ExecutionStrategy interface {
    Execute(pipeline *Pipeline) error
}

// 실행 전략 선택
func DetermineStrategy(pipeline *Pipeline) ExecutionStrategy {
    if isSimpleLinear(pipeline) {
        return &ShellPipeStrategy{}  // Unix pipe
    }
    return &GoStreamStrategy{}      // 세밀한 제어
}
```

### 1. Unix Pipe 전략 (기본)
- 단순 선형 파이프라인에 사용
- 커널 버퍼링으로 데드락 방지
- 높은 성능과 안정성

```bash
# YAML을 쉘 명령으로 변환
tar cf - /data | gzip -9 | tee backup.tar.gz
```

### 2. Go Stream 전략 (필요시)
- 진행률 표시가 필요한 경우
- 복잡한 분기가 있는 경우
- 세밀한 에러 처리가 필요한 경우

```go
type GoStreamStrategy struct {
    streamManager *StreamManager
}

// 진행률 표시를 위한 래퍼
type ProgressReader struct {
    reader    io.Reader
    processed int64
    total     int64
}
```

### 로깅 아키텍처 (필수)
모든 실행은 추적성을 위해 로깅됨:

```bash
# Unix pipe에서도 로깅 보장
(tar cf - /data 2>step1.err | tee step1.log) | \
(gzip -9 2>step2.err | tee step2.log) > output.gz
```

### Go 구현 구조
- ExecutionStrategy 인터페이스: 전략 패턴
- ShellPipeExecutor: Unix pipe 기반 실행
- GoStreamExecutor: io.Reader/Writer 기반 실행
- Logger: 모든 전략에서 공통 사용

## 데이터 흐름

### 명령 실행 흐름
- User Input → CLI Parser
- YAML Loader → Pipeline Builder
- Input/Output Mapping → DAG Construction
- Variable Binding ← Environment/Flags/Config
- Step Executor → Command Runner (고루틴)
- Stream Manager → Named Streams
- Logger → Operation Store

### 입출력 연결 흐름 (하이브리드)

#### Unix Pipe 모드
```
1. YAML 파싱 → Step 정의 추출
2. 입출력 이름 매칭 → 파이프라인 검증
3. 쉘 스크립트 생성 → 파이프(|) 연결
4. 로깅 래퍼 추가 → tee 명령 삽입
5. 쉘 실행 → 커널이 파이프 관리
6. 로그 수집 → 파일로 저장
```

#### Go Stream 모드
```
1. YAML 파싱 → Step 정의 추출
2. 입출력 이름 매칭 → 연결 그래프 생성
3. DAG 분석 → 동시 실행 가능 확인
4. 스트림 생성 → io.Reader/Writer 할당
5. Step 실행 → 고루틴으로 동시 실행
6. 진행률 추적 → ProgressReader 적용
```

### 재실행 흐름
- Replay Command → Operation Loader
- Pipeline Reconstructor → Input/Output 복원
- Context Validator → 환경 확인
- Step Re-executor → 동일한 실행 흐름

## 저장 구조

### 디렉토리 레이아웃
```
~/.cli-pipe/
├── pipelines/          # YAML 정의들
│   ├── builtin/        # 내장 파이프라인
│   └── custom/         # 사용자 정의
├── operations/         # 실행 기록
│   └── 2024-01-13/
│       └── xxxxx/
│           ├── operation.json
│           ├── logs/
│           └── artifacts/
└── index.db           # 검색용 인덱스
```

### 파일 형식
- operation.json: 실행 메타데이터
- logs/: stdout, stderr, events
- artifacts/: 생성된 파일들

## 확장 포인트

### 새 Context 추가
- Docker, Cloud Shell 등
- Context 인터페이스 구현
- 실행 환경별 특화 로직

### 새 Stream Processor
- 압축, 암호화, 필터링
- Stream 인터페이스 구현
- 체이닝 가능한 프로세서

### 새 Pipeline 소스
- Git, HTTP, S3에서 YAML 로드
- Loader 인터페이스 구현
- 동적 파이프라인 로딩
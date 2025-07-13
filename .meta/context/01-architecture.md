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

### 스트림 매니저
```go
type StreamManager struct {
    streams map[string]io.ReadCloser  // 이름으로 스트림 관리
    pipes   map[string]*io.PipeWriter // 활성 파이프
}

// Step 실행 시 입출력 연결
func (sm *StreamManager) Connect(step Step) (io.Reader, io.Writer) {
    input := sm.GetInput(step.Input)
    output := sm.CreateOutput(step.Output)
    return input, output
}
```

## 스트림 처리 아키텍처

### 파이프 분기 메커니즘
- 단일 입력 스트림을 여러 출력으로 분기
- tee 명령어 패턴 활용
- 체크섬, 진행률, 압축을 동시 처리

### 고루틴 기반 동시 실행
```go
// 각 Step을 고루틴으로 실행
for _, step := range pipeline.Steps {
    go func(s Step) {
        input := streamManager.GetInput(s.Input)
        output := streamManager.CreateOutput(s.Output)
        
        cmd := exec.Command("sh", "-c", s.Run)
        cmd.Stdin = input
        cmd.Stdout = output
        
        cmd.Run()
        output.Close()
    }(step)
}
```

### Go 구현 구조
- Stream 인터페이스: Read, Split 메서드
- StreamManager: 스트림 생명주기 관리
- Context 인터페이스: Execute, CreatePipe 메서드
- Engine 구조체: parser, executor, logger 포함

## 데이터 흐름

### 명령 실행 흐름
- User Input → CLI Parser
- YAML Loader → Pipeline Builder
- Input/Output Mapping → DAG Construction
- Variable Binding ← Environment/Flags/Config
- Step Executor → Command Runner (고루틴)
- Stream Manager → Named Streams
- Logger → Operation Store

### 입출력 연결 흐름
```
1. YAML 파싱 → Step 정의 추출
2. 입출력 이름 매칭 → 연결 그래프 생성
3. DAG 분석 → 실행 순서 결정
4. 스트림 생성 → io.Pipe 할당
5. Step 실행 → 입출력 연결
6. 데이터 흐름 → 파이프를 통한 전달
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
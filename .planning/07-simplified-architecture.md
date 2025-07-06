# Simplified Architecture (v0.2.0)

## Overview
극단적으로 단순화된 아키텍처로 복잡도 점수 15-20 유지

## Complexity Score Analysis

### Previous Approach (Score: 80)
```
- 26개 파일 구조: +30
- 여러 레이어 분리: +25  
- Functional state management: +15
- View interface 추상화: +10
```

### New Approach (Score: 15)
```
- main.go 단일 파일: +5
- Golden file 읽기: +5
- 환경변수 스위치: +5
```

## Core Principles

### 1. Start with main.go
- 모든 코드를 main.go에서 시작
- 500줄 제한 엄수
- 필요시에만 분리

### 2. Golden File Testing
```
testdata/
└── kubectl/
    ├── get-namespaces-o-json.golden
    ├── get-pods-n-default-o-json.golden
    └── error-no-access.golden
```

### 3. Simple Command Runner
```go
type Runner interface {
    Run(cmd string, args ...string) ([]byte, error)
}

// Development: Read from files
type GoldenRunner struct {
    dir string
}

// Production: Execute real commands  
type ShellRunner struct{}

// Switch by environment
func NewRunner() Runner {
    if os.Getenv("USE_GOLDEN") == "true" {
        return &GoldenRunner{"testdata"}
    }
    return &ShellRunner{}
}
```

## Development Workflow

### 1. TDD Cycle
```
1. Write test with golden file
2. Implement minimal code
3. Refactor if needed
4. Keep complexity low
```

### 2. File Split Criteria
```
Lines of Code:
- < 300: Keep in main.go
- 300-400: Consider splitting
- 400-500: Should split
- > 500: Must split
```

### 3. Progressive Structure
```
v0.2.0: main.go only
v0.2.1: main.go + 2-3 files if needed
v0.3.0: internal/ package structure
```

## Benefits

1. **Immediate Start**: No complex setup
2. **Easy Testing**: Golden files = predictable
3. **Fast Development**: No abstractions
4. **Clear Code**: Everything visible
5. **Low Complexity**: Score stays under 20

## Migration Path

When main.go exceeds 500 lines:
```
main.go splits into:
├── main.go        # CLI entry (< 200 lines)
├── tui.go         # Bubble Tea UI (< 300 lines)
├── kubectl.go     # K8s operations (< 200 lines)
└── golden.go      # Test runner (< 100 lines)
```

## Anti-patterns to Avoid

1. **Early Abstraction**: No interfaces until needed
2. **Deep Nesting**: Keep directory structure flat
3. **Many Files**: Start with one, split when painful
4. **Over-engineering**: YAGNI principle

## Success Metrics

- Test Coverage > 90%
- Complexity Score < 20
- File Size < 500 lines
- Function Size < 50 lines
- Minimal Dependencies
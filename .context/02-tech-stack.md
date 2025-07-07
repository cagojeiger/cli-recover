# 기술 스택 결정

## UI 프레임워크
### Bubble Tea (유지 결정)
- **장점**
  - 이미 구현되어 있음 (마이그레이션 비용 절감)
  - Elm 아키텍처 (단방향 데이터 플로우)
  - 활발한 커뮤니티와 Charm 생태계
  - teatest로 UI 테스트 가능
  
- **제약사항**
  - goroutine 사용 금지 → tea.Cmd 사용 필수
  - 전체 화면 다시 그리기 (최적화 필요)
  - 상태 불변성 요구

### 대안 검토 결과
- **tview**: k9s에서 사용, 위젯 기반이지만 마이그레이션 비용 높음
- **tcell**: 저수준 라이브러리, 구현 복잡도 증가
- **결론**: Bubble Tea 유지가 최선

## 스타일링
### lipgloss (Charm 생태계)
```go
var (
    headerStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("62")).
        Foreground(lipgloss.Color("230")).
        Bold(true)
        
    selectedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("170"))
)
```

## 테스트 프레임워크
### 단위 테스트
- 표준 `testing` 패키지
- `testify/assert` (선택적)
- `gomock` (인터페이스 모킹)

### UI 테스트
- `teatest` (Bubble Tea 공식)
- 시나리오 기반 테스트 가능

### 통합 테스트
- `testcontainers-go` (K8s API 서버 모킹)
- 실제 kubectl 명령 테스트

## 데이터 구조
### Ring Buffer
```go
type RingBuffer struct {
    data     []string
    size     int
    writePos int
    full     bool
}
```

### 이벤트 소싱
```go
type Event struct {
    ID        string
    Type      EventType
    Timestamp time.Time
    Data      interface{}
}
```

## 설정 관리
### viper 사용
```go
viper.SetConfigName("config")
viper.SetConfigType("yaml")
viper.AddConfigPath("$HOME/.cli-recover")
```

## 로깅
### 구조화된 로깅
```go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Error(msg string, fields ...Field)
}
```
- 파일 기반 로깅: `~/.cli-recover/logs/`
- 로테이션 지원

## 국제화(i18n)
### go-i18n 사용
```go
bundle := i18n.NewBundle(language.Korean)
bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
bundle.LoadMessageFile("locales/ko.yaml")
bundle.LoadMessageFile("locales/en.yaml")
```

## 의존성 주입
### Wire 또는 수동 DI
```go
// 수동 DI 예시
func NewApp(
    kubeClient KubernetesClient,
    backupService BackupService,
    jobManager JobManager,
) *App {
    return &App{
        kubeClient: kubeClient,
        backupService: backupService,
        jobManager: jobManager,
    }
}
```

## 빌드 도구
### Makefile
```makefile
.PHONY: test build run

test:
    go test -v -cover ./...

build:
    go build -ldflags "-X main.version=$(VERSION)" -o cli-recover

run:
    go run ./cmd/cli-recover
```

## 코드 품질 도구
- `golangci-lint`: 정적 분석
- `go fmt`: 코드 포맷팅
- `go vet`: 버그 검출
- `gocyclo`: 순환 복잡도 측정
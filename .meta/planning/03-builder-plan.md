# Builder (v0.3) 구현 계획

## 목표
- **기간**: 2주 (Pipeline 완료 후)
- **목적**: 사용성 극대화
- **범위**: CLI 빌더, 템플릿, 모니터링

## 핵심 기능

### 1. CLI 대화형 빌더

#### 명령어 구조
```bash
# 새 파이프라인 생성
$ cli-pipe new

# 템플릿에서 생성
$ cli-pipe new --template backup

# 기존 파일 편집
$ cli-pipe edit pipeline.yaml
```

#### 대화형 플로우
```
$ cli-pipe new
? Pipeline name: backup-mongodb
? Description: Backup MongoDB with compression
? Add a step? Yes
? Step name: extract
? Command: mongodump --archive
? Output name: mongo-data
? Add another step? Yes
? Step name: compress
? Command: gzip -9
? Input: mongo-data
? Output name: compressed
? Add another step? No

Created: backup-mongodb.yaml
```

#### 구현
```go
type Builder struct {
    pipeline *Pipeline
    prompt   *promptui.Prompt
}

func (b *Builder) Interactive() error {
    b.askBasicInfo()
    for b.askAddStep() {
        step := b.buildStep()
        b.pipeline.Steps = append(b.pipeline.Steps, step)
    }
    return b.save()
}
```

### 2. 템플릿 시스템

#### 템플릿 구조
```
~/.cli-pipe/templates/
├── backup/
│   ├── filesystem.yaml
│   ├── database.yaml
│   └── kubernetes.yaml
├── data/
│   ├── etl.yaml
│   └── migration.yaml
└── custom/
    └── user-templates.yaml
```

#### 템플릿 정의
```yaml
# templates/backup/kubernetes.yaml
template:
  name: kubernetes-backup
  description: "Backup Kubernetes pod data"
  category: backup
  
  parameters:
    - name: pod
      prompt: "Pod name"
      type: string
      required: true
    - name: namespace
      prompt: "Namespace"
      type: string
      default: "default"
      
  pipeline:
    name: "backup-{{.pod}}"
    steps:
      - name: extract
        run: "kubectl exec {{.pod}} -n {{.namespace}} -- tar cf - /data"
        output: pod-data
      - name: compress
        run: gzip -9
        input: pod-data
        output: compressed
      - name: save
        run: "cat > {{.pod}}-backup.tar.gz"
        input: compressed
```

#### 템플릿 관리
```go
type TemplateManager struct {
    templates map[string]*Template
}

func (tm *TemplateManager) List() []Template
func (tm *TemplateManager) Get(name string) *Template
func (tm *TemplateManager) Create(t *Template) error
func (tm *TemplateManager) InstantiateInteractive(name string) (*Pipeline, error)
```

### 3. 실행 모니터링

#### 실시간 로그 뷰어
```
$ cli-pipe run pipeline.yaml --watch

Pipeline: backup-mongodb
Status: Running

Steps:
[✓] extract     - Completed (2.3s)
[⟳] compress    - Running (45% - 1.2s)
[ ] save        - Pending

Logs (compress):
Input: 1.2GB processed
Output: 453MB written
Compression: 62%
```

#### 구현
```go
type Monitor struct {
    execution *Execution
    ui        *termui.UI
}

func (m *Monitor) Start() {
    ticker := time.NewTicker(100 * time.Millisecond)
    for range ticker.C {
        m.updateUI()
        if m.execution.IsComplete() {
            break
        }
    }
}
```

### 4. 파이프라인 검증

#### 검증 레벨
```go
type ValidationLevel int

const (
    ValidationSyntax ValidationLevel = iota
    ValidationLogic
    ValidationRuntime
)
```

#### 검증 구현
```go
func (v *Validator) Validate(p *Pipeline, level ValidationLevel) []ValidationError {
    var errors []ValidationError
    
    // 문법 검증
    errors = append(errors, v.validateSyntax(p)...)
    
    // 논리 검증
    if level >= ValidationLogic {
        errors = append(errors, v.validateInputOutput(p)...)
        errors = append(errors, v.validateDAG(p)...)
    }
    
    // 런타임 검증
    if level >= ValidationRuntime {
        errors = append(errors, v.validateCommands(p)...)
    }
    
    return errors
}
```

### 5. 파이프라인 시각화

#### ASCII 다이어그램
```
$ cli-pipe visualize pipeline.yaml

┌─────────┐     ┌──────────┐     ┌──────────┐
│ extract │ ──> │ compress │ ──> │   save   │
└─────────┘     └──────────┘     └──────────┘
                      │
                      v
                ┌──────────┐
                │ checksum │
                └──────────┘
```

#### Mermaid 출력
```
$ cli-pipe visualize pipeline.yaml --format mermaid

graph LR
    extract[extract] --> compress[compress]
    compress --> save[save]
    compress --> checksum[checksum]
```

## UI/UX 설계

### 1. 명령어 체계
```bash
# 생성
cli-pipe new                    # 대화형
cli-pipe new --template backup  # 템플릿
cli-pipe init                   # 현재 디렉토리

# 실행
cli-pipe run pipeline.yaml      # 기본 실행
cli-pipe run -w pipeline.yaml   # 모니터링
cli-pipe run -p key=val         # 파라미터

# 관리
cli-pipe list                   # 파이프라인 목록
cli-pipe validate pipeline.yaml # 검증
cli-pipe history                # 실행 이력
```

### 2. 설정 파일
```yaml
# ~/.cli-pipe/config.yaml
templates:
  path: ~/.cli-pipe/templates
  auto_update: true

execution:
  default_timeout: 300s
  max_parallel: 10
  
ui:
  color: true
  progress: auto
  log_level: info
```

## 구현 일정

### Week 1: 핵심 Builder

#### Day 1-2: CLI 프레임워크
- [ ] Cobra 통합
- [ ] 명령어 구조
- [ ] 기본 플래그

#### Day 3-4: 대화형 빌더
- [ ] Promptui 통합
- [ ] Step 빌더
- [ ] 검증 통합

#### Day 5-7: 템플릿 시스템
- [ ] 템플릿 파서
- [ ] 템플릿 관리
- [ ] 기본 템플릿 작성

### Week 2: 모니터링 및 완성

#### Day 1-2: 실행 모니터링
- [ ] 실시간 UI
- [ ] 로그 스트리밍
- [ ] 진행률 표시

#### Day 3-4: 시각화
- [ ] ASCII 렌더러
- [ ] Mermaid 생성
- [ ] 의존성 그래프

#### Day 5-6: 통합 테스트
- [ ] E2E 시나리오
- [ ] 사용성 테스트
- [ ] 문서 작성

#### Day 7: 릴리즈
- [ ] 패키징
- [ ] 설치 스크립트
- [ ] 릴리즈 노트

## 템플릿 카탈로그

### 기본 제공 템플릿
1. **Backup**
   - filesystem-backup
   - database-backup
   - kubernetes-backup

2. **Data Processing**
   - etl-pipeline
   - log-processor
   - data-migration

3. **DevOps**
   - ci-pipeline
   - deployment
   - monitoring

### 커뮤니티 템플릿
- GitHub 저장소
- 자동 동기화
- 버전 관리

## 성공 지표

### 사용성
- 5분 안에 첫 파이프라인 생성
- 템플릿으로 80% 사용 사례 커버
- 직관적인 에러 메시지

### 성능
- 빌더 시작: < 100ms
- 템플릿 로드: < 50ms
- UI 업데이트: 60fps

### 품질
- 테스트 커버리지: 90%
- 문서 완성도: 100%
- 예제 파이프라인: 20+

## 향후 계획 (v0.4+)

### TUI 모드
- tview 기반 전체 TUI
- 시각적 파이프라인 편집기
- 드래그 앤 드롭

### 원격 실행
- SSH 지원
- Kubernetes Job
- 분산 실행

### 웹 UI
- 웹 기반 빌더
- 실행 대시보드
- 협업 기능
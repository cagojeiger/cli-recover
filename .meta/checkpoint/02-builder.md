# Checkpoint: Builder (v0.3)

## 상태: 계획됨

### 목표
- 사용자 친화적인 파이프라인 생성 도구
- 템플릿 기반 빠른 시작
- 실시간 실행 모니터링

### 완료 기준
- [ ] CLI 대화형 빌더 (`cli-pipe new`)
- [ ] 템플릿 시스템 구현
- [ ] 실행 모니터링 UI
- [ ] 파이프라인 검증 도구
- [ ] 시각화 기능

### 핵심 기능

#### 1. 대화형 빌더
```bash
$ cli-pipe new
? Pipeline name: backup-daily
? Description: Daily backup routine
? Add a step? Yes
? Step name: collect-data
? Command: find /data -mtime -1 -type f
? Output stream name: recent-files
? Add another step? Yes
...
Created: backup-daily.yaml
```

#### 2. 템플릿 시스템
```bash
$ cli-pipe new --template backup
? Select backup type: Kubernetes Pod
? Pod name: postgres-primary
? Namespace: production
? Include logs? Yes
Created: postgres-backup.yaml from template
```

템플릿 구조:
```yaml
template:
  name: k8s-pod-backup
  category: backup
  parameters:
    - name: pod
      type: string
      prompt: "Pod name"
    - name: namespace
      type: string
      default: "default"
  pipeline:
    name: "backup-{{.pod}}"
    steps: [...]
```

#### 3. 실행 모니터링
```
$ cli-pipe run pipeline.yaml --watch

Pipeline: data-processing
Started: 2024-01-15 10:00:00

┌─ Steps ──────────────────────────────┐
│ [✓] extract    100% (2.3s)          │
│ [▶] transform   45% (1.2s) Running  │
│ [ ] load         0% Pending         │
└──────────────────────────────────────┘

┌─ Current Step Logs ──────────────────┐
│ Processing record 4,523 of 10,000   │
│ Transform rate: 3,769 rec/sec       │
└──────────────────────────────────────┘
```

#### 4. 파이프라인 시각화
```bash
$ cli-pipe visualize pipeline.yaml

extract ──┬──> transform ──> load
          │
          └──> validate ──> report

Dependencies detected:
- transform depends on extract
- validate depends on extract  
- load depends on transform
- report depends on validate
```

### UI/UX 설계

#### 명령어 구조
```bash
# 생성
cli-pipe new                     # 대화형
cli-pipe new --from template.yaml
cli-pipe init                    # 현재 디렉토리

# 실행
cli-pipe run pipeline.yaml
cli-pipe run -w pipeline.yaml    # watch 모드
cli-pipe test pipeline.yaml      # dry run

# 관리
cli-pipe list                    # 사용 가능한 파이프라인
cli-pipe validate pipeline.yaml
cli-pipe explain step-name       # step 설명
```

#### 설정 파일
```yaml
# ~/.cli-pipe/config.yaml
builder:
  default_author: "John Doe"
  preferred_editor: "vim"
  
templates:
  sources:
    - ~/.cli-pipe/templates
    - https://github.com/cli-pipe/templates
    
ui:
  theme: "dark"
  update_interval: 100ms
```

### 아키텍처 추가

```
internal/
├── builder/
│   ├── interactive.go   # 대화형 인터페이스
│   ├── template.go      # 템플릿 엔진
│   └── validator.go     # 검증 로직
├── monitor/
│   ├── watcher.go       # 실행 감시
│   ├── ui.go           # 터미널 UI
│   └── renderer.go      # 렌더링
└── visualizer/
    ├── ascii.go         # ASCII 다이어그램
    └── mermaid.go       # Mermaid 생성
```

### 의존성 추가
- github.com/spf13/cobra (CLI 프레임워크)
- github.com/AlecAivazis/survey/v2 (대화형 프롬프트)
- github.com/gizak/termui/v3 (터미널 UI)

### 템플릿 카탈로그

기본 제공:
1. **Backup 템플릿**
   - filesystem-backup
   - database-backup (MySQL, PostgreSQL, MongoDB)
   - kubernetes-backup

2. **Data 처리 템플릿**
   - etl-pipeline
   - log-processor
   - batch-job

3. **DevOps 템플릿**
   - ci-build
   - deployment
   - health-check

### 성공 지표
- 5분 내 첫 파이프라인 생성
- 템플릿으로 80% 사용 사례 커버
- 직관적인 UI/UX (사용자 테스트 90% 만족)

### 다음 단계
- v0.4: TUI 모드 (tview)
- v1.0: 웹 UI
- v2.0: 분산 실행
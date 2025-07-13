# Phase 3: Multi-Context - 완료 상태

## 목표
- 다양한 실행 환경 지원 (Local, SSH, Kubernetes)
- 컨텍스트 간 전환 및 조합
- 원격 파일 전송 (SCP, kubectl cp)
- 환경별 최적화

## 완성된 아키텍처

```
┌─────────────────────────────────────┐
│         Pipeline Engine             │
│    (Context-aware execution)        │
└────────────────┬────────────────────┘
                 ▼
        ┌────────────────┐
        │Context Manager │
        │ - Registry     │
        │ - Switcher     │
        └────────┬───────┘
                 │
    ┌────────────┼────────────┐
    ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌────────┐
│ Local  │  │  SSH   │  │  K8s   │
│Context │  │Context │  │Context │
└────────┘  └────────┘  └────────┘
    │            │            │
    ▼            ▼            ▼
Local Exec   ssh/scp    kubectl
                        exec/cp/pf
```

## 추가된 프로젝트 구조

```
cli-pipe/
├── internal/
│   ├── domain/
│   │   └── context.go      # Context 인터페이스
│   ├── application/
│   │   └── context/
│   │       ├── manager.go  # Context 관리
│   │       └── registry.go # Context 레지스트리
│   └── infrastructure/
│       └── context/
│           ├── local/      # 로컬 실행
│           │   └── executor.go
│           ├── ssh/        # SSH 실행
│           │   ├── executor.go
│           │   └── transfer.go
│           └── k8s/        # Kubernetes
│               ├── executor.go
│               ├── transfer.go
│               └── portforward.go
├── config/
│   └── contexts.yaml       # 컨텍스트 설정
└── tests/
    └── context/            # 컨텍스트 테스트
```

## 컨텍스트 정의

### contexts.yaml
```yaml
contexts:
  # 로컬 컨텍스트
  local:
    type: local
    default: true
    
  # SSH 컨텍스트들
  prod-server:
    type: ssh
    host: prod.example.com
    user: deploy
    key: ~/.ssh/id_rsa
    
  backup-server:
    type: ssh
    host: backup.internal
    user: backup
    port: 2222
    
  # Kubernetes 컨텍스트들
  k8s-prod:
    type: kubernetes
    context: production
    namespace: default
    
  k8s-staging:
    type: kubernetes
    context: staging
    namespace: apps
```

## 멀티 컨텍스트 파이프라인

### 크로스 컨텍스트 백업
```yaml
pipeline:
  name: cross-context-backup
  steps:
    # K8s에서 데이터 추출
    - name: extract-from-pod
      context: k8s-prod
      command: "kubectl exec {{.pod}} -- tar cf - {{.path}}"
      output: stream
      
    # 로컬에서 압축
    - name: compress-local
      context: local
      command: "gzip -9"
      input: pipe
      output: stream
      
    # 백업 서버로 전송
    - name: save-to-backup
      context: backup-server
      command: "cat > /backups/{{.filename}}"
      input: pipe
```

### 포트포워딩 + MinIO
```yaml
pipeline:
  name: minio-backup
  steps:
    # 포트포워딩 설정
    - name: setup-portforward
      context: k8s-prod
      type: port-forward
      service: minio
      ports: "9000:9000"
      background: true
      
    # MinIO 작업
    - name: mirror-bucket
      context: local
      command: "mc mirror minio/{{.bucket}} ./backup/"
      depends_on: setup-portforward
      
    # 정리
    - name: cleanup
      type: cleanup
      stop: setup-portforward
```

## 동작하는 기능

### 1. 컨텍스트 전환
```bash
$ cli-pipe run backup.yaml --context=k8s-prod
Running in context: k8s-prod (kubernetes)

$ cli-pipe context list
NAME           TYPE         DEFAULT   STATUS
local          local        ✓         ready
prod-server    ssh                    ready
backup-server  ssh                    ready
k8s-prod       kubernetes             ready (current)
k8s-staging    kubernetes             not connected
```

### 2. 원격 실행
```bash
# SSH 실행
$ cli-pipe run --context=prod-server "df -h"
Operation ID: 2024-01-14-170234-ssh
Context: prod-server (ssh://deploy@prod.example.com)
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1       100G   45G   55G  45% /

# Kubernetes 실행
$ cli-pipe run --context=k8s-prod "kubectl exec nginx-pod -- ls /usr/share/nginx/html"
index.html
50x.html
```

### 3. 파일 전송
```bash
# SCP 전송
$ cli-pipe transfer local:./backup.tar.gz prod-server:/backups/
Transfer: local → prod-server
Size: 234MB
Progress: [=========>..] 78% 183MB/234MB 12MB/s

# kubectl cp
$ cli-pipe transfer k8s-prod:nginx-pod:/var/log/nginx/ local:./logs/
Transfer: k8s-prod:nginx-pod → local
Files: 23
Progress: [=====>.....] 52% 12/23 files
```

### 4. 복합 시나리오
```bash
$ cli-pipe run multi-env-deploy.yaml
[1/5] build: (local) docker build -t myapp:v2
      ✓ Success (45s)
      
[2/5] push: (local) docker push registry/myapp:v2  
      ✓ Success (23s)
      
[3/5] deploy: (k8s-staging) kubectl set image deployment/myapp myapp=registry/myapp:v2
      ✓ Success (2s)
      
[4/5] wait: (k8s-staging) kubectl rollout status deployment/myapp
      ✓ Success (30s)
      
[5/5] notify: (local) curl -X POST https://slack.webhook...
      ✓ Success (0.5s)
```

## 컨텍스트 메타데이터

### Context-aware Operation JSON
```json
{
  "id": "2024-01-14-170234-ctx",
  "contexts_used": [
    {
      "name": "k8s-prod",
      "type": "kubernetes",
      "steps": ["extract-from-pod"],
      "connection_time_ms": 150
    },
    {
      "name": "local", 
      "type": "local",
      "steps": ["compress-local"]
    },
    {
      "name": "backup-server",
      "type": "ssh",
      "steps": ["save-to-backup"],
      "connection_time_ms": 230
    }
  ],
  "transfers": [
    {
      "from": "k8s-prod:nginx-pod",
      "to": "local",
      "bytes": 1073741824,
      "duration_ms": 12450
    }
  ]
}
```

## 핵심 인터페이스 추가

```go
// domain/context.go
type Context interface {
    Name() string
    Type() ContextType
    Execute(cmd Command) (*Result, error)
    Transfer(src, dst string) error
    Stream(cmd Command) (Stream, error)
    Close() error
}

// application/context/manager.go
type ContextManager interface {
    Register(name string, ctx Context) error
    Get(name string) (Context, error)
    Switch(name string) error
    Current() Context
}

// SSH Context 특화
type SSHContext interface {
    Context
    PortForward(local, remote string) (Forwarder, error)
}

// K8s Context 특화
type K8sContext interface {
    Context
    Exec(pod, container string, cmd Command) (*Result, error)
    Copy(pod, container, src, dst string) error
    PortForward(service string, ports []string) (Forwarder, error)
}
```

## 컨텍스트별 최적화

### 로컬 최적화
- 직접 시스템 콜 사용
- 파일 시스템 네이티브 작업

### SSH 최적화
- 연결 재사용 (ControlMaster)
- 압축 전송 옵션
- 병렬 전송 지원

### Kubernetes 최적화
- API 클라이언트 캐싱
- 스트리밍 exec/attach
- 효율적인 파일 전송

## 테스트 커버리지
- Context Interface: 100%
- Local Context: 95%
- SSH Context: 85%
- K8s Context: 80%
- Integration: 멀티 컨텍스트 시나리오

## 다음 Phase로의 연결점
- 명령줄 플래그 → 자동 CLI 생성
- 설정 파일 → 인터랙티브 TUI
- 파라미터 입력 → 자동 완성
- YAML 정의 → UI 자동 생성
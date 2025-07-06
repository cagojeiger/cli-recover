# TUI Backup Feature Design

## 사용자 플로우

### 1. TUI 모드 (`cli-restore tui`)
```
CLI Restore - Interactive Mode
==============================

? Select namespace: 
  > default
    kube-system  
    production
    
? Select pod: 
  > my-app-pod        (Running, 2/2)
    nginx-pod         (Running, 1/1)
    mongo-pod         (Running, 1/1)
    
? Select path to backup: 
  > /data
    /logs
    /config
    Custom path...
    
? Split size: 
  > 1G
    2G
    5G
    Custom...
    
? Confirm backup settings:
  Pod: my-app-pod
  Namespace: default
  Path: /data
  Split: 1G
  
  > Execute backup
    Show CLI command
    Cancel
    
Generated CLI command:
cli-restore backup my-app-pod /data --namespace default --split-size 1G

Executing backup...
```

### 2. CLI 모드 (`cli-restore backup`)
```bash
# 직접 실행 (스크립트용)
cli-restore backup my-app-pod /data --namespace default --split-size 1G

# 도움말
cli-restore backup --help
```

## 기술 스택

### TUI 라이브러리
- **Survey**: 가장 가벼운 프롬프트 라이브러리
- **의존성**: 단일 패키지만 추가
- **기능**: Select, Input, Confirm 프롬프트

### kubectl 통합
- `kubectl get namespaces`: 네임스페이스 목록
- `kubectl get pods`: Pod 목록 (상태 포함)
- `kubectl exec`: 실제 백업 실행

## 구현 우선순위

### Phase 1: 기본 TUI
1. Survey 의존성 추가
2. 네임스페이스 선택 프롬프트
3. Pod 선택 프롬프트
4. 경로 입력 프롬프트
5. 확인 및 CLI 명령어 생성

### Phase 2: 백업 실행
1. kubectl exec + tar 파이프라인
2. 분할 압축 (1G 단위)
3. 진행률 표시
4. 에러 처리

### Phase 3: 고급 기능
1. 일반적인 경로 자동 감지
2. Pod 상태 표시
3. 커스텀 옵션 지원

## 에러 처리

### 의존성 확인
- kubectl 설치 확인
- kubeconfig 접근 확인
- 네트워크 연결 확인

### 사용자 가이드
- kubectl 미설치 시 설치 가이드
- 권한 부족 시 해결 방법
- Pod 접근 불가 시 디버깅 정보

## 명령어 구조

```
cli-restore
├── --version          # 기존 기능
├── tui                # 새로운 TUI 모드
└── backup <pod> <path> # 새로운 CLI 모드
    ├── --namespace, -n
    ├── --split-size, -s
    └── --output, -o
```
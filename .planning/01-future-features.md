# Future Features Planning

## Backup Command Design

### Command Structure
```bash
cli-restore backup <pod-name> <source-path> [flags]
```

### Flags
- `--namespace, -n`: Kubernetes namespace (default: default)
- `--split-size, -s`: 분할 크기 (예: 100M, 1G)
- `--output, -o`: 출력 디렉토리 (default: ./backup)
- `--compress-level`: 압축 레벨 (1-9)
- `--exclude`: 제외할 파일/폴더 패턴

### Implementation Architecture

#### Directory Structure (Phase 2)
```
internal/
├── commands/
│   ├── root.go
│   ├── version.go
│   └── backup.go
├── k8s/
│   ├── client.go      # kubeconfig 로드
│   ├── pod.go         # Pod 작업
│   └── exec.go        # kubectl exec 래퍼
├── archive/
│   ├── tar.go         # tar 생성
│   ├── split.go       # 파일 분할
│   └── compress.go    # gzip 압축
└── transfer/
    └── copy.go        # 파일 전송 로직
```

### Technical Approach
1. kubectl exec로 Pod 내부에서 tar 실행
2. stdout으로 tar 스트림 받기
3. 로컬에서 분할 및 압축
4. 진행 상황 실시간 표시

### Error Handling
- Pod 접근 권한 확인
- 디스크 공간 체크
- 네트워크 중단 시 재시도
- 부분 백업 복구
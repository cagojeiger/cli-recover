# 진행률 보고 예제 모음

## 1. 백업 작업 예제

### 시나리오
1GB 크기의 디렉토리를 백업하는 상황

### 터미널 출력
```bash
$ cli-recover backup filesystem nginx-pod /data
[2025-01-08 15:30:01] [INFO] Starting backup for pod: nginx-pod
[2025-01-08 15:30:02] [INFO] Estimated size: 1.2GB
Creating backup... [████████████░░░░░░░░] 60% (720MB/1.2GB) ETA: 30s
```

**진행 중 모습 (애니메이션)**
```
Creating backup... [█░░░░░░░░░░░░░░░░░░░] 5% (60MB/1.2GB) ETA: 1m 35s
Creating backup... [███░░░░░░░░░░░░░░░░░] 15% (180MB/1.2GB) ETA: 1m 20s
Creating backup... [██████░░░░░░░░░░░░░░] 30% (360MB/1.2GB) ETA: 1m 05s
Creating backup... [█████████░░░░░░░░░░░] 45% (540MB/1.2GB) ETA: 50s
Creating backup... [████████████░░░░░░░░] 60% (720MB/1.2GB) ETA: 35s
Creating backup... [███████████████░░░░░] 75% (900MB/1.2GB) ETA: 20s
Creating backup... [██████████████████░░] 90% (1.08GB/1.2GB) ETA: 10s
Creating backup... [████████████████████] 100% (1.2GB/1.2GB) ETA: 0s
✓ Backup created successfully: backup-20250108-153001.tar
```

### 로그 파일 출력
```log
2025-01-08T15:30:01Z [INFO] Starting backup for pod: nginx-pod namespace=default
2025-01-08T15:30:02Z [INFO] Estimated size: 1288490188 bytes
2025-01-08T15:30:02Z [INFO] Creating backup file: backup-20250108-153001.tar
2025-01-08T15:30:12Z [INFO] Progress update message="Creating backup" bytes_current=128849018 bytes_total=1288490188 percent=10
2025-01-08T15:30:22Z [INFO] Progress update message="Creating backup" bytes_current=257698036 bytes_total=1288490188 percent=20
2025-01-08T15:30:32Z [INFO] Progress update message="Creating backup" bytes_current=386547054 bytes_total=1288490188 percent=30
2025-01-08T15:30:42Z [INFO] Progress update message="Creating backup" bytes_current=515396072 bytes_total=1288490188 percent=40
2025-01-08T15:30:52Z [INFO] Progress update message="Creating backup" bytes_current=644245090 bytes_total=1288490188 percent=50
2025-01-08T15:31:02Z [INFO] Progress update message="Creating backup" bytes_current=773094108 bytes_total=1288490188 percent=60
2025-01-08T15:31:12Z [INFO] Progress update message="Creating backup" bytes_current=901943126 bytes_total=1288490188 percent=70
2025-01-08T15:31:22Z [INFO] Progress update message="Creating backup" bytes_current=1030792144 bytes_total=1288490188 percent=80
2025-01-08T15:31:32Z [INFO] Progress update message="Creating backup" bytes_current=1159641162 bytes_total=1288490188 percent=90
2025-01-08T15:31:42Z [INFO] Progress update message="Creating backup" bytes_current=1288490188 bytes_total=1288490188 percent=100
2025-01-08T15:31:42Z [INFO] Operation completed message="Backup created successfully"
2025-01-08T15:31:42Z [INFO] Backup metadata saved backup_id=backup-20250108-153001 size=1288490188 duration=100s
```

### CI/CD 환경 출력
```
$ CI=true cli-recover backup filesystem nginx-pod /data
[2025-01-08 15:30:01] [INFO] Starting backup for pod: nginx-pod
[2025-01-08 15:30:02] [INFO] Estimated size: 1.2GB
[2025-01-08 15:30:02] [INFO] Creating backup file: backup-20250108-153001.tar
[2025-01-08 15:30:12] [INFO] Progress: 128MB/1.2GB (10%)
[2025-01-08 15:30:22] [INFO] Progress: 256MB/1.2GB (20%)
[2025-01-08 15:30:32] [INFO] Progress: 384MB/1.2GB (30%)
[2025-01-08 15:30:42] [INFO] Progress: 512MB/1.2GB (40%)
[2025-01-08 15:30:52] [INFO] Progress: 640MB/1.2GB (50%)
[2025-01-08 15:31:02] [INFO] Progress: 768MB/1.2GB (60%)
[2025-01-08 15:31:12] [INFO] Progress: 896MB/1.2GB (70%)
[2025-01-08 15:31:22] [INFO] Progress: 1024MB/1.2GB (80%)
[2025-01-08 15:31:32] [INFO] Progress: 1152MB/1.2GB (90%)
[2025-01-08 15:31:42] [INFO] Progress: 1280MB/1.2GB (100%)
[2025-01-08 15:31:42] [INFO] Backup completed successfully
[2025-01-08 15:31:42] [INFO] Output file: backup-20250108-153001.tar (1.2GB)
```

## 2. 도구 다운로드 예제

### 터미널 출력
```bash
$ cli-recover backup filesystem nginx-pod /data
kubectl not found in PATH
Downloading kubectl v1.28.0... [██████████████░░░░░░] 70% (35MB/50MB) 2.5MB/s ETA: 6s
```

**완료 후**
```
kubectl not found in PATH
Downloading kubectl v1.28.0... [████████████████████] 100% (50MB/50MB) 2.8MB/s
✓ kubectl installed to ~/.cli-recover/tools/kubectl
[2025-01-08 15:35:01] [INFO] Starting backup for pod: nginx-pod
```

### 네트워크 오류 시
```bash
$ cli-recover backup filesystem nginx-pod /data
kubectl not found in PATH
Downloading kubectl v1.28.0... [████░░░░░░░░░░░░░░░░] 20% (10MB/50MB) 1.2MB/s
✗ Download failed: connection timeout

Please install kubectl manually:
  macOS:   brew install kubectl
  Linux:   curl -LO https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl
  Windows: Download from https://dl.k8s.io/release/v1.28.0/bin/windows/amd64/kubectl.exe
```

## 3. 복원 작업 예제

### 터미널 출력
```bash
$ cli-recover restore filesystem nginx-pod backup-20250108-153001.tar
[2025-01-08 16:00:01] [INFO] Starting restore to pod: nginx-pod
[2025-01-08 16:00:02] [INFO] Backup size: 1.2GB
Restoring files... [███████░░░░░░░░░░░░░] 35% (420MB/1.2GB) ETA: 1m 10s
```

### 다중 파일 진행률
```
Restoring files... (156/1024 files) [███████░░░░░░░░░░░░░] 35% (420MB/1.2GB)
  └─ Current: /data/logs/app-2025-01-07.log (45MB)
```

## 4. 크기를 모르는 작업

### tar 스트리밍 백업
```bash
$ cli-recover backup filesystem nginx-pod /data --no-estimate
Creating backup... 523MB processed (12.5MB/s) [streaming]
```

### 로그 파일
```log
2025-01-08T16:10:12Z [INFO] Progress update message="Creating backup" bytes_current=134217728 bytes_total=0
2025-01-08T16:10:22Z [INFO] Progress update message="Creating backup" bytes_current=268435456 bytes_total=0
2025-01-08T16:10:32Z [INFO] Progress update message="Creating backup" bytes_current=402653184 bytes_total=0
```

## 5. TUI 통합 예제

### TUI 화면
```
┌─ Backup Progress ─────────────────────────────────────┐
│                                                       │
│  Pod:        nginx-pod                                │
│  Namespace:  default                                  │
│  Path:       /data                                    │
│  Type:       filesystem                               │
│                                                       │
│  Progress:                                            │
│  ████████████████████░░░░░░░░░  72%                  │
│                                                       │
│  Size:       864MB / 1.2GB                           │
│  Speed:      15.2 MB/s                               │
│  Time:       00:01:23 / 00:01:55                    │
│  ETA:        32 seconds                              │
│                                                       │
│  Recent Logs:                                         │
│  [16:20:32] Processing /data/images/photo_1024.jpg   │
│  [16:20:31] Processing /data/images/photo_1023.jpg   │
│  [16:20:30] Processing /data/images/photo_1022.jpg   │
│                                                       │
│                              [Cancel] [Background]    │
└───────────────────────────────────────────────────────┘
```

### 코드에서 TUI 채널 사용
```go
// TUI 시작 시
progressCh := make(chan Progress, 100)
reporter.SetProgressChannel(progressCh)

// TUI 업데이트 루프
go func() {
    for p := range progressCh {
        tui.UpdateProgress(p.Current, p.Total, p.Message)
        tui.SetSpeed(p.Speed)
        tui.SetETA(calculateETA(p))
    }
}()
```

## 6. 에러 처리 예제

### 백업 중단
```bash
$ cli-recover backup filesystem nginx-pod /data
Creating backup... [██████░░░░░░░░░░░░░░] 30% (360MB/1.2GB) ETA: 1m 05s
^C
✗ Backup interrupted by user
Cleaning up temporary files... Done
```

### 공간 부족
```bash
$ cli-recover backup filesystem nginx-pod /data
Creating backup... [████████████████░░░░] 80% (960MB/1.2GB) ETA: 15s
✗ Backup failed: no space left on device
Cleaning up temporary files... Done

Free space required: 240MB
Available space: 50MB
```

## 7. 병렬 작업 진행률

### 여러 Pod 동시 백업
```bash
$ cli-recover backup filesystem pod1,pod2,pod3 /data --parallel
Starting parallel backups...

[pod1] Creating backup... [████████░░░░░░░░░░░░] 40% (200MB/500MB)
[pod2] Creating backup... [██████████████░░░░░░] 70% (700MB/1GB)
[pod3] Creating backup... [████████████████████] 100% (300MB/300MB) ✓

Total progress: 2/3 pods completed, 1.2GB/1.8GB (67%)
```

## 8. 로그 모드 강제

### --log-progress 플래그 사용
```bash
$ cli-recover backup filesystem nginx-pod /data --log-progress
[2025-01-08 16:30:01] [INFO] Starting backup for pod: nginx-pod
[2025-01-08 16:30:02] [INFO] Estimated size: 1.2GB
[2025-01-08 16:30:12] [INFO] Progress: 10% (120MB/1.2GB)
[2025-01-08 16:30:22] [INFO] Progress: 20% (240MB/1.2GB)
# 터미널에서도 \r 없이 로그 형태로 출력
```

## 9. 상세 진행률 (Verbose)

### -v 플래그 사용
```bash
$ cli-recover backup filesystem nginx-pod /data -v
[2025-01-08 16:40:01] [DEBUG] Connecting to Kubernetes cluster
[2025-01-08 16:40:01] [DEBUG] Found pod: nginx-pod (Ready)
[2025-01-08 16:40:02] [INFO] Estimating backup size...
[2025-01-08 16:40:02] [DEBUG] Running: kubectl exec nginx-pod -- du -sb /data
[2025-01-08 16:40:03] [INFO] Estimated size: 1.2GB
Creating backup... [████░░░░░░░░░░░░░░░░] 20% (240MB/1.2GB) 12MB/s
  └─ Processing: /data/logs/access.log (15MB)
  └─ Files: 156/1850
  └─ Skipped: 0
  └─ Errors: 0
```

## 10. JSON 출력 모드

### --output json 사용
```bash
$ cli-recover backup filesystem nginx-pod /data --output json --progress
{"timestamp":"2025-01-08T16:50:01Z","level":"INFO","message":"Starting backup","pod":"nginx-pod"}
{"timestamp":"2025-01-08T16:50:11Z","level":"INFO","message":"Progress","current":128000000,"total":1288490188,"percent":10}
{"timestamp":"2025-01-08T16:50:21Z","level":"INFO","message":"Progress","current":256000000,"total":1288490188,"percent":20}
```

## 구현 팁

### 1. 부드러운 애니메이션
```go
// 스무스한 진행바를 위한 업데이트 간격
ticker := time.NewTicker(200 * time.Millisecond)
defer ticker.Stop()
```

### 2. 터미널 크기 대응
```go
// 터미널 크기 변경 감지
sigwinch := make(chan os.Signal, 1)
signal.Notify(sigwinch, syscall.SIGWINCH)

go func() {
    for range sigwinch {
        // 진행바 다시 그리기
        updateProgressBar()
    }
}()
```

### 3. 색상 사용 (옵션)
```go
// 진행률에 따른 색상
switch {
case percent < 30:
    color = "\033[31m" // 빨강
case percent < 70:
    color = "\033[33m" // 노랑
default:
    color = "\033[32m" // 초록
}
```
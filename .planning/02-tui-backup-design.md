# TUI Backup Feature Design

## 사용자 플로우

### 1. TUI 모드 (`cli-restore`)
```
┌─ CLI Restore v0.2.0 ────────────────────────── cluster: prod-k8s ─┐
│ Main Menu                                                          │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│   > Backup                                                         │
│     Restore                                                        │
│     Verify                                                         │
│     History                                                        │
│                                                                    │
├─ Command Preview ──────────────────────────────────────────────────┤
│ $ cli-restore backup                                               │
├────────────────────────────────────────────────────────────────────┤
│ ↑/↓ Navigate  Enter Select  q Quit  ? Help                        │
└────────────────────────────────────────────────────────────────────┘
```

백업 타겟 선택:
```
┌─ CLI Restore v0.2.0 ────────────────────────── cluster: prod-k8s ─┐
│ Main Menu > Backup > Select Target                                 │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│   Containers                                                       │
│   > Pod (Files/Directories)                                        │
│                                                                    │
│   Databases                                                        │
│     MongoDB (Bitnami)                                              │
│     PostgreSQL                                                     │
│     MySQL                                                          │
│                                                                    │
│   Object Storage                                                   │
│     MinIO (Bitnami)                                                │
│     S3 Compatible                                                  │
│                                                                    │
├─ Command Preview ──────────────────────────────────────────────────┤
│ $ cli-restore backup pod                                           │
├────────────────────────────────────────────────────────────────────┤
│ ↑/↓ Navigate  Enter Select  b Back  q Quit                        │
└────────────────────────────────────────────────────────────────────┘
```

### 2. CLI 모드 (`cli-restore [action] [target]`)
```bash
# Pod 파일시스템 백업
cli-restore backup pod nginx-app /data --namespace prod --split-size 1G

# MongoDB 백업
cli-restore backup mongodb mongo-primary --all-databases

# MinIO 백업
cli-restore backup minio minio-server my-bucket --recursive

# 도움말
cli-restore backup --help
```

## 기술 스택

### TUI 프레임워크
- **Bubble Tea**: Elm-inspired functional TUI framework
- **특징**: 
  * 풀스크린 모드
  * 실시간 업데이트
  * 복잡한 상태 관리
  * k9s 스타일 구현 가능

### kubectl 통합
- `kubectl get namespaces`: 네임스페이스 목록
- `kubectl get pods`: Pod 목록 (상태 포함)
- `kubectl exec`: 백업/복원 실행
- `kubectl port-forward`: 서비스 접근

## 구현 우선순위

### Phase 1: 기본 TUI 프레임워크
1. Bubble Tea 기반 구조 설계
2. 표준 레이아웃 시스템 구현
3. 네비게이션 스택 관리
4. 명령어 프리뷰 시스템
5. 단축키 바인딩

### Phase 2: 백업 기능 구현
1. Pod 파일시스템 백업
   - kubectl exec + tar 스트리밍
   - 분할 압축 (크기 기반)
   - 진행률 실시간 표시
2. MongoDB 백업 (mongodump)
3. MinIO 백업 (mc 자동 처리)

### Phase 3: 고급 기능
1. 용량 기반 백업 전략 자동 선택
2. 오프라인 모드 (임베디드 바이너리)
3. 백업 히스토리 관리
4. 스케줄링 지원

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
├── --version                    # 버전 확인
├── (no args)                    # TUI 모드 실행
└── [action] [target] [options]  # CLI 직접 실행
    ├── backup
    │   ├── pod <name> <path>
    │   ├── mongodb <name>
    │   └── minio <name> <bucket>
    ├── restore
    │   ├── pod <backup> <name>
    │   ├── mongodb <dump> <name>
    │   └── minio <backup> <name>
    └── verify, history, schedule
```

## 백업 전략

### 크기 기반 자동 선택
- **< 10GB**: Pod 내부 백업 (공간이 2배 이상일 때)
- **10-100GB**: 항상 스트리밍
- **> 100GB**: 병렬/증분 백업

### Bitnami 차트 대응
- **MongoDB**: mongodump 포함 ✓
- **MinIO**: mc 미포함 → 자동 주입
- **PostgreSQL**: pg_dump 포함 ✓
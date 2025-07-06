# cli-restore

Kubernetes 환경을 위한 통합 백업/복원 도구. Pod 파일시스템, 데이터베이스, 오브젝트 스토리지를 지원합니다.

## 설치

### 바이너리 다운로드 (권장)

[Releases](https://github.com/cagojeiger/cli-recover/releases) 페이지에서 플랫폼에 맞는 바이너리를 다운로드하세요.

```bash
# 표준 버전 (온라인 환경)
wget https://github.com/cagojeiger/cli-recover/releases/latest/download/cli-restore-$(uname -s)-$(uname -m)
chmod +x cli-restore-*
sudo mv cli-restore-* /usr/local/bin/cli-restore

# 오프라인 버전 (mc 포함)
wget https://github.com/cagojeiger/cli-recover/releases/latest/download/cli-restore-offline-$(uname -s)-$(uname -m)
```

### 소스에서 빌드

```bash
git clone https://github.com/cagojeiger/cli-recover.git
cd cli-recover
make build
```

## 사용법

### 기본 명령어
```bash
cli-restore --version    # 버전 확인
cli-restore --help       # 도움말
```

### 대화형 모드 (TUI) - 권장
```bash
cli-restore              # k9s 스타일 풀스크린 TUI 실행
```

### 명령어 패턴
```bash
cli-restore [action] [target] [resource] [options]
```

### 백업 예시

#### Pod 파일시스템
```bash
cli-restore backup pod nginx-app /data --namespace prod --split-size 1G
```

#### MongoDB
```bash
# 자동 스트리밍 (대용량 안전)
cli-restore backup mongodb mongo-primary --all-databases

# 특정 데이터베이스
cli-restore backup mongodb mongo-primary --database myapp,sessions
```

#### MinIO
```bash
# mc가 없어도 자동 처리
cli-restore backup minio minio-server my-bucket --recursive
```

### 복원 예시
```bash
cli-restore restore pod ./backup-20240107.tar nginx-app
cli-restore restore mongodb ./dump.gz mongo-primary --drop
```

### 주요 기능
- **자동 전략 선택**: 데이터 크기에 따라 최적 백업 방법 자동 선택
- **Port Forward 관리**: 필요시 자동으로 포트 포워딩 설정
- **오프라인 지원**: 인터넷 없는 환경에서도 작동 (오프라인 빌드)
- **Bitnami 차트 호환**: Bitnami MongoDB, MinIO 등 자동 인식

### 필수 요구사항
- `kubectl` 설치 및 클러스터 연결
- 백업 대상에 대한 접근 권한

## 개발

```bash
# 개발 환경 설정
go mod download
make test

# 릴리즈 생성
git tag v0.1.1
git push origin v0.1.1
```

자세한 개발 가이드는 `make help`를 확인하세요.
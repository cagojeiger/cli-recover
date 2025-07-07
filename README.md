# cli-recover

Kubernetes 환경을 위한 백업/복원 도구. 현재 Pod 파일시스템 백업을 지원하며, 데이터베이스와 오브젝트 스토리지 지원은 향후 추가 예정입니다.

## 설치

### 바이너리 다운로드 (권장)

[Releases](https://github.com/cagojeiger/cli-recover/releases) 페이지에서 플랫폼에 맞는 바이너리를 다운로드하세요.

```bash
# 표준 버전 (온라인 환경)
wget https://github.com/cagojeiger/cli-recover/releases/latest/download/cli-recover-$(uname -s)-$(uname -m)
chmod +x cli-recover-*
sudo mv cli-recover-* /usr/local/bin/cli-recover

# 오프라인 버전 (mc 포함)
wget https://github.com/cagojeiger/cli-recover/releases/latest/download/cli-recover-offline-$(uname -s)-$(uname -m)
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
cli-recover --version    # 버전 확인
cli-recover --help       # 도움말
```

### 대화형 모드 (TUI)
```bash
cli-recover              # 간단한 텍스트 기반 TUI 실행
```

### 명령어 구조
```bash
cli-recover [command] [subcommand] [arguments] [flags]
```

### 백업 예시

#### Pod 파일시스템
```bash
# 기본 백업 (gzip 압축)
cli-recover backup filesystem nginx-app /data --namespace prod

# 압축 옵션 지정
cli-recover backup filesystem nginx-app /data --compression bzip2

# 특정 파일 제외
cli-recover backup filesystem nginx-app /data --exclude "*.log" --exclude "*.tmp"

# 출력 파일 지정
cli-recover backup filesystem nginx-app /data -o backup-nginx.tar.gz
```

### 복원 예시
```bash
# 파일시스템 복원
cli-recover restore filesystem backup-20240107.tar.gz nginx-app --namespace prod

# 특정 경로로 복원
cli-recover restore filesystem backup.tar.gz nginx-app --target-path /restore
```

### 주요 기능
- **파일시스템 백업**: Pod 내부 파일/디렉토리를 tar 아카이브로 백업
- **다양한 압축 지원**: gzip, bzip2, xz 압축 옵션
- **진행률 표시**: 실시간 백업 진행 상황 모니터링
- **간편한 TUI**: 대화형 인터페이스로 쉬운 백업 작업

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
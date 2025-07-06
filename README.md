# cli-recover

Kubernetes Pod 파일/폴더 백업 도구

## 설치

### 바이너리 다운로드 (권장)

[Releases](https://github.com/cagojeiger/cli-recover/releases) 페이지에서 플랫폼에 맞는 바이너리를 다운로드하세요.

```bash
# 다운로드 후 실행 권한 추가
chmod +x cli-restore
sudo mv cli-restore /usr/local/bin/
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

### 대화형 백업 (TUI)
```bash
cli-restore tui          # 대화형 인터페이스로 백업 설정
```

### 직접 백업 (CLI)
```bash
cli-restore backup <pod> <path> [flags]

# 예시
cli-restore backup my-app-pod /data
cli-restore backup my-app-pod /data --namespace production --split-size 2G
```

### 플래그 옵션
- `--namespace, -n`: Kubernetes 네임스페이스 (기본값: default)
- `--split-size, -s`: 분할 크기 (기본값: 1G)
- `--output, -o`: 출력 디렉토리 (기본값: ./backup)

### 필수 요구사항
- `kubectl` 설치 및 클러스터 연결
- 백업 대상 Pod에 대한 접근 권한

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
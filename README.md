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

```bash
cli-restore --version    # 버전 확인
cli-restore --help       # 도움말
```

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
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

### 초기 설정
```bash
cli-recover init         # 설정 파일 생성 (~/.cli-recover/config.yaml)
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

### 백업 목록 조회
```bash
# 모든 백업 목록
cli-recover list backups

# 특정 네임스페이스
cli-recover list backups --namespace prod
```

### 주요 기능
- **파일시스템 백업/복원**: Pod 내부 파일/디렉토리를 tar 아카이브로 백업 및 복원
- **다양한 압축 지원**: gzip, bzip2, xz 압축 옵션
- **진행률 표시**: 실시간 백업 진행 상황 모니터링
- **메타데이터 관리**: 백업 정보 자동 저장 및 조회
- **구조화된 로깅**: 파일/콘솔 로그 with 로테이션

### 필수 요구사항
- `kubectl` 설치 및 클러스터 연결
- 백업 대상에 대한 접근 권한

## 아키텍처

프로젝트는 헥사고날 아키텍처(Ports & Adapters)를 따릅니다:

```
internal/
├── domain/              # 비즈니스 로직
│   ├── backup/         # 백업 도메인
│   ├── restore/        # 복원 도메인
│   ├── metadata/       # 메타데이터 관리
│   └── logger/         # 로깅 인터페이스
├── infrastructure/      # 외부 시스템 연동
│   ├── kubernetes/     # K8s 클라이언트
│   ├── logger/         # 로거 구현체
│   ├── providers/      # 백업 프로바이더
│   └── runner/         # 명령 실행기
└── application/        # 애플리케이션 서비스
    ├── adapters/       # CLI 어댑터
    └── config/         # 설정 관리
```

## 개발

```bash
# 개발 환경 설정
go mod download
make test

# 테스트 커버리지
make test-coverage

# 릴리즈 생성
git tag v0.3.0
git push origin v0.3.0
```

자세한 개발 가이드는 `make help`를 확인하세요.

## 로드맵

- [x] Phase 1: CLI 핵심 기능 (filesystem 백업/복원)
- [x] Phase 2: 아키텍처 고도화 (hexagonal architecture)
- [ ] Phase 3: CLI 고도화 (백그라운드 실행, 작업 관리)
- [ ] Phase 4: TUI 재구현
- [ ] Phase 5: Provider 확장 (MinIO, MongoDB)
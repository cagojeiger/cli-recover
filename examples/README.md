# CLI-Pipe Examples

이 디렉토리에는 cli-pipe를 사용한 다양한 파이프라인 예시들이 포함되어 있습니다.

## 기본 예시들

### 🚀 hello-world.yaml
가장 간단한 "Hello, World!" 예시
```bash
cli-pipe run hello-world.yaml
```

### 📝 word-count.yaml
텍스트 파일의 단어 수를 세는 예시
```bash
cli-pipe run word-count.yaml
```

### 📄 file-processing.yaml
파일 처리 파이프라인 예시
```bash
cli-pipe run file-processing.yaml
```

### 🕐 date-time.yaml
날짜와 시간 처리 예시
```bash
cli-pipe run date-time.yaml
```

## Kubernetes 백업 예시들

### 📦 kubectl-backup-tmp.yaml
**vscode 파드의 /tmp 디렉토리 백업**
```bash
cli-pipe run kubectl-backup-tmp.yaml
```
- kubectl exec를 사용하여 파드 내부의 /tmp 디렉토리를 tar.gz로 압축
- 타임스탬프가 포함된 파일명으로 로컬에 저장
- 실제 운영 환경에서 바로 사용 가능

### 🏗️ kubectl-backup-advanced.yaml
**다중 디렉토리 종합 백업**
```bash
cli-pipe run kubectl-backup-advanced.yaml
```
- /tmp, /etc 설정파일, /var/log 최근 로그를 각각 별도 파일로 백업
- 시스템 중요 설정 파일들 (/etc/passwd, /etc/hosts 등) 포함
- 적당한 크기로 빠른 백업 실행

### 🌊 kubectl-streaming-backup.yaml
**스트리밍 백업 with 진행률 모니터링**
```bash
cli-pipe run kubectl-streaming-backup.yaml
```
- /tmp 디렉토리 스트리밍 백업 (50MB 예상 크기)
- `pv` 명령어로 진행률 표시
- GPG를 사용한 암호화 백업
- 특정 파일 타입만 선별 백업 (/usr/local에서 코드 파일들)

### 🎯 kubectl-pod-specific-backup.yaml
**특정 파드 지정 백업**
```bash
cli-pipe run kubectl-pod-specific-backup.yaml
```
- 네임스페이스와 파드명을 정확히 지정
- 동적으로 파드 이름 조회하여 /var/log 백업
- 시스템 설정 파일들 (/etc/hostname, /etc/hosts 등) 백업

### 🪶 kubectl-lightweight-backup.yaml
**가벼운 파일들만 선별 백업**
```bash
cli-pipe run kubectl-lightweight-backup.yaml
```
- /tmp에서 로그/임시파일 제외하고 깨끗한 백업
- 시스템 설정 파일들만 선별 백업
- 최근 1일간의 작은 로그 파일들만 백업
- 시스템 정보 파일들 (/proc/version, /proc/cpuinfo 등) 백업

### 🗄️ kubectl-app-backup.yaml
**애플리케이션별 전문 백업**
```bash
cli-pipe run kubectl-app-backup.yaml
```
- PostgreSQL 데이터베이스 덤프
- Redis 데이터 백업
- Kubernetes ConfigMaps/Secrets 백업
- 퍼시스턴트 볼륨 백업

## 로깅 시스템

모든 파이프라인 실행은 다음과 같이 자동으로 로깅됩니다:

```
~/.cli-pipe/logs/
├── kubectl-backup-tmp_20240714_120000/     # 파이프라인명_실행시간
│   ├── pipeline.log                         # 모든 출력 (stdout)
│   ├── stderr.log                           # 에러 출력
│   └── summary.txt                          # 실행 요약
```

### 로그 특징
- **투명한 스트리밍**: 콘솔에서 실시간 출력을 보면서 동시에 로그 파일에 자동 저장
- **바이너리 안전**: tar 압축 데이터도 안전하게 로깅
- **자동 정리**: retention_days 설정에 따라 오래된 로그 자동 삭제
- **복구 가능**: 네트워크 중단 시 pipeline.log에서 부분 복구 가능

## 사용법

1. **파이프라인 실행**:
   ```bash
   cli-pipe run <example-file>.yaml
   ```

2. **실시간 모니터링**:
   - 콘솔에서 실행 상황 실시간 확인
   - 처리된 바이트 수, 전송 속도 등 자동 표시

3. **로그 확인**:
   ```bash
   ls ~/.cli-pipe/logs/
   cat ~/.cli-pipe/logs/kubectl-backup-tmp_*/summary.txt
   ```

## 주의사항

### Kubernetes 백업 시
- **권한**: kubectl이 설정되어 있고 파드에 접근 권한이 있어야 함
- **네트워크**: 파드와의 안정적인 네트워크 연결 필요
- **디스크 공간**: 백업 파일 크기만큼 로컬 디스크 공간 필요
- **보안**: 민감한 데이터 백업 시 암호화 옵션 사용 권장

### 성능 최적화
- 대용량 백업 시 `pv` 명령어로 진행률 모니터링
- 불필요한 파일 제외로 백업 크기 최소화
- 네트워크 대역폭 고려하여 압축 레벨 조정

## 확장 가능성

이 예시들을 기반으로 다음과 같은 커스텀 백업 파이프라인을 만들 수 있습니다:

- 특정 애플리케이션 백업 스케줄링
- 다중 클러스터 백업
- 백업 파일 자동 업로드 (S3, GCS 등)
- 백업 무결성 검증
- 알림 시스템과 연동
# CLI Phase 1 스프린트

## 스프린트 정보
- **시작일**: 2025-01-07
- **종료일**: 2025-01-14 (1주)
- **목표**: CLI 핵심 기능 완성

## 스프린트 목표
- 3가지 백업 타입 (filesystem, minio, mongodb) CLI 지원
- 표준화된 명령 체계 구축
- 일관된 진행률 및 에러 처리

## 작업 항목

### 1. CLI 명령 체계 표준화 (1일)
- [ ] cobra 또는 urfave/cli 프레임워크 선택
- [ ] 명령 구조 설계
  ```bash
  cli-recover backup <type> <pod> <path> [options]
  cli-recover restore <type> <pod> <backup-file> [options]
  cli-recover list backups
  cli-recover status <job-id>
  ```
- [ ] 공통 플래그 정의 (--namespace, --output, --verbose 등)
- [ ] 도움말 시스템 구현

### 2. Filesystem Provider 안정화 (1일)
- [ ] 현재 backup_filesystem.go 리팩토링
- [ ] BackupProvider 인터페이스 정의
- [ ] 진행률 스트리밍 개선
- [ ] 에러 처리 표준화
- [ ] 단위 테스트 작성

### 3. MinIO Provider 구현 (2일)
- [ ] MinIO BackupProvider 구조체 구현
- [ ] S3 명령어 빌더 (`mc mirror` 또는 `aws s3 sync`)
- [ ] 크기 추정 로직
- [ ] 진행률 파싱
- [ ] MinIO 연결 테스트
- [ ] 통합 테스트

### 4. MongoDB Provider 구현 (2일)
- [ ] MongoDB BackupProvider 구조체 구현
- [ ] mongodump 명령어 빌더
- [ ] 컬렉션별 진행률 추적
- [ ] 압축 옵션 지원
- [ ] MongoDB 연결 테스트
- [ ] 통합 테스트

### 5. 공통 기능 구현 (1일)
- [ ] 통합 진행률 인터페이스
  ```go
  type Progress struct {
      Current   int64
      Total     int64
      Speed     float64
      ETA       time.Duration
      Message   string
  }
  ```
- [ ] 에러 타입 정의 및 핸들링
- [ ] 로깅 시스템 (logrus 또는 zap)
- [ ] 설정 파일 지원 기초

## 일일 작업 계획

### Day 1 (화)
- CLI 프레임워크 선택 및 기본 구조 구현
- 명령어 라우팅 시스템 구축

### Day 2 (수)
- Filesystem provider 리팩토링
- BackupProvider 인터페이스 확정

### Day 3-4 (목-금)
- MinIO provider 구현
- S3 통합 테스트

### Day 5-6 (월-화)
- MongoDB provider 구현
- mongodump 통합 테스트

### Day 7 (수)
- 공통 기능 통합
- 전체 테스트 및 문서화

## BackupProvider 인터페이스

```go
// internal/domain/backup/provider.go
package backup

import (
    "context"
    "io"
)

type Provider interface {
    // 기본 정보
    Name() string
    Description() string
    
    // 백업 실행
    Execute(ctx context.Context, opts Options) error
    
    // 크기 추정
    EstimateSize(opts Options) (int64, error)
    
    // 진행률 스트림
    StreamProgress() <-chan Progress
    
    // 옵션 검증
    ValidateOptions(opts Options) error
}

type Options struct {
    Namespace  string
    PodName    string
    SourcePath string
    OutputFile string
    Compress   bool
    Exclude    []string
    // Provider별 추가 옵션
    Extra      map[string]interface{}
}

type Progress struct {
    Current   int64
    Total     int64
    Speed     float64  // bytes per second
    ETA       string
    Message   string
}
```

## 디렉토리 구조

```
cmd/cli-recover/
├── main.go
├── commands/
│   ├── backup.go
│   ├── restore.go
│   ├── list.go
│   └── status.go
└── handlers/
    ├── filesystem.go
    ├── minio.go
    └── mongodb.go

internal/
├── domain/
│   └── backup/
│       ├── provider.go      # 인터페이스
│       ├── options.go       # 옵션 구조체
│       └── progress.go      # 진행률 타입
├── providers/
│   ├── filesystem/
│   │   └── filesystem.go
│   ├── minio/
│   │   └── minio.go
│   └── mongodb/
│       └── mongodb.go
└── kubernetes/
    └── client.go            # kubectl 래퍼
```

## 성공 지표
- [ ] 3가지 백업 타입 모두 CLI로 실행 가능
- [ ] 각 provider별 단위 테스트 작성
- [ ] 통합 테스트 통과
- [ ] 일관된 진행률 표시
- [ ] 표준화된 에러 메시지
- [ ] 기본 사용 문서 작성

## 위험 요소
1. **MinIO/MongoDB 환경 설정**
   - Docker Compose로 테스트 환경 구축
   - Kind 클러스터에 테스트 Pod 배포

2. **진행률 파싱 복잡도**
   - 각 도구마다 다른 출력 형식
   - 정규식 기반 파싱 필요

3. **에러 처리 일관성**
   - 각 provider별 에러 타입 통일
   - 사용자 친화적 메시지 변환

## 테스트 계획

### 단위 테스트
```go
// filesystem_test.go
func TestFilesystemProvider_EstimateSize(t *testing.T) {
    // kubectl exec로 du 명령 실행 테스트
}

func TestFilesystemProvider_Execute(t *testing.T) {
    // tar 명령 실행 및 진행률 파싱 테스트
}
```

### 통합 테스트
```bash
# 테스트 환경 설정
kind create cluster --name test-cluster
kubectl apply -f test/fixtures/test-pods.yaml

# CLI 테스트 실행
go test ./test/integration/...
```

## 참고사항
- 기존 filesystem 백업이 계속 동작하도록 주의
- 과도한 추상화 지양 (YAGNI 원칙)
- 실용적이고 단순한 구현 우선

## 다음 스프린트 예고
- Phase 2: 아키텍처 정리
  - 도메인 레이어 분리
  - 의존성 주입
  - 플러그인 레지스트리
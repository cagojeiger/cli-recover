# Architecture Insights - 현재 구조의 한계와 개선 방향

## 발견된 문제점

### 1. 과도한 추상화
```go
// 현재: 모든 Provider가 같은 인터페이스 강제
type Provider interface {
    Execute(ctx context.Context, opts Options) error
    EstimateSize(opts Options) (int64, error)
    StreamProgress() <-chan Progress
}
```

**문제점**:
- Filesystem: `EstimateSize`는 `du -s`로 간단
- MongoDB: 실제 크기는 dump 후에나 알 수 있음
- MinIO: 객체 메타데이터로 즉시 계산 가능

각 Provider마다 동작이 완전히 다른데 같은 인터페이스로 강제하니 불필요한 복잡도 발생.

### 2. Restore ≠ Backup의 역연산

**현재 가정**: Restore는 Backup의 반대
**실제 발견**:
- Filesystem backup: `tar -czf` (압축 아카이브 생성)
- Filesystem restore: `tar -xzf` (실제로는 `cp -r`에 가까움)
- 사용자는 보통 특정 파일만 복원하고 싶어함

### 3. 명령어 구조의 경직성

```bash
# 현재
cli-recover backup filesystem <pod> <path> -o backup.tar
cli-recover restore filesystem <pod> backup.tar

# 실제 사용자 니즈
cli-recover cp <pod>:<path> backup.tar     # kubectl cp 스타일
cli-recover sync <pod>:<path> <pod>:<path> # rsync 스타일
```

### 4. 에러 처리의 일관성 부족

```go
// 현재: 에러 발생 시 전체 help 출력
Error: restore failed: tar: File exists
[전체 help 메시지 200줄...]

// 개선: 상황별 맞춤 가이드
Error: Files already exist at target
Fix: Use --force to overwrite
Example: cli-recover restore ... --force
```

## 아키텍처 개선 방향

### 1. Provider 특화 구조
```
before/
├── domain/
│   ├── backup/
│   │   └── provider.go  # 공통 인터페이스 강제
│   └── restore/
│       └── provider.go  # 동일한 패턴

after/
├── providers/
│   ├── filesystem/      # 독립적 구현
│   ├── mongodb/         # 독립적 구현
│   └── shared/          # 진짜 공통 부분만
```

### 2. 동작 중심 명령어 설계

```bash
# Filesystem
cli-recover fs backup <pod>:<path> -o backup.tar
cli-recover fs restore <pod>:<path> backup.tar
cli-recover fs copy <src-pod>:<path> <dst-pod>:<path>
cli-recover fs sync <pod>:<path> local-dir/

# MongoDB
cli-recover mongo dump <pod> -o backup/
cli-recover mongo restore <pod> backup/
cli-recover mongo export <pod> --collection users

# MinIO
cli-recover s3 mirror <pod> local-dir/
cli-recover s3 sync <bucket> backup/
```

### 3. 진행률 표시 개선

**현재**: 모든 Provider가 동일한 Progress 구조체
```go
type Progress struct {
    Current int64
    Total   int64
    Message string
}
```

**개선**: Provider별 특화
```go
// Filesystem
type TarProgress struct {
    FilesProcessed int
    CurrentFile    string
    BytesWritten   int64
}

// MongoDB
type MongoProgress struct {
    Collections    int
    CurrentColl    string
    Documents      int64
}
```

### 4. CLI/TUI 통합 레이어

```go
// 메타데이터 기반 UI 자동 생성
type CommandMetadata struct {
    Provider    string
    Operation   string
    Arguments   []Argument
    Flags       []Flag
    Examples    []Example
}

// CLI와 TUI가 같은 메타데이터 사용
func (m CommandMetadata) GenerateCLI() *cobra.Command
func (m CommandMetadata) GenerateTUI() *tview.Form
```

## 코드 재사용 전략

### 유지할 것
- ✅ 로깅 시스템 (잘 설계됨)
- ✅ 진행률 리포터 패턴 (유연함)
- ✅ 에러 구조화 (CLIError)
- ✅ kubectl 클라이언트 래핑

### 제거할 것
- ❌ 공통 Provider 인터페이스
- ❌ Options 구조체 (Provider별로 다름)
- ❌ 일반화된 Result 구조체

### 재설계할 것
- ⚠️ 명령어 구조 (동작 중심으로)
- ⚠️ 설정 시스템 (Provider별 설정)
- ⚠️ 메타데이터 저장 (Provider 특화)

## 성능 고려사항

### 1. 스트리밍 최적화
- 현재: 모든 데이터를 메모리에 버퍼링
- 개선: Provider별 최적 스트리밍
  - Filesystem: pipe 직접 연결
  - MongoDB: cursor 기반 스트리밍
  - MinIO: multipart upload

### 2. 병렬 처리
- Filesystem: 대용량 파일은 분할 백업
- MongoDB: 컬렉션별 병렬 덤프
- MinIO: 동시 다중 객체 전송

## 보안 고려사항

### 1. 권한 분리
- Provider별 필요 권한만 요청
- Filesystem: pod exec 권한
- MongoDB: DB 접근 권한
- MinIO: S3 API 권한

### 2. 암호화
- 전송 중 암호화 (TLS)
- 저장 시 암호화 (옵션)
- Provider별 암호화 방식

## 결론

현재 아키텍처는 "모든 것을 일반화"하려다가 
오히려 각 Provider의 장점을 살리지 못하고 있습니다.

Provider별 특성을 인정하고 독립적으로 구현하되,
사용자 경험은 CLI/TUI 통합 레이어로 일관성을 유지하는 것이
더 나은 방향입니다.
# Provider Analysis - 각 백업 타입의 실제 동작 분석

## 1. Filesystem Provider

### 실제 사용 패턴
```bash
# 현재 구현
kubectl exec -n <namespace> <pod> -- tar -czf - <path> > backup.tar.gz
kubectl exec -i -n <namespace> <pod> -- tar -xzf - -C <path> < backup.tar.gz

# 실제 사용자 니즈
kubectl cp <pod>:<path> ./backup/        # 단순 복사
kubectl exec <pod> -- rsync -av <src> <dst>  # 동기화
```

### 특성 분석
- **백업**: 아카이브 생성 (tar)
- **복원**: 파일 복사 (cp)에 가까움
- **특이사항**: 
  - 심볼릭 링크 처리
  - 권한 보존 옵션
  - 증분 백업 가능

### 구현 고려사항
```go
type FilesystemProvider struct {
    // tar 옵션
    Compression   string   // none, gzip, bzip2, xz
    PreservePerms bool
    FollowSymlinks bool
    
    // 제외 패턴
    ExcludePatterns []string
    
    // 증분 백업
    SnapshotFile string
}
```

## 2. MongoDB Provider

### 실제 사용 패턴
```bash
# 전체 백업
mongodump --host=<host> --port=<port> --out=/backup

# 특정 DB/컬렉션
mongodump --db=mydb --collection=users --out=/backup

# 복원
mongorestore --host=<host> --drop /backup

# 실시간 백업 (oplog)
mongodump --oplog --out=/backup
```

### 특성 분석
- **백업 형식**: BSON + 메타데이터
- **특수 기능**:
  - Oplog 기반 point-in-time 복구
  - 샤드 클러스터 지원
  - 인덱스 재생성 옵션

### 구현 고려사항
```go
type MongoDBProvider struct {
    // 연결 정보
    URI      string
    AuthDB   string
    Username string
    Password string // from secret
    
    // 백업 옵션
    IncludeOplog     bool
    ExcludeCollections []string
    
    // 복원 옵션
    Drop            bool  // 기존 데이터 삭제
    RestoreIndexes  bool
    OplogReplay     bool
}
```

### MongoDB 특화 진행률
```go
type MongoProgress struct {
    Phase         string // "scanning", "dumping", "building-indexes"
    Database      string
    Collection    string
    DocumentCount int64
    BytesWritten  int64
}
```

## 3. MinIO/S3 Provider

### 실제 사용 패턴
```bash
# MinIO Client (mc)
mc mirror myalias/bucket /backup/bucket
mc cp --recursive myalias/bucket/path /backup/

# AWS CLI
aws s3 sync s3://bucket /backup/bucket
aws s3 cp s3://bucket/object /backup/object

# 버전 관리
mc cp --rewind 7d myalias/bucket/object old-version.dat
```

### 특성 분석
- **객체 스토리지**: 파일시스템과 다른 개념
- **특수 기능**:
  - 버전 관리
  - 멀티파트 업로드
  - 서버 사이드 암호화
  - 라이프사이클 정책

### 구현 고려사항
```go
type MinIOProvider struct {
    // 연결 정보
    Endpoint        string
    AccessKey       string
    SecretKey       string // from secret
    UseSSL          bool
    
    // 백업 옵션
    IncludeVersions bool
    IncludeMetadata bool
    
    // 성능 옵션
    Concurrency     int  // 동시 전송 수
    PartSize        int64 // 멀티파트 크기
}
```

### S3 특화 진행률
```go
type S3Progress struct {
    Bucket         string
    CurrentObject  string
    ObjectsTotal   int
    ObjectsDone    int
    BytesTotal     int64
    BytesTransferred int64
    TransferSpeed  float64 // MB/s
}
```

## 4. PostgreSQL Provider (향후)

### 실제 사용 패턴
```bash
# SQL 덤프
pg_dump -h host -U user -d dbname > backup.sql
pg_dump -Fc -h host -U user -d dbname > backup.dump  # custom format

# 복원
psql -h host -U user -d dbname < backup.sql
pg_restore -h host -U user -d dbname backup.dump

# 특정 테이블만
pg_dump -t table1 -t table2 dbname > tables.sql
```

### 특성 분석
- **백업 형식**: 
  - Plain SQL (텍스트)
  - Custom format (압축)
  - Directory format (병렬)
- **특수 기능**:
  - 스키마만 백업
  - 데이터만 백업
  - 병렬 백업/복원

## 5. Redis Provider (향후)

### 실제 사용 패턴
```bash
# RDB 스냅샷
redis-cli BGSAVE
cp /var/lib/redis/dump.rdb /backup/

# AOF 백업
cp /var/lib/redis/appendonly.aof /backup/

# 실시간 복제
redis-cli --rdb /backup/dump.rdb
```

### 특성 분석
- **백업 방식**:
  - RDB: 시점 스냅샷
  - AOF: 명령어 로그
- **특수 고려사항**:
  - 메모리 사용량
  - 백업 중 성능 영향

## Provider별 공통점과 차이점

### 공통점
1. 진행률 리포팅 필요
2. 에러 처리 및 재시도
3. 로깅 및 감사
4. 보안 (인증/암호화)

### 차이점

| 특성 | Filesystem | MongoDB | MinIO/S3 | PostgreSQL |
|------|------------|---------|----------|------------|
| 백업 단위 | 파일/디렉토리 | DB/컬렉션 | 버킷/객체 | DB/스키마/테이블 |
| 스트리밍 | tar pipe | BSON stream | HTTP stream | SQL stream |
| 메타데이터 | 파일 속성 | 인덱스 정의 | 객체 태그 | 스키마 정의 |
| 증분 백업 | tar --listed | oplog | 버전 비교 | WAL 아카이브 |
| 병렬 처리 | 파일 단위 | 컬렉션 단위 | 객체 단위 | 테이블 단위 |

## 구현 우선순위

### 1단계: Filesystem 개선
- 현재 구현을 Provider 독립 구조로
- cp/sync 스타일 명령어 추가

### 2단계: MongoDB 추가
- mongodump/restore 래핑
- Kubernetes 환경 최적화

### 3단계: MinIO/S3 추가  
- mc/aws-cli 통합
- 멀티파트 업로드 지원

### 4단계: PostgreSQL/Redis
- 수요에 따라 우선순위 결정

## 테스트 전략

### Provider별 독립 테스트
```go
// filesystem_test.go
func TestFilesystemBackup(t *testing.T)
func TestFilesystemRestore(t *testing.T)

// mongodb_test.go  
func TestMongoDBDump(t *testing.T)
func TestMongoDBRestore(t *testing.T)
```

### 통합 테스트는 최소화
- 실제 Kubernetes 클러스터 필요
- CI/CD에서 특정 Provider만 테스트

## 결론

각 Provider는 완전히 다른 특성을 가지고 있으며,
하나의 인터페이스로 통합하는 것은 오히려 제약이 됩니다.

Provider별 독립 구현으로 각각의 장점을 최대한 활용하고,
사용자 경험은 CLI/TUI 레이어에서 통합하는 것이 최선입니다.
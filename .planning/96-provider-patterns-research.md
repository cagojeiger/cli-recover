# Provider Usage Patterns Research

## 1. Filesystem (tar) Patterns

### 기본 백업 옵션
```bash
# 압축 타입별 백업
tar -czf backup.tar.gz /data     # gzip (빠름, 적당한 압축)
tar -cjf backup.tar.bz2 /data    # bzip2 (느림, 높은 압축)
tar -cJf backup.tar.xz /data     # xz (매우 느림, 최고 압축)

# Kubernetes Pod에서 스트리밍
kubectl exec -n prod app-pod -- tar -czf - /data > backup.tar.gz
```

### 증분 백업
```bash
# Level 0 (전체 백업)
tar -czf backup-full.tar.gz --listed-incremental=snapshot.snar /data

# Level 1 (증분 백업)
cp snapshot.snar snapshot.snar.1
tar -czf backup-incr-1.tar.gz --listed-incremental=snapshot.snar.1 /data

# 복원 (순서대로)
tar -xzf backup-full.tar.gz --listed-incremental=/dev/null
tar -xzf backup-incr-1.tar.gz --listed-incremental=/dev/null
```

### 진행률 및 제외
```bash
# 진행률 표시
tar -cvzf backup.tar.gz /data          # 파일별 출력
tar -czf backup.tar.gz --totals /data  # 전체 통계

# 제외 패턴
tar -czf backup.tar.gz \
  --exclude='*.log' \
  --exclude='node_modules' \
  --exclude='__pycache__' \
  /data
```

## 2. MongoDB Patterns

### 기본 백업
```bash
# 전체 백업
mongodump --uri="mongodb://user:pass@host:27017/db?authSource=admin"

# 특정 DB/컬렉션
mongodump --db=myapp --collection=users --out=/backup

# 압축 및 아카이브
mongodump --archive=/backup/mongo.archive --gzip
```

### Point-in-Time 백업
```bash
# Oplog 포함 (Replica Set 필수)
mongodump --oplog --out=/backup

# 복원 시 oplog 재생
mongorestore --oplogReplay /backup
```

### Kubernetes 환경
```bash
# Pod 내부에서 백업
kubectl exec -n prod mongo-0 -- mongodump \
  --uri="mongodb://localhost:27017" \
  --archive --gzip | gzip > mongo-backup.gz

# TLS/SSL 사용
kubectl exec -n prod mongo-0 -- mongodump \
  --ssl --sslCAFile=/certs/ca.pem \
  --uri="mongodb://host:27017"
```

## 3. MinIO/S3 Patterns

### mc 명령어 차이점
```bash
# cp: 단순 복사 (메타데이터 미포함)
mc cp source/file.txt dest/

# mirror: 단방향 동기화 (삭제 포함)
mc mirror --overwrite source/ dest/

# sync: 변경된 파일만 복사 (삭제 안함)
mc sync source/ dest/

# replicate: 실시간 복제 설정
mc replicate add source/ --remote-bucket dest/
```

### 버전 관리
```bash
# 버전 활성화
mc version enable myminio/bucket

# 특정 버전 복구
mc cp --version-id=<id> myminio/bucket/file restored-file

# 7일 전 버전으로
mc cp --rewind 7d myminio/bucket/file old-version
```

### 성능 최적화
```bash
# 대역폭 제한
mc cp --limit-upload 10MB --limit-download 5MB

# 멀티파트 설정
export MC_UPLOAD_SIZE=64MiB    # 파트 크기
export MC_CONCURRENT=8         # 동시 업로드 수

# Watch 모드 (실시간 동기화)
mc watch --recursive myminio/bucket
```

### Kubernetes Job 예제
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: minio-backup
spec:
  template:
    spec:
      containers:
      - name: backup
        image: minio/mc:latest
        command:
        - sh
        - -c
        - |
          mc config host add myminio $MINIO_ENDPOINT $ACCESS_KEY $SECRET_KEY
          mc mirror myminio/prod-data /backup/
        env:
        - name: MINIO_ENDPOINT
          value: "https://minio.example.com"
        volumeMounts:
        - name: backup
          mountPath: /backup
```

## 4. PostgreSQL Patterns

### 출력 형식별 특징
```bash
# Plain SQL (텍스트, 파이프 가능)
pg_dump dbname > backup.sql
pg_dump dbname | gzip > backup.sql.gz

# Custom Format (압축, 병렬 복원 가능)
pg_dump -Fc dbname > backup.dump
pg_restore -j 4 backup.dump  # 4개 job으로 병렬 복원

# Directory Format (병렬 덤프/복원)
pg_dump -Fd -j 4 -f backup_dir dbname
pg_restore -j 4 backup_dir

# Tar Format (이식성)
pg_dump -Ft dbname > backup.tar
```

### 선택적 백업
```bash
# 스키마만
pg_dump --schema-only dbname > schema.sql

# 데이터만
pg_dump --data-only dbname > data.sql

# 특정 테이블
pg_dump -t users -t orders dbname > tables.sql

# 패턴 매칭
pg_dump -t 'public.user_*' dbname > user_tables.sql
```

### Kubernetes CronJob
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:14
            command:
            - sh
            - -c
            - |
              export PGPASSWORD=$POSTGRES_PASSWORD
              pg_dump -h $POSTGRES_HOST -U $POSTGRES_USER \
                -Fc -Z9 $POSTGRES_DB > /backup/db-$(date +%Y%m%d).dump
            env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: password
```

## 공통 패턴 및 고려사항

### 1. 스트리밍 vs 파일 저장
- **스트리밍**: 메모리 효율적, 중간 저장소 불필요
- **파일 저장**: 재시도 가능, 검증 용이

### 2. 압축 옵션
- **gzip**: 균형잡힌 선택 (속도/압축률)
- **bzip2**: 높은 압축률, 느린 속도
- **xz**: 최고 압축률, 매우 느림
- **없음**: 네트워크가 빠르거나 이미 압축된 데이터

### 3. 병렬 처리
- **tar**: 파일 단위 병렬화 어려움
- **PostgreSQL**: -j 옵션으로 테이블별 병렬
- **MongoDB**: 컬렉션별 병렬 가능
- **MinIO**: 객체별 병렬 전송

### 4. 증분/차등 백업
- **tar**: --listed-incremental
- **PostgreSQL**: WAL 아카이브
- **MongoDB**: oplog
- **MinIO**: 타임스탬프/버전 기반

### 5. Kubernetes 특화
- Init Container로 준비 작업
- ConfigMap으로 스크립트 관리
- Secret으로 인증 정보 관리
- PVC로 백업 저장소 마운트
- Resource Limits 설정 필수

## 구현 시 주의사항

### 1. 에러 처리
```bash
# 파이프 실패 감지
set -o pipefail

# 에러 시 즉시 종료
set -e

# 미정의 변수 사용 금지
set -u
```

### 2. 로깅
```bash
# 타임스탬프 포함
echo "[$(date -Iseconds)] Starting backup..."

# stderr 리다이렉션
exec 2>&1 | tee -a backup.log
```

### 3. 정리 작업
```bash
# trap으로 정리
trap 'rm -f $TMPFILE' EXIT

# 임시 파일 안전하게 생성
TMPFILE=$(mktemp)
```

### 4. 검증
```bash
# 백업 후 검증
tar -tzf backup.tar.gz > /dev/null
echo "Backup verified successfully"

# 체크섬
sha256sum backup.tar.gz > backup.tar.gz.sha256
```

이러한 패턴들을 CLI-Recover에 구현하면 
각 Provider의 특성을 최대한 활용하면서도
일관된 사용자 경험을 제공할 수 있습니다.
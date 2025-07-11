# Backup Tools Usage Patterns

This document provides comprehensive examples and common use cases for various backup tools, with a focus on Kubernetes pod contexts.

## 1. tar (Filesystem Backups)

### Common Options

- **Compression Types:**
  - `-czf`: gzip compression (most common, good balance)
  - `-cjf`: bzip2 compression (better compression, slower)
  - `-cJf`: xz compression (best compression, slowest)

### Incremental Backup with `--listed-incremental`

The `--listed-incremental` option uses a snapshot file to track file changes, enabling efficient incremental backups.

#### Basic Syntax
```bash
tar --create --listed-incremental=snapshot.snar --file=backup.tar [directory]
```

#### Level 0 (Full) Backup
```bash
# First full backup
tar --create --gzip --listed-incremental=/backup/data.snar --file=/backup/data0.tar.gz /data
```

#### Level 1 (Incremental) Backup
```bash
# Subsequent incremental backup (uses existing snapshot file)
tar --create --gzip --listed-incremental=/backup/data.snar --file=/backup/data1.tar.gz /data
```

### Kubernetes Pod Examples

#### Full Backup from Pod
```bash
kubectl exec -it my-pod -- tar -czf - /app/data > backup-full.tar.gz
```

#### Incremental Backup from Pod
```bash
# First, create initial full backup
kubectl exec -it my-pod -- tar -C /app/data \
  --create \
  --listed-incremental=/tmp/backup.snar \
  --gzip \
  --file=- . > backup-level0.tar.gz

# Subsequent incremental backups
kubectl exec -it my-pod -- tar -C /app/data \
  --create \
  --listed-incremental=/tmp/backup.snar \
  --gzip \
  --file=- . > backup-level1.tar.gz
```

### Exclude Patterns
```bash
tar --create \
  --gzip \
  --file=backup.tar.gz \
  --exclude='*.log' \
  --exclude='node_modules' \
  --exclude='.git' \
  /app/data
```

### Progress Display
```bash
# Verbose output
tar -czvf backup.tar.gz /data

# Show totals
tar --totals -czf backup.tar.gz /data
```

### Restoring Backups
```bash
# Restore full backup first
tar --extract --listed-incremental=/dev/null --file=backup-level0.tar.gz

# Then restore incremental backups in order
tar --extract --listed-incremental=/dev/null --file=backup-level1.tar.gz
```

## 2. mongodump (MongoDB Backups)

### Connection Options
- `--host`: MongoDB host
- `--port`: MongoDB port (default: 27017)
- `--authenticationDatabase`: Authentication database (usually "admin")
- `--username`, `--password`: Authentication credentials

### Basic Commands

#### Simple Backup from Kubernetes Pod
```bash
kubectl exec -it mongodb-pod -- mongodump --out /tmp/backup
kubectl cp mongodb-pod:/tmp/backup ./mongodb-backup
```

#### Backup with Authentication
```bash
kubectl exec -it mongodb-pod -- mongodump \
  --host=localhost:27017 \
  --authenticationDatabase=admin \
  --username=admin \
  --password=$ADMIN_PASSWORD \
  --out=/tmp/backup
```

### Selective Backup

#### Specific Database
```bash
mongodump --db=myapp --out=/backup
```

#### Specific Collection
```bash
mongodump --db=myapp --collection=users --out=/backup
```

### Point-in-Time with Oplog

The `--oplog` option captures all operations during the backup, ensuring consistency.

```bash
kubectl exec -it mongodb-pod -- mongodump \
  --oplog \
  --out=/tmp/backup \
  --host=rs0/mongodb:27017 \
  --username=$ADMIN_USER \
  --password=$ADMIN_PASSWORD \
  --authenticationDatabase=admin
```

### With SSL/TLS in Kubernetes
```bash
kubectl exec -it mongodb-pod -- bash -c '
  cat /certs/tls.crt /certs/tls.key > /tmp/mongo.pem
  mongodump \
    --oplog \
    --out=/tmp/backup \
    --host=mongodb:27017 \
    --username=$ADMIN_USER \
    --password=$ADMIN_PASSWORD \
    --authenticationDatabase=admin \
    --ssl \
    --sslCAFile=/certs/ca.crt \
    --sslPEMKeyFile=/tmp/mongo.pem
'
```

### Compression Options
```bash
# Default compression (gzip)
mongodump --gzip --out=/backup

# Archive format (single file)
mongodump --archive=/backup/mongodb.archive --gzip
```

### Restoring with Oplog Replay
```bash
mongorestore --oplogReplay --dir=/backup
```

## 3. MinIO mc Tool

### Key Commands Comparison

- **`mc cp`**: One-time copy, specific files, no version history
- **`mc mirror`**: Synchronization, includes `--watch` mode, current version only
- **`mc sync`**: Similar to mirror but with additional options
- **`mc replicate`**: Full backup with version history and metadata

### mc mirror Examples

#### Basic Mirror
```bash
mc mirror /local/data myminio/mybucket
```

#### Watch Mode (Continuous Sync)
```bash
mc mirror --watch --remove /local/data myminio/mybucket
```

#### With Bandwidth Limiting
```bash
mc mirror --limit-download "10MiB/s" /local/data myminio/mybucket
```

### mc cp Examples

#### Recursive Copy
```bash
mc cp --recursive ~/mydata/ myminio/mybucket/
```

#### Copy with Older Versions
```bash
mc cp --version-id=3deb0b96-68df-4961-8c35-506a3f18e8ae myminio/mybucket/file.txt ./
```

### Versioning Support

#### Enable Versioning
```bash
mc version enable myminio/mybucket
```

#### List Versions
```bash
mc ls --versions myminio/mybucket
```

### Replication for True Backup

#### Setup Replication with Full Options
```bash
mc replicate add \
  --remote-bucket https://user:secret@backup-minio:9000/backup-bucket \
  --replicate "delete,delete-marker,existing-objects,metadata-sync" \
  --bandwidth "50MiB/s" \
  --limit-upload "25MiB/s" \
  --limit-download "25MiB/s" \
  myminio/mybucket
```

### Kubernetes Examples

#### Backup Job
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: minio-backup
spec:
  template:
    spec:
      containers:
      - name: mc-backup
        image: minio/mc:latest
        command: ["/bin/sh", "-c"]
        args:
        - |
          mc alias set source http://minio-source:9000 $ACCESS_KEY $SECRET_KEY
          mc alias set backup http://minio-backup:9000 $BACKUP_ACCESS_KEY $BACKUP_SECRET_KEY
          mc mirror --bandwidth "50MiB/s" source/data backup/data-$(date +%Y%m%d)
        env:
        - name: ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: minio-creds
              key: access-key
```

### Multipart Upload Configuration
```bash
# Set part size for large files
mc cp --part-size=100MiB largefile.zip myminio/mybucket/
```

## 4. pg_dump (PostgreSQL Backups)

### Output Formats

- **Plain SQL** (`-Fp` or default): Human-readable SQL script
- **Custom** (`-Fc`): Compressed, flexible format (recommended for single file)
- **Directory** (`-Fd`): Multiple files, supports parallel dumps
- **Tar** (`-Ft`): Tar archive format

### Basic Commands

#### Simple Backup from Kubernetes Pod
```bash
kubectl exec -it postgres-pod -- pg_dump -U postgres mydb > backup.sql
```

#### Custom Format with Compression
```bash
kubectl exec -it postgres-pod -- pg_dump -U postgres -d mydb -Fc -Z6 > backup.dump
```

### Parallel Dumps (Directory Format Only)

```bash
# Inside pod with 4 parallel jobs
kubectl exec -it postgres-pod -- pg_dump \
  -U postgres \
  -d mydb \
  -Fd \
  -j 4 \
  -f /tmp/dump_dir

# Copy directory to local
kubectl cp postgres-pod:/tmp/dump_dir ./pg_backup
```

### Compression Options

#### PostgreSQL 16+ Syntax
```bash
# Specify compression method and level
pg_dump -Fc --compress=gzip:9 -d mydb -f backup.dump
pg_dump -Fc --compress=lz4:1 -d mydb -f backup.dump
pg_dump -Fc --compress=zstd:3 -d mydb -f backup.dump
```

#### Legacy Syntax
```bash
# Compression level (0-9)
pg_dump -Fc -Z9 -d mydb -f backup.dump
```

### Schema-Only vs Data-Only

```bash
# Schema only
pg_dump --schema-only -d mydb > schema.sql

# Data only
pg_dump --data-only -d mydb > data.sql
```

### Table Selection Patterns

```bash
# Specific tables
pg_dump -t users -t orders -d mydb > tables.sql

# Pattern matching
pg_dump -t 'public.user_*' -d mydb > user_tables.sql

# Exclude tables
pg_dump -T logs -T temp_* -d mydb > backup.sql
```

### Kubernetes CronJob Example

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
          - name: pg-backup
            image: postgres:15
            command: ["/bin/bash", "-c"]
            args:
            - |
              DATE=$(date +%Y%m%d-%H%M%S)
              pg_dump -h postgres-service \
                -U $POSTGRES_USER \
                -d $POSTGRES_DB \
                -Fc -Z6 \
                > /backup/db-$DATE.dump
            env:
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: username
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: password
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            persistentVolumeClaim:
              claimName: postgres-backup-pvc
```

### Best Practices

1. **For Large Databases**: Use directory format with parallel jobs
   ```bash
   pg_dump -Fd -j $(nproc) -d large_db -f /backup/large_db_dir
   ```

2. **For Network Transfer**: Use custom format with compression
   ```bash
   pg_dump -Fc -Z6 -d mydb | ssh backup-server "cat > /backups/mydb.dump"
   ```

3. **For Development**: Plain SQL for easy inspection
   ```bash
   pg_dump -d mydb --no-owner --no-privileges > dev_backup.sql
   ```

## Common Kubernetes Patterns

### Init Container for Pre-Backup Tasks
```yaml
initContainers:
- name: prepare-backup
  image: busybox
  command: ['sh', '-c', 'mkdir -p /backup && chmod 777 /backup']
  volumeMounts:
  - name: backup-volume
    mountPath: /backup
```

### Using ConfigMaps for Backup Scripts
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backup-scripts
data:
  backup.sh: |
    #!/bin/bash
    set -e
    echo "Starting backup..."
    tar -czf /backup/app-$(date +%Y%m%d).tar.gz /app/data
    echo "Backup completed"
```

### Resource Limits for Backup Jobs
```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "2Gi"
    cpu: "2000m"
```
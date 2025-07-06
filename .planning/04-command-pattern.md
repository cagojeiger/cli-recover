# Command Pattern Specification

## Core Pattern
```
cli-restore [action] [target] [resource] [options]
```

## Actions

### backup
Export data from various sources
```bash
cli-restore backup pod nginx-app /data --namespace prod
cli-restore backup mongodb mongo-primary --all-databases
cli-restore backup minio bucket-name --recursive
```

### restore
Import data to various targets
```bash
cli-restore restore pod ./backup.tar nginx-app --namespace prod
cli-restore restore mongodb ./dump.gz mongo-primary
cli-restore restore minio ./bucket-backup.tar bucket-name
```

### verify
Check backup integrity
```bash
cli-restore verify backup ./backup.tar
cli-restore verify checksum ./backup.tar.sha256
```

### schedule
Set up automated backups
```bash
cli-restore schedule create "0 2 * * *" --target pod --resource nginx
cli-restore schedule list
cli-restore schedule delete schedule-id
```

### history
View operation history
```bash
cli-restore history list --days 7
cli-restore history show backup-id
cli-restore history clean --older-than 30d
```

## Targets

### Kubernetes Resources
- `pod` - Container filesystems
- `configmap` - Configuration data
- `secret` - Encrypted secrets
- `pvc` - Persistent volumes

### Databases
- `mongodb` - MongoDB databases
- `postgres` - PostgreSQL databases
- `mysql` - MySQL/MariaDB databases
- `redis` - Redis data
- `elastic` - Elasticsearch indices

### Object Storage
- `minio` - MinIO buckets
- `s3` - AWS S3 compatible

## Common Options

### Global Flags
```
--namespace, -n    Kubernetes namespace
--context         Kubernetes context
--kubeconfig      Path to kubeconfig
--timeout         Operation timeout
--verbose, -v     Verbose output
--dry-run         Preview without execution
--format          Output format (json/yaml/table)
```

### Backup Flags
```
--output, -o      Output directory
--compress        Compression type (gzip/none)
--split-size      Split archive size
--exclude         Exclude patterns
--include         Include patterns
--parallel        Parallel operations
```

### Restore Flags
```
--source          Backup source path
--target          Restore target
--overwrite       Overwrite existing
--verify          Verify after restore
```

## Examples

### Pod Backup Scenarios
```bash
# Simple backup
cli-restore backup pod nginx /data

# Multiple paths
cli-restore backup pod nginx /data,/logs,/config

# With options
cli-restore backup pod nginx /data \
  --namespace production \
  --split-size 1G \
  --compress gzip \
  --output ./backups/

# Pattern-based
cli-restore backup pod nginx / \
  --include "*.conf,*.log" \
  --exclude "*.tmp,*.cache"
```

### MongoDB Backup Scenarios
```bash
# All databases
cli-restore backup mongodb mongo-primary --all-databases

# Specific databases
cli-restore backup mongodb mongo-primary \
  --database app,sessions \
  --auth-db admin \
  --archive

# With authentication
cli-restore backup mongodb mongo-primary \
  --username backup \
  --password-file /secure/pass \
  --ssl
```

### MinIO Backup Scenarios
```bash
# Entire bucket
cli-restore backup minio my-bucket --recursive

# Specific prefix
cli-restore backup minio my-bucket/2024/ --recursive

# With versioning
cli-restore backup minio my-bucket \
  --versions \
  --endpoint https://minio.local:9000
```

### Restore Scenarios
```bash
# Pod restore
cli-restore restore pod ./backup-20240107.tar nginx-new \
  --namespace production \
  --verify

# MongoDB restore
cli-restore restore mongodb ./dump-20240107.gz mongo-primary \
  --drop-existing \
  --database myapp

# MinIO restore
cli-restore restore minio ./bucket-backup.tar restored-bucket \
  --endpoint https://minio.local:9000
```

## Port Forwarding

When needed, commands automatically handle port forwarding:
```bash
# Automatic port forward for MongoDB
cli-restore backup mongodb mongo-primary --port-forward 27017

# Manual port forward command generated
kubectl port-forward -n prod mongo-primary 27017:27017 &
```

## Command Composition in TUI

The TUI builds commands step by step:

1. **Action**: User selects `backup`
   - Command: `cli-restore backup`

2. **Target**: User selects `pod`
   - Command: `cli-restore backup pod`

3. **Resource**: User selects `nginx-app`
   - Command: `cli-restore backup pod nginx-app`

4. **Configuration**: User selects paths `/data,/logs`
   - Command: `cli-restore backup pod nginx-app /data,/logs`

5. **Options**: User configures options
   - Final: `cli-restore backup pod nginx-app /data,/logs --namespace prod --split-size 1G`

## Future Extensions

Easy to add new targets:
```bash
# Potential future targets
cli-restore backup cassandra cluster-name --keyspace myapp
cli-restore backup etcd etcd-cluster --snapshot
cli-restore backup vault vault-server --path secret/
```
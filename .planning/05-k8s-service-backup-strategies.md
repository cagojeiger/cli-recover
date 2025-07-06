# Kubernetes Service Backup Strategies

## Overview
Detailed backup strategies for services running in Kubernetes pods, with focus on Bitnami charts.

## Service-Specific Strategies

### 1. MongoDB (Bitnami)

#### Environment
- Image: `bitnami/mongodb:6.0`
- Tools included: mongodump, mongorestore ✓
- Typical data size: 10GB - 1TB+

#### Backup Methods

##### Method A: External Streaming (Recommended for > 10GB)
```bash
# Port forward
kubectl port-forward -n prod mongodb-primary 27017:27017 &

# Stream backup
mongodump --host localhost:27017 --archive --gzip > backup.gz

# OR direct streaming from pod
kubectl exec -n prod mongodb-primary -- \
  mongodump --archive --gzip > backup.gz
```

##### Method B: Pod Internal (Only for < 10GB)
```bash
# Create dump in pod
kubectl exec -n prod mongodb-primary -- \
  mongodump --out /tmp/dump

# Copy to local
kubectl cp prod/mongodb-primary:/tmp/dump ./backup

# Cleanup
kubectl exec -n prod mongodb-primary -- rm -rf /tmp/dump
```

#### Capacity Planning
```
Data Size | Method | Space Needed | Time Estimate
< 10GB    | Any    | 2x in pod    | 5-10 min
10-100GB  | Stream | Local only   | 10-60 min
> 100GB   | Stream | Local only   | 1-5 hours
```

### 2. MinIO (Bitnami)

#### Environment
- Image: `bitnami/minio:latest`
- Tools included: None (mc NOT included) ✗
- Typical data size: 100GB - 10TB+

#### Backup Methods

##### Method A: Local mc + Port Forward (Required)
```bash
# Port forward
kubectl port-forward -n prod minio-0 9000:9000 &

# Configure mc
mc alias set k8s-minio http://localhost:9000 ACCESS_KEY SECRET_KEY

# Mirror backup
mc mirror k8s-minio/bucket ./backup/

# For specific paths
mc cp --recursive k8s-minio/bucket/path/ ./backup/
```

##### Method B: Embedded mc Injection
```bash
# If mc not available locally, inject embedded binary
cli-restore inject-tool mc minio-0

# Then use injected mc
kubectl exec -n prod minio-0 -- \
  /tmp/mc mirror local/bucket /tmp/backup
```

#### Large Dataset Handling
```bash
# Parallel download for speed
mc cp --recursive --parallel=4 k8s-minio/bucket/ ./backup/

# Resume on failure
mc cp --recursive --continue k8s-minio/bucket/ ./backup/
```

### 3. PostgreSQL (Bitnami)

#### Environment
- Image: `bitnami/postgresql:15`
- Tools included: pg_dump, pg_restore ✓
- Typical data size: 1GB - 500GB

#### Backup Methods

##### Method A: Streaming Backup
```bash
# Direct streaming
kubectl exec -n prod postgres-primary -- \
  pg_dump -U postgres -d mydb --compress=9 > backup.sql.gz

# With port forward
kubectl port-forward -n prod postgres-primary 5432:5432 &
pg_dump -h localhost -U postgres mydb | gzip > backup.sql.gz
```

##### Method B: Custom Format (Parallel Restore)
```bash
# Custom format allows parallel restore
kubectl exec -n prod postgres-primary -- \
  pg_dump -U postgres -d mydb -Fc > backup.dump
```

### 4. MySQL/MariaDB (Bitnami)

#### Environment
- Image: `bitnami/mysql:8.0`
- Tools included: mysqldump ✓
- Typical data size: 1GB - 200GB

#### Backup Methods
```bash
# Streaming backup
kubectl exec -n prod mysql-primary -- \
  mysqldump -u root -p$MYSQL_ROOT_PASSWORD \
  --all-databases --single-transaction | gzip > backup.sql.gz
```

### 5. Redis (Bitnami)

#### Environment
- Image: `bitnami/redis:7.0`
- Backup method: RDB snapshot or AOF

#### Backup Methods
```bash
# Trigger RDB snapshot
kubectl exec -n prod redis-master -- redis-cli BGSAVE

# Wait for completion
kubectl exec -n prod redis-master -- redis-cli LASTSAVE

# Copy RDB file
kubectl cp prod/redis-master:/data/dump.rdb ./redis-backup.rdb
```

## Backup Decision Tree

```
1. Check data size
   ├─ < 10GB
   │  └─ Check pod free space
   │     ├─ Space > 2x data → Pod internal OK
   │     └─ Space < 2x data → Use streaming
   └─ > 10GB
      └─ Always use streaming

2. Check tool availability
   ├─ Tool in pod (mongodump, pg_dump)
   │  └─ Can use pod exec directly
   └─ Tool NOT in pod (mc for MinIO)
      ├─ Check local tool
      │  ├─ Available → Use with port-forward
      │  └─ Not available → Inject embedded or guide install
      └─ Offline mode
         └─ Use embedded binary
```

## Port Forwarding Management

### Automated Port Forward
```go
type PortForwardManager struct {
    forwards map[string]*PortForward
}

func (pfm *PortForwardManager) EnsurePortForward(
    pod, namespace string, 
    localPort, remotePort int,
) error {
    key := fmt.Sprintf("%s/%s", namespace, pod)
    
    if pf, exists := pfm.forwards[key]; exists && pf.IsAlive() {
        return nil // Already forwarding
    }
    
    // Start new port forward
    pf := &PortForward{
        Pod:       pod,
        Namespace: namespace,
        LocalPort: localPort,
        PodPort:   remotePort,
    }
    
    if err := pf.Start(); err != nil {
        return err
    }
    
    pfm.forwards[key] = pf
    return pf.WaitReady(30 * time.Second)
}
```

## Progress Monitoring

### Size Estimation
```bash
# MongoDB size
kubectl exec mongodb-primary -- mongo --eval "db.stats()"

# PostgreSQL size
kubectl exec postgres-primary -- \
  psql -U postgres -c "SELECT pg_database_size('mydb')"

# MinIO bucket size
mc du k8s-minio/bucket
```

### Progress Display
```bash
# Using pv for progress
kubectl exec pod -- mongodump --archive | \
  pv -s 50G | gzip > backup.gz

# Custom progress for mc
mc cp --json k8s-minio/bucket ./backup | \
  jq -r '.size' | progress-bar
```

## Error Recovery

### Resume Strategies
1. **Checkpointing**: Save progress periodically
2. **Incremental**: Only backup changes
3. **Retry logic**: Automatic retry on network failure
4. **Partial backup**: Complete what's possible

### Common Issues
1. **Pod storage full**: Switch to streaming
2. **Network timeout**: Implement keep-alive
3. **Auth failure**: Verify credentials
4. **Tool missing**: Use embedded binary
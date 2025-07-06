# Backup Strategy Refinement Session

## Date: 2025-07-06

## Key Insights

### 1. Bitnami Chart Analysis
- **MongoDB**: Includes mongodump/mongorestore in image ✓
- **MinIO**: Does NOT include mc client ✗
- **PostgreSQL**: Includes pg_dump/pg_restore ✓
- **MySQL**: Includes mysqldump ✓

### 2. Storage Capacity Issues

#### Problem
When backing up inside pod:
- Database: 50GB
- Temp backup file: ~40GB (compressed)
- Required free space: At least 40GB
- Risk: Pod storage exhaustion

#### Solution: Streaming Strategy
```bash
# Direct streaming - no pod storage needed
kubectl exec pod -- mongodump --archive --gzip | pv > backup.gz
```

### 3. Backup Strategy by Size

#### Small (< 10GB)
- Pod internal backup OK if space > 2x data size
- Fast and simple

#### Medium (10-100GB)  
- Streaming required
- Progress monitoring essential
- Network bandwidth consideration

#### Large (> 100GB)
- Parallel processing
- Collection/table level splits
- Incremental options

### 4. Tool Availability Matrix

| Service | Local Tool | Pod Tool (Bitnami) | Strategy |
|---------|------------|-------------------|----------|
| MongoDB | mongodump | ✓ Included | Both work |
| MinIO | mc | ✗ Not included | Local + port-forward |
| PostgreSQL | pg_dump | ✓ Included | Both work |
| Files | tar | ✓ Always available | Pod exec |

### 5. Offline Environment Support

#### Binary Embedding Strategy
- Use Go 1.16+ embed feature
- Include essential binaries (mc, mongodump)
- Compress with gzip -9
- Platform-specific builds

#### Size Estimates
- Base cli-restore: ~20MB
- +mc (4 platforms): +60MB
- +mongodb-tools: +180MB  
- Total offline version: ~300-400MB

#### Security Considerations
- Temporary injection to /tmp
- Cleanup guaranteed
- SHA256 verification
- No persistent pod changes
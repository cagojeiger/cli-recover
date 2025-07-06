# Offline Environment Support Plan

## Overview
Strategy for supporting environments without internet access by embedding essential binaries.

## Binary Embedding Architecture

### 1. Go Embed Implementation

```go
//go:embed binaries/*
var embeddedFS embed.FS

type EmbeddedBinary struct {
    Name     string
    Version  string
    Platform string // linux-amd64, darwin-arm64, etc.
    Size     int64
    SHA256   string
}
```

### 2. Binary Selection

#### Essential Binaries (Always Include)
- **mc (MinIO Client)**
  - Why: Not included in Bitnami MinIO
  - Size: ~15MB compressed per platform
  - Platforms: linux-amd64, linux-arm64

#### Optional Binaries (Separate Builds)
- **MongoDB Database Tools**
  - mongodump, mongorestore
  - Size: ~45MB compressed per platform
  - Note: Usually included in Bitnami MongoDB

- **PostgreSQL Client**
  - pg_dump, pg_restore
  - Size: ~5MB compressed per platform
  - Note: Usually included in Bitnami PostgreSQL

### 3. Build Variants

```makefile
# Minimal build (mc only for common platforms)
build-minimal:
	go build -tags minimal -o cli-restore

# Standard build (mc for all platforms)  
build-standard:
	go build -tags standard -o cli-restore

# Full offline build (all tools, all platforms)
build-offline-full:
	go build -tags offline_full -o cli-restore

# Custom platform build
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -tags "minimal linux amd64" -o cli-restore
```

### 4. Size Analysis

```
Build Variant         | Binary Size | Included Tools
---------------------|-------------|----------------
Minimal              | ~35MB       | cli-restore + mc (linux-amd64 only)
Standard             | ~80MB       | cli-restore + mc (4 platforms)
Full Offline         | ~300MB      | cli-restore + all tools (4 platforms)
Platform Specific    | ~40MB       | cli-restore + tools (1 platform)
```

### 5. Binary Management

#### Extraction Flow
```go
func (bm *BinaryManager) GetBinary(tool string) (string, error) {
    // 1. Check cache
    cachedPath := bm.getCachedPath(tool)
    if exists(cachedPath) {
        return cachedPath, nil
    }
    
    // 2. Extract from embedded
    if bm.hasEmbedded(tool) {
        return bm.extractEmbedded(tool)
    }
    
    // 3. Try download (if online)
    if bm.isOnline() {
        return bm.downloadBinary(tool)
    }
    
    return "", ErrBinaryNotAvailable
}
```

#### Pod Injection
```go
func (bm *BinaryManager) InjectToPod(
    tool string, 
    pod, namespace string,
) (string, error) {
    // 1. Extract to temp
    localPath, err := bm.GetBinary(tool)
    if err != nil {
        return "", err
    }
    
    // 2. Generate unique path in pod
    podPath := fmt.Sprintf("/tmp/%s-%s", tool, uuid.New())
    
    // 3. Copy to pod
    if err := kubectl.Copy(localPath, namespace, pod, podPath); err != nil {
        return "", err
    }
    
    // 4. Make executable
    if err := kubectl.Exec(namespace, pod, "chmod", "+x", podPath); err != nil {
        return "", err
    }
    
    // 5. Register for cleanup
    bm.registerCleanup(namespace, pod, podPath)
    
    return podPath, nil
}
```

### 6. Security Measures

#### Binary Verification
```go
var trustedBinaries = map[string]string{
    "mc-linux-amd64-RELEASE.2024-01-05": "sha256:abc123...",
    "mongodump-linux-amd64-100.9.0":     "sha256:def456...",
}

func verifyBinary(data []byte, tool, version string) error {
    key := fmt.Sprintf("%s-%s", tool, version)
    expectedHash := trustedBinaries[key]
    
    actualHash := sha256.Sum256(data)
    if actualHash != expectedHash {
        return ErrBinaryCompromised
    }
    
    return nil
}
```

#### Cleanup Guarantee
```go
type Cleanup struct {
    Namespace string
    Pod       string
    Path      string
}

func (bm *BinaryManager) ensureCleanup(cleanup Cleanup) {
    // Cleanup in defer
    defer func() {
        kubectl.Exec(cleanup.Namespace, cleanup.Pod, 
            "rm", "-f", cleanup.Path)
    }()
    
    // Also register signal handler
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        bm.cleanupAll()
    }()
}
```

### 7. User Experience

#### Offline Mode Detection
```
┌─ Offline Mode Detected ─────────────────────────────────────────────────┐
│                                                                         │
│  🔌 No internet connection detected                                    │
│                                                                         │
│  Available offline tools:                                               │
│  ✅ mc (MinIO Client) - v2024.01.05                                   │
│  ✅ kubectl - system installed                                        │
│  ❌ mongodump - not included in this build                            │
│                                                                         │
│  Options:                                                               │
│  1. Continue with available tools                                      │
│  2. Use tools from pod (if available)                                 │
│  3. Exit and use offline-full build                                   │
│                                                                         │
│  Note: Download cli-restore-offline-full for complete offline support │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Binary Injection Consent
```
┌─ Binary Injection Required ─────────────────────────────────────────────┐
│                                                                         │
│  MinIO backup requires 'mc' client                                     │
│                                                                         │
│  Proposed action:                                                       │
│  1. Extract embedded mc binary (15MB)                                 │
│  2. Copy to minio-0:/tmp/mc-temp-xxx                                  │
│  3. Use for backup operation                                          │
│  4. Remove after completion                                           │
│                                                                         │
│  Security:                                                              │
│  • Binary SHA256: abc123...def456                                     │
│  • Official MinIO release 2024.01.05                                  │
│  • Temporary only, no persistent changes                              │
│                                                                         │
│  [y] Yes, proceed  [n] No, cancel  [v] Verify binary                 │
└─────────────────────────────────────────────────────────────────────────┘
```

### 8. Distribution Strategy

```yaml
# GitHub Release Assets
cli-restore-darwin-amd64          # Minimal, no embedded
cli-restore-darwin-arm64          # Minimal, no embedded  
cli-restore-linux-amd64           # Minimal, no embedded
cli-restore-linux-arm64           # Minimal, no embedded

cli-restore-offline-darwin-amd64  # With mc embedded
cli-restore-offline-darwin-arm64  # With mc embedded
cli-restore-offline-linux-amd64   # With mc embedded
cli-restore-offline-linux-arm64   # With mc embedded

cli-restore-offline-full-linux-amd64  # All tools embedded
```

### 9. Update Mechanism

```go
// Check for binary updates when online
func (bm *BinaryManager) CheckUpdates() {
    if !bm.isOnline() {
        return
    }
    
    updates := bm.registry.GetAvailableUpdates()
    if len(updates) > 0 {
        fmt.Printf("Binary updates available:\n")
        for _, update := range updates {
            fmt.Printf("- %s: %s → %s\n", 
                update.Tool, update.Current, update.Latest)
        }
        fmt.Printf("\nRun 'cli-restore update-tools' to update\n")
    }
}
```

### 10. Testing Strategy

```bash
# Test offline functionality
docker run --rm -it --network none \
  -v $PWD:/app \
  cli-restore-test \
  /app/cli-restore-offline backup minio test-bucket

# Verify no network calls
strace -e network cli-restore-offline backup ...
```

## Conclusion

This offline support plan ensures:
1. **Flexibility**: Multiple build variants for different needs
2. **Security**: Binary verification and cleanup
3. **Usability**: Clear user communication
4. **Efficiency**: Minimal size increase for common cases
5. **Reliability**: Works completely offline when needed
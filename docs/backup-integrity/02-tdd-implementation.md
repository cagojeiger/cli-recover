# TDD로 백업 무결성 구현하기

## TDD (Test-Driven Development) 개요

### Red-Green-Refactor 사이클
```
┌─────────┐     ┌─────────┐     ┌──────────┐
│   RED   │ --> │  GREEN  │ --> │ REFACTOR │
│         │     │         │     │          │
│ 실패하는│     │ 테스트  │     │  코드    │
│ 테스트  │     │  통과   │     │  개선    │
└─────────┘     └─────────┘     └──────────┘
     ↑                                 │
     └─────────────────────────────────┘
```

### TDD의 3가지 규칙 (Uncle Bob)
1. 실패하는 단위 테스트를 작성하기 전에는 프로덕션 코드를 작성하지 않는다
2. 컴파일은 실패하지 않으면서 실행이 실패하는 정도로만 단위 테스트를 작성한다
3. 현재 실패하는 테스트를 통과할 정도로만 실제 코드를 작성한다

## 파일시스템 테스트 전략

### 1. 인터페이스 기반 추상화

```go
// filesystem_interface.go
type FileSystem interface {
    Create(name string) (File, error)
    Open(name string) (File, error)
    Remove(name string) error
    Rename(oldpath, newpath string) error
    Stat(name string) (os.FileInfo, error)
}

type File interface {
    io.WriteCloser
    Sync() error
}

// 프로덕션 구현
type OSFileSystem struct{}

func (fs *OSFileSystem) Create(name string) (File, error) {
    return os.Create(name)
}

func (fs *OSFileSystem) Rename(oldpath, newpath string) error {
    return os.Rename(oldpath, newpath)
}

// 테스트용 Mock
type MockFileSystem struct {
    files      map[string]*mockFile
    shouldFail map[string]bool
}
```

### 2. 메모리 기반 Mock 구현

```go
type mockFile struct {
    name   string
    buffer *bytes.Buffer
    closed bool
    synced bool
}

func (f *mockFile) Write(p []byte) (n int, err error) {
    if f.closed {
        return 0, errors.New("file closed")
    }
    return f.buffer.Write(p)
}

func (f *mockFile) Close() error {
    f.closed = true
    return nil
}

func (f *mockFile) Sync() error {
    f.synced = true
    return nil
}
```

## Step 1: RED - 실패하는 테스트 작성

### Test 1: 백업 중단 시 임시 파일만 남아야 함

```go
func TestBackupInterruption_LeavesOnlyTempFile(t *testing.T) {
    // Arrange
    fs := NewMockFileSystem()
    provider := NewProvider(fs)
    opts := BackupOptions{
        OutputFile: "backup.tar",
    }
    
    // 쓰기 중 실패 시뮬레이션
    fs.SetWriteFailure("backup.tar.tmp", afterBytes(1024))
    
    // Act
    err := provider.Execute(context.Background(), opts)
    
    // Assert
    assert.Error(t, err)
    assert.False(t, fs.Exists("backup.tar"), "최종 파일이 존재하면 안됨")
    assert.True(t, fs.Exists("backup.tar.tmp"), "임시 파일은 존재해야 함")
}
```

### Test 2: 성공 시 최종 파일만 존재해야 함

```go
func TestBackupSuccess_OnlyFinalFileExists(t *testing.T) {
    // Arrange
    fs := NewMockFileSystem()
    provider := NewProvider(fs)
    opts := BackupOptions{
        OutputFile: "backup.tar",
    }
    
    // Act
    err := provider.Execute(context.Background(), opts)
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, fs.Exists("backup.tar"), "최종 파일이 존재해야 함")
    assert.False(t, fs.Exists("backup.tar.tmp"), "임시 파일은 삭제되어야 함")
}
```

### Test 3: 체크섬이 정확히 계산되어야 함

```go
func TestChecksum_CalculatedDuringStreaming(t *testing.T) {
    // Arrange
    fs := NewMockFileSystem()
    provider := NewProvider(fs)
    testData := []byte("test backup data")
    expectedChecksum := sha256.Sum256(testData)
    
    opts := BackupOptions{
        OutputFile: "backup.tar",
        SourceData: bytes.NewReader(testData),
    }
    
    // Act
    result, err := provider.ExecuteWithResult(context.Background(), opts)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, hex.EncodeToString(expectedChecksum[:]), 
                   result.Checksum, "체크섬이 일치해야 함")
}
```

## Step 2: GREEN - 최소한의 구현

### 초기 구현 (하드코딩 허용)

```go
func (p *Provider) Execute(ctx context.Context, opts BackupOptions) error {
    tempFile := opts.OutputFile + ".tmp"
    
    // 임시 파일 생성
    f, err := p.fs.Create(tempFile)
    if err != nil {
        return err
    }
    defer f.Close()
    
    // 데이터 쓰기 (아직 체크섬 없음)
    _, err = io.Copy(f, opts.SourceData)
    if err != nil {
        p.fs.Remove(tempFile)  // 실패 시 정리
        return err
    }
    
    // 성공 시 이동
    return p.fs.Rename(tempFile, opts.OutputFile)
}
```

## Step 3: REFACTOR - 코드 개선

### ChecksumWriter 추출

```go
// RED: 체크섬 기능을 위한 테스트 추가
func TestChecksumWriter_CalculatesCorrectly(t *testing.T) {
    // Arrange
    buffer := &bytes.Buffer{}
    writer := NewChecksumWriter(buffer)
    testData := []byte("hello world")
    
    // Act
    n, err := writer.Write(testData)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, len(testData), n)
    assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfa"+
                   "c484efe37a5380ee9088f7ace2efcde9",
                   writer.Sum())
}

// GREEN: ChecksumWriter 구현
type ChecksumWriter struct {
    writer io.Writer
    hash   hash.Hash
}

func NewChecksumWriter(w io.Writer) *ChecksumWriter {
    return &ChecksumWriter{
        writer: w,
        hash:   sha256.New(),
    }
}

func (cw *ChecksumWriter) Write(p []byte) (n int, err error) {
    n, err = cw.writer.Write(p)
    if err != nil {
        return n, err
    }
    cw.hash.Write(p[:n])
    return n, nil
}

// REFACTOR: Provider에 통합
func (p *Provider) Execute(ctx context.Context, opts BackupOptions) error {
    tempFile := opts.OutputFile + ".tmp"
    
    var success bool
    defer func() {
        if !success {
            p.fs.Remove(tempFile)
        }
    }()
    
    f, err := p.fs.Create(tempFile)
    if err != nil {
        return err
    }
    defer f.Close()
    
    // ChecksumWriter 사용
    cw := NewChecksumWriter(f)
    _, err = io.Copy(cw, opts.SourceData)
    if err != nil {
        return err
    }
    
    if err := f.Sync(); err != nil {
        return err
    }
    
    if err := p.fs.Rename(tempFile, opts.OutputFile); err != nil {
        return err
    }
    
    success = true
    return nil
}
```

## 통합 테스트

### 실제 파일시스템 테스트

```go
func TestIntegration_RealFileSystem(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Arrange
    tempDir := t.TempDir()
    outputFile := filepath.Join(tempDir, "backup.tar")
    
    provider := NewProvider(&OSFileSystem{})
    opts := BackupOptions{
        OutputFile: outputFile,
        SourceData: generateTestData(1024 * 1024), // 1MB
    }
    
    // Act
    err := provider.Execute(context.Background(), opts)
    
    // Assert
    assert.NoError(t, err)
    assert.FileExists(t, outputFile)
    
    // 체크섬 검증
    info, _ := os.Stat(outputFile)
    assert.Equal(t, int64(1024*1024), info.Size())
}
```

### 동시성 테스트

```go
func TestConcurrentBackups_DifferentFiles(t *testing.T) {
    // Arrange
    fs := NewMockFileSystem()
    provider := NewProvider(fs)
    
    var wg sync.WaitGroup
    errors := make([]error, 10)
    
    // Act - 10개 동시 백업
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            opts := BackupOptions{
                OutputFile: fmt.Sprintf("backup%d.tar", idx),
                SourceData: bytes.NewReader([]byte(fmt.Sprintf("data%d", idx))),
            }
            errors[idx] = provider.Execute(context.Background(), opts)
        }(i)
    }
    
    wg.Wait()
    
    // Assert - 모두 성공
    for i, err := range errors {
        assert.NoError(t, err, "백업 %d 실패", i)
        assert.True(t, fs.Exists(fmt.Sprintf("backup%d.tar", i)))
    }
}
```

## 엣지 케이스 테스트

### Test: 디스크 공간 부족

```go
func TestDiskSpaceFull_CleansUpTempFile(t *testing.T) {
    // Arrange
    fs := NewMockFileSystem()
    fs.SetDiskQuota(1024) // 1KB 제한
    
    provider := NewProvider(fs)
    opts := BackupOptions{
        OutputFile: "backup.tar",
        SourceData: bytes.NewReader(make([]byte, 2048)), // 2KB 데이터
    }
    
    // Act
    err := provider.Execute(context.Background(), opts)
    
    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "no space left")
    assert.False(t, fs.Exists("backup.tar"))
    assert.False(t, fs.Exists("backup.tar.tmp")) // 정리됨
}
```

### Test: Rename 실패 (크로스 파일시스템)

```go
func TestCrossFilesystemRename_FallbackToCopy(t *testing.T) {
    // Arrange
    fs := NewMockFileSystem()
    fs.SetCrossFilesystem("/tmp/backup.tar.tmp", "/mnt/backup.tar")
    
    provider := NewProvider(fs)
    opts := BackupOptions{
        OutputFile: "/mnt/backup.tar",
        SourceData: bytes.NewReader([]byte("test data")),
    }
    
    // Act
    err := provider.Execute(context.Background(), opts)
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, fs.Exists("/mnt/backup.tar"))
    assert.Equal(t, "test data", fs.ReadFile("/mnt/backup.tar"))
}
```

## 성능 테스트

### 벤치마크 테스트

```go
func BenchmarkBackupWithChecksum(b *testing.B) {
    fs := NewMockFileSystem()
    provider := NewProvider(fs)
    
    data := make([]byte, 10*1024*1024) // 10MB
    rand.Read(data)
    
    b.ResetTimer()
    b.SetBytes(int64(len(data)))
    
    for i := 0; i < b.N; i++ {
        opts := BackupOptions{
            OutputFile: fmt.Sprintf("backup%d.tar", i),
            SourceData: bytes.NewReader(data),
        }
        provider.Execute(context.Background(), opts)
    }
}
```

### 메모리 프로파일링

```go
func TestMemoryUsage_ConstantRegardlessOfFileSize(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping memory test")
    }
    
    // 메모리 사용량 측정
    var m1, m2 runtime.MemStats
    
    // 1MB 파일 백업
    runtime.GC()
    runtime.ReadMemStats(&m1)
    backupFile(1 * 1024 * 1024)
    runtime.ReadMemStats(&m2)
    mem1MB := m2.Alloc - m1.Alloc
    
    // 100MB 파일 백업
    runtime.GC()
    runtime.ReadMemStats(&m1)
    backupFile(100 * 1024 * 1024)
    runtime.ReadMemStats(&m2)
    mem100MB := m2.Alloc - m1.Alloc
    
    // 메모리 사용량이 파일 크기에 비례하지 않아야 함
    ratio := float64(mem100MB) / float64(mem1MB)
    assert.Less(t, ratio, 2.0, "메모리 사용량이 일정해야 함")
}
```

## TDD 모범 사례

### 1. 테스트 명명 규칙

```go
// Given_When_Then 패턴
func TestBackup_WhenInterrupted_LeavesOnlyTempFile(t *testing.T)

// Should 패턴
func TestBackup_ShouldCalculateChecksumDuringStreaming(t *testing.T)

// 시나리오 기반
func TestBackup_SuccessScenario_CreatesValidArchive(t *testing.T)
```

### 2. 테스트 구조화 (AAA 패턴)

```go
func TestExample(t *testing.T) {
    // Arrange (준비)
    fs := NewMockFileSystem()
    provider := NewProvider(fs)
    
    // Act (실행)
    err := provider.Execute(ctx, opts)
    
    // Assert (검증)
    assert.NoError(t, err)
}
```

### 3. 테스트 격리

```go
func TestWithIsolation(t *testing.T) {
    // 각 테스트는 독립적이어야 함
    t.Run("subtest1", func(t *testing.T) {
        t.Parallel() // 병렬 실행 가능
        // 독립된 테스트
    })
    
    t.Run("subtest2", func(t *testing.T) {
        t.Parallel()
        // 독립된 테스트
    })
}
```

## TDD 안티패턴 피하기

### 1. 과도한 Mock 피하기

```go
// ❌ 나쁜 예 - 모든 것을 Mock
func TestBadExample(t *testing.T) {
    mockFS := NewMockFileSystem()
    mockHash := NewMockHash()
    mockWriter := NewMockWriter()
    // 테스트가 구현을 그대로 반복함
}

// ✅ 좋은 예 - 필요한 부분만 Mock
func TestGoodExample(t *testing.T) {
    fs := NewMockFileSystem() // 파일시스템만 Mock
    provider := NewProvider(fs)
    // 실제 로직 테스트
}
```

### 2. 테스트 가능한 설계

```go
// ❌ 테스트 어려운 설계
func BackupFiles(source, dest string) error {
    // os 패키지 직접 사용
    file, err := os.Create(dest)
    // ...
}

// ✅ 테스트 가능한 설계
func BackupFiles(fs FileSystem, source, dest string) error {
    // 인터페이스 사용
    file, err := fs.Create(dest)
    // ...
}
```

## 지속적 개선

### 커버리지 목표

```bash
# 테스트 커버리지 확인
go test -cover ./...

# 상세 커버리지 리포트
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 목표: 핵심 로직 90% 이상
```

### 리팩토링 체크리스트

1. ✅ 모든 테스트가 통과하는가?
2. ✅ 중복 코드가 제거되었는가?
3. ✅ 명확한 이름을 사용하는가?
4. ✅ 단일 책임 원칙을 따르는가?
5. ✅ 의존성이 올바른 방향인가?

TDD는 단순히 테스트를 먼저 작성하는 것이 아니라, 테스트를 통해 설계를 개선하고 코드 품질을 높이는 개발 방법론입니다. Red-Green-Refactor 사이클을 통해 점진적으로 기능을 구현하면서 동시에 안전망을 구축할 수 있습니다.
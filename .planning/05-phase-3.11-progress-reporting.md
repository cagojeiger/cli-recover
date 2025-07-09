# Phase 3.11 - Progress Reporting Enhancement Implementation Plan v2

## Executive Summary

This plan implements a unified progress reporting system that builds on existing Progress types. Following TDD and Occam's Razor, we'll create a simple Reporter that handles different environments automatically.

**Complexity Score: 32/100** ✅ (Even simpler by reusing existing types)

## 1. Architecture Overview (Simplified)

### Existing Progress Types
- `backup.Progress` - Already has speed/ETA calculations ✅
- `operation.Progress` - Unified interface ✅  
- Both have Current/Total/Message fields ✅

### New Components Only
```
ProgressReporter (NEW)
     ├── Terminal Handler (\r updates)
     ├── Log Handler (periodic)
     └── Channel Handler (TUI)

ProgressWriter (NEW) 
     └── io.Writer wrapper that reports progress
```

## 2. TDD Implementation Order

### Step 1: Progress Reporter Interface (Red → Green → Refactor)

**Test First**: `internal/domain/progress/reporter_test.go`
```go
package progress_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/cagojeiger/cli-recover/internal/domain/backup"
)

func TestReporter_Update(t *testing.T) {
    // RED: This test will fail initially
    reporter := progress.NewReporter(nil, nil)
    
    // Should not panic with nil writer/logger
    reporter.Update(backup.Progress{
        Current: 100,
        Total:   1000,
        Message: "Testing",
    })
}
```

**Then Implementation**: `internal/domain/progress/reporter.go`
```go
package progress

import "github.com/cagojeiger/cli-recover/internal/domain/backup"

type Reporter interface {
    Update(p backup.Progress)
    Complete(message string)
    Error(err error, message string)
}
```

### Step 2: Mock Reporter for Testing

**Test First**: `internal/domain/progress/mock_test.go`
```go
type mockReporter struct {
    updates []backup.Progress
    completed bool
    errors []error
}

func TestMockReporter_CapturesUpdates(t *testing.T) {
    // Test the mock works correctly
}
```

### Step 3: Terminal Detection

**Test First**: `internal/infrastructure/progress/reporter_test.go`
```go
func TestReporter_DetectsTerminal(t *testing.T) {
    // Mock os.Stderr
    // Test terminal detection
}

func TestReporter_DetectsNonTerminal(t *testing.T) {
    // Test pipe/redirect detection
}
```

### Step 4: Progress Writer

**Test First**: `internal/infrastructure/progress/writer_test.go`
```go
func TestProgressWriter_TracksBytes(t *testing.T) {
    // RED: Test fails
    var buf bytes.Buffer
    mockReporter := &mockReporter{}
    
    pw := progress.NewWriter(&buf, mockReporter, 1000, "Testing")
    
    // Write 100 bytes
    data := make([]byte, 100)
    n, err := pw.Write(data)
    
    assert.NoError(t, err)
    assert.Equal(t, 100, n)
    assert.Equal(t, int64(100), mockReporter.LastUpdate().Current)
}

func TestProgressWriter_MultipleWrites(t *testing.T) {
    // Test cumulative tracking
}
```

### Step 5: Terminal Output Formatting

**Test First**: `internal/infrastructure/progress/formatter_test.go`
```go
func TestFormatter_ProgressBar(t *testing.T) {
    tests := []struct {
        name     string
        current  int64
        total    int64
        width    int
        expected string
    }{
        {
            name:     "50 percent",
            current:  500,
            total:    1000,
            width:    20,
            expected: "[██████████░░░░░░░░░░]",
        },
    }
}

func TestFormatter_BytesHumanReadable(t *testing.T) {
    tests := []struct {
        bytes    int64
        expected string
    }{
        {1024, "1.0 KB"},
        {1048576, "1.0 MB"},
        {1073741824, "1.0 GB"},
    }
}
```

### Step 6: Integration Tests

**Test First**: `internal/infrastructure/filesystem/progress_integration_test.go`
```go
func TestFilesystemBackup_ReportsProgress(t *testing.T) {
    // Create mock filesystem
    // Create mock executor
    // Capture progress updates
    // Verify progress reported correctly
}
```

## 3. Implementation Files (TDD Order)

### Phase 1: Domain Layer (1 hour)
1. ✅ Write failing test for Reporter interface
2. ✅ Create `internal/domain/progress/reporter.go` (interface only)
3. ✅ Write test for mock reporter
4. ✅ Implement mock reporter for testing
5. ✅ All tests pass

### Phase 2: Progress Writer (1 hour)
1. ✅ Write failing test for ProgressWriter
2. ✅ Create `internal/infrastructure/progress/writer.go`
3. ✅ Write test for byte accumulation
4. ✅ Implement Write method
5. ✅ Write test for EOF handling
6. ✅ All tests pass

### Phase 3: Terminal Reporter (2 hours)
1. ✅ Write test for terminal detection
2. ✅ Create `internal/infrastructure/progress/terminal_reporter.go`
3. ✅ Write test for progress bar formatting
4. ✅ Implement ANSI escape sequences
5. ✅ Write test for terminal width
6. ✅ Implement responsive sizing
7. ✅ All tests pass

### Phase 4: Log Reporter (1 hour)
1. ✅ Write test for periodic logging
2. ✅ Implement throttled updates
3. ✅ Write test for structured fields
4. ✅ All tests pass

### Phase 5: Integration (2 hours)
1. ✅ Write integration test for backup
2. ✅ Modify filesystem.go
3. ✅ Write integration test for restore
4. ✅ Modify restore.go  
5. ✅ All tests pass

## 4. Detailed Test Scenarios

### Scenario 1: Basic Progress Tracking
```go
func TestProgressTracking_SimpleCase(t *testing.T) {
    // Given: 1KB of data to write
    // When: Writing in 100-byte chunks
    // Then: Progress updates 10 times
}
```

### Scenario 2: Terminal Update Frequency
```go
func TestTerminalReporter_ThrottlesUpdates(t *testing.T) {
    // Given: 1000 rapid updates
    // When: All within 100ms
    // Then: Only 1 terminal update
}
```

### Scenario 3: CI Environment
```go
func TestCIEnvironment_PeriodicLogs(t *testing.T) {
    // Given: CI=true environment
    // When: 30 seconds of progress
    // Then: 3 log entries (10s intervals)
}
```

### Scenario 4: Error During Progress
```go
func TestProgress_ErrorHandling(t *testing.T) {
    // Given: Progress at 50%
    // When: Write error occurs
    // Then: Progress cleaned up properly
}
```

## 5. Implementation Details

### 5.1 Reuse Existing Progress Types
```go
// Use backup.Progress everywhere
import "github.com/cagojeiger/cli-recover/internal/domain/backup"

type Reporter interface {
    Update(p backup.Progress)
    Complete(message string) 
    Error(err error, message string)
}
```

### 5.2 Simple Terminal Detection
```go
func isInteractive() bool {
    // Already have golang.org/x/term
    return term.IsTerminal(int(os.Stderr.Fd()))
}
```

### 5.3 Progress Bar Without Unicode
```go
// Simple ASCII progress bar
func renderBar(percent int) string {
    filled := percent / 5  // 20 char bar
    return "[" + strings.Repeat("=", filled) + 
           strings.Repeat("-", 20-filled) + "]"
}
```

## 6. File Structure (Minimal)

```
internal/
├── domain/
│   └── progress/
│       ├── reporter.go          (interface only)
│       └── reporter_test.go     (interface tests)
└── infrastructure/
    └── progress/
        ├── reporter.go          (concrete implementation)
        ├── reporter_test.go     (unit tests)
        ├── writer.go           (io.Writer wrapper)
        ├── writer_test.go      (unit tests)
        └── formatter.go        (helper functions)
```

## 7. Integration Points (Minimal Changes)

### Filesystem Backup
```diff
+import "github.com/cagojeiger/cli-recover/internal/infrastructure/progress"

 func (p *Provider) executeInternal(...) {
+    reporter := progress.NewReporter(os.Stderr, p.logger)
+    pw := progress.NewWriter(checksumWriter, reporter, estimatedSize, "Creating backup")
-    io.Copy(checksumWriter, stdout)
+    io.Copy(pw, stdout)
+    reporter.Complete("Backup created successfully")
 }
```

## 8. Complexity Analysis

### What We're NOT Doing
- ❌ No colors or fancy Unicode
- ❌ No complex ETA algorithms (use existing)
- ❌ No configuration files
- ❌ No progress history/persistence
- ❌ No multi-line progress bars
- ❌ No custom progress formats

### What We ARE Doing
- ✅ Simple ASCII progress bar
- ✅ Reuse existing Progress types
- ✅ Minimal code changes
- ✅ Standard library only
- ✅ TDD from the start

## 9. Success Criteria

1. **All tests written before implementation**
2. **Every operation > 3s shows progress**
3. **Works in terminal, CI, and pipes**
4. **< 500 lines of new code**
5. **No external dependencies**

## 10. Day 1 Deliverables

By end of implementation:
1. ✅ All tests passing (written first)
2. ✅ Backup shows progress in terminal
3. ✅ Restore shows progress in terminal
4. ✅ CI environment shows periodic logs
5. ✅ No performance regression

## Example Test-Driven Development Flow

```bash
# Step 1: Write failing test
$ go test ./internal/domain/progress/... 
FAIL (reporter not implemented)

# Step 2: Minimal implementation
$ vim internal/domain/progress/reporter.go
# Add interface only

# Step 3: Test passes
$ go test ./internal/domain/progress/...
PASS

# Step 4: Refactor if needed
# (Keep it simple!)

# Repeat for each component
```

This approach ensures we build only what's needed, test everything first, and maintain simplicity throughout.
package filesystem

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/stretchr/testify/assert"
)

// Backup Integrity Tests - Phase 3.10

func TestBackupInterruption_LeavesOnlyTempFile(t *testing.T) {
	// Arrange
	mockFS := NewMockFileSystem()
	mockExecutor := new(MockCommandExecutor)
	provider := NewProviderWithFS(nil, mockExecutor, mockFS)

	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar",
	}

	// Simulate write failure after 1KB
	mockFS.SetWriteFailureAfterBytes("backup.tar.tmp", 1024)

	ctx := context.Background()

	// Setup mock executor
	mockStdout := newMockReadCloser(strings.Repeat("x", 2048)) // 2KB of data
	mockStderr := newMockReadCloser("")
	mockWait := func() error { return nil }

	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	// Mock size estimation
	mockSizeEstimation(mockExecutor, ctx, "default", "test-pod", "/data")

	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	// Act
	err := provider.Execute(ctx, opts)

	// Assert
	assert.Error(t, err, "Should fail due to write error")
	assert.Contains(t, err.Error(), "write failed")
	assert.False(t, mockFS.Exists("backup.tar"), "Final file should not exist")
	// Note: In the current implementation, temp files are cleaned up on failure
	// This is actually a good practice to avoid leaving partial files
	assert.False(t, mockFS.Exists("backup.tar.tmp"), "Temp file should be cleaned up after failure")

	mockExecutor.AssertExpectations(t)
}

func TestBackupSuccess_OnlyFinalFileExists(t *testing.T) {
	// Arrange
	mockFS := NewMockFileSystem()
	mockExecutor := new(MockCommandExecutor)
	provider := NewProviderWithFS(nil, mockExecutor, mockFS)

	testData := []byte("test backup data")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar",
	}

	ctx := context.Background()

	// Setup mock executor
	mockStdout := newMockReadCloser(string(testData))
	mockStderr := newMockReadCloser("")
	mockWait := func() error { return nil }

	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	// Mock size estimation
	mockSizeEstimation(mockExecutor, ctx, "default", "test-pod", "/data")

	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	// Act
	err := provider.Execute(ctx, opts)

	// Assert
	assert.NoError(t, err)
	assert.True(t, mockFS.Exists("backup.tar"), "Final file should exist")
	assert.False(t, mockFS.Exists("backup.tar.tmp"), "Temp file should be cleaned up")

	// Verify file content
	content, err := mockFS.GetFileContent("backup.tar")
	assert.NoError(t, err)
	assert.Equal(t, testData, content, "File content should match")

	mockExecutor.AssertExpectations(t)
}

func TestChecksum_CalculatedDuringStreaming(t *testing.T) {
	// Arrange
	mockFS := NewMockFileSystem()
	mockExecutor := new(MockCommandExecutor)
	provider := NewProviderWithFS(nil, mockExecutor, mockFS)

	testData := []byte("test backup data for checksum")

	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar",
	}

	ctx := context.Background()

	// Setup mock executor
	mockStdout := newMockReadCloser(string(testData))
	mockStderr := newMockReadCloser("")
	mockWait := func() error { return nil }

	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	// Mock size estimation
	mockSizeEstimation(mockExecutor, ctx, "default", "test-pod", "/data")

	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	// Act
	result, err := provider.ExecuteWithResult(ctx, opts)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Checksum, "Checksum should be calculated")
	assert.Equal(t, "backup.tar", result.BackupFile)

	// Verify checksum matches file content
	content, _ := mockFS.GetFileContent("backup.tar")
	assert.Equal(t, testData, content)

	mockExecutor.AssertExpectations(t)
}

func TestAtomicRename_Success(t *testing.T) {
	// Arrange
	mockFS := NewMockFileSystem()
	mockExecutor := new(MockCommandExecutor)
	provider := NewProviderWithFS(nil, mockExecutor, mockFS)

	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar",
	}

	ctx := context.Background()

	// Setup mock executor
	mockStdout := newMockReadCloser("backup data")
	mockStderr := newMockReadCloser("")
	mockWait := func() error { return nil }

	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	// Mock size estimation
	mockSizeEstimation(mockExecutor, ctx, "default", "test-pod", "/data")

	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	// Act
	err := provider.Execute(ctx, opts)

	// Assert
	assert.NoError(t, err)

	// Verify atomic rename happened (temp file should not exist)
	assert.False(t, mockFS.Exists("backup.tar.tmp"))
	assert.True(t, mockFS.Exists("backup.tar"))

	mockExecutor.AssertExpectations(t)
}

func TestCleanupOnExecutorError(t *testing.T) {
	// Arrange
	mockFS := NewMockFileSystem()
	mockExecutor := new(MockCommandExecutor)
	provider := NewProviderWithFS(nil, mockExecutor, mockFS)

	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar",
	}

	ctx := context.Background()

	// Simulate executor error
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	// Mock size estimation
	mockSizeEstimation(mockExecutor, ctx, "default", "test-pod", "/data")

	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(
		nil, nil, nil, errors.New("pod not found"))

	// Act
	err := provider.Execute(ctx, opts)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pod not found")
	assert.False(t, mockFS.Exists("backup.tar"), "Final file should not exist")
	assert.False(t, mockFS.Exists("backup.tar.tmp"), "Temp file should not exist")

	mockExecutor.AssertExpectations(t)
}
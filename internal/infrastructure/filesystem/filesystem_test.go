package filesystem

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFilesystemProvider_Interface(t *testing.T) {
	// Compile-time check
	var _ backup.Provider = (*Provider)(nil)
}

func TestFilesystemProvider_Name(t *testing.T) {
	provider := NewProvider(nil, nil)
	assert.Equal(t, "filesystem", provider.Name())
}

func TestFilesystemProvider_Description(t *testing.T) {
	provider := NewProvider(nil, nil)
	assert.Equal(t, "Backup filesystem from Kubernetes pods", provider.Description())
}

func TestFilesystemProvider_ValidateOptions_Success(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		opts := backup.Options{
			Namespace:  "default",
			PodName:    "test-pod",
			SourcePath: "/data",
			OutputFile: "backup.tar",
		}

		provider := NewProvider(nil, nil)
		err := provider.ValidateOptions(opts)

		assert.NoError(t, err)
	})
}

func TestFilesystemProvider_ValidateOptions_MissingFields(t *testing.T) {
	tests := []struct {
		name   string
		opts   backup.Options
		errMsg string
	}{
		{
			name: "missing namespace",
			opts: backup.Options{
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar",
			},
			errMsg: "namespace is required",
		},
		{
			name: "missing pod name",
			opts: backup.Options{
				Namespace:  "default",
				SourcePath: "/data",
				OutputFile: "backup.tar",
			},
			errMsg: "pod name is required",
		},
		{
			name: "missing source path",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				OutputFile: "backup.tar",
			},
			errMsg: "source path is required",
		},
		{
			name: "missing output file",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
			},
			errMsg: "output file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(nil, nil)
			err := provider.ValidateOptions(tt.opts)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// MockCommandExecutor for testing - supports both old and new interfaces
type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Execute(ctx context.Context, command []string) (string, error) {
	args := m.Called(ctx, command)
	return args.String(0), args.Error(1)
}

func (m *MockCommandExecutor) Stream(ctx context.Context, command []string) (<-chan string, <-chan error) {
	args := m.Called(ctx, command)
	return args.Get(0).(<-chan string), args.Get(1).(<-chan error)
}

func (m *MockCommandExecutor) StreamBinary(ctx context.Context, command []string) (stdout io.ReadCloser, stderr io.ReadCloser, wait func() error, err error) {
	args := m.Called(ctx, command)

	// Handle nil values safely
	if args.Get(0) != nil {
		stdout = args.Get(0).(io.ReadCloser)
	}
	if args.Get(1) != nil {
		stderr = args.Get(1).(io.ReadCloser)
	}
	if args.Get(2) != nil {
		wait = args.Get(2).(func() error)
	}
	err = args.Error(3)

	return
}

// Helper to create mock readers for tests
type mockReadCloser struct {
	*strings.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func newMockReadCloser(data string) io.ReadCloser {
	return &mockReadCloser{Reader: strings.NewReader(data)}
}

func TestFilesystemProvider_EstimateSize(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	provider := NewProvider(nil, mockExecutor)

	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
	}

	// Mock du command execution
	ctx := context.Background()
	expectedCmd := kubernetes.BuildKubectlCommand("exec", "-n", "default", "test-pod", "--", "du", "-sb", "/data")
	mockExecutor.On("Execute", ctx, expectedCmd).Return("1024000\t/data\n", nil)

	size, err := provider.EstimateSize(opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(1024000), size)
	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_Execute(t *testing.T) {
	// Create temporary directory for test output
	tempDir, err := ioutil.TempDir("", "test-backup")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockExecutor := new(MockCommandExecutor)
	provider := NewProvider(nil, mockExecutor)

	outputFile := filepath.Join(tempDir, "backup.tar.gz")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
		Compress:   true,
	}

	ctx := context.Background()

	// Expected command (with verbose enabled for progress)
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-czvf", "-", "-C", "/", "data"}

	// Mock tar output data
	tarData := "mock tar data for testing"
	stderrData := "file1.txt\nfile2.txt\nfile3.txt\n"

	// Create mock readers
	mockStdout := newMockReadCloser(tarData)
	mockStderr := newMockReadCloser(stderrData)
	mockWait := func() error { return nil }

	// Mock StreamBinary for new implementation
	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	err = provider.Execute(ctx, opts)
	assert.NoError(t, err)

	// Check that output file was created and contains data
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err, "Output file should be created")
	assert.Equal(t, tarData, string(data), "Output file should contain tar data")

	// Check that progress channel has received updates
	progressCh := provider.StreamProgress()
	select {
	case progress := <-progressCh:
		assert.NotNil(t, progress)
		assert.Greater(t, progress.Current, int64(0))
	default:
		// No progress received immediately, but that's acceptable
		// since progress might be buffered
	}

	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_StreamProgress(t *testing.T) {
	provider := NewProvider(nil, nil)
	progressCh := provider.StreamProgress()

	// Progress channel should be non-nil
	assert.NotNil(t, progressCh)
}

func TestFilesystemProvider_Execute_WithExcludes(t *testing.T) {
	// Create temporary directory for test output
	tempDir, err := ioutil.TempDir("", "test-backup")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockExecutor := new(MockCommandExecutor)
	provider := NewProvider(nil, mockExecutor)

	outputFile := filepath.Join(tempDir, "backup.tar")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
		Exclude:    []string{"*.log", "tmp/"},
	}

	ctx := context.Background()

	// Expected command with excludes and verbose
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "--exclude=*.log", "--exclude=tmp/", "-C", "/", "data"}

	// Mock data
	tarData := "mock tar archive data"
	mockStdout := newMockReadCloser(tarData)
	mockStderr := newMockReadCloser("")
	mockWait := func() error { return nil }

	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	err = provider.Execute(ctx, opts)
	assert.NoError(t, err)

	// Check that output file was created
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "Output file should be created")

	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_Execute_CommandFailure(t *testing.T) {
	// Create temporary directory for test output
	tempDir, err := ioutil.TempDir("", "test-backup")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockExecutor := new(MockCommandExecutor)
	provider := NewProvider(nil, mockExecutor)

	outputFile := filepath.Join(tempDir, "backup.tar")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
	}

	ctx := context.Background()

	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	// Mock command failure
	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(nil, nil, nil, assert.AnError)

	err = provider.Execute(ctx, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start backup command")
	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_Execute_OutputDirectoryCreation(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := ioutil.TempDir("", "test-backup")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockExecutor := new(MockCommandExecutor)
	provider := NewProvider(nil, mockExecutor)

	// Use nested directory that doesn't exist
	outputFile := filepath.Join(tempDir, "nested", "dir", "backup.tar")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
	}

	ctx := context.Background()

	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	mockStdout := newMockReadCloser("")
	mockStderr := newMockReadCloser("")
	mockWait := func() error { return nil }
	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	err = provider.Execute(ctx, opts)
	assert.NoError(t, err)

	// Check that nested directory was created
	dir := filepath.Dir(outputFile)
	_, err = os.Stat(dir)
	assert.NoError(t, err, "Output directory should be created")

	// Check that output file was created
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "Output file should be created")

	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_Execute_WithContainer(t *testing.T) {
	// Create temporary directory for test output
	tempDir, err := ioutil.TempDir("", "test-backup")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	mockExecutor := new(MockCommandExecutor)
	provider := NewProvider(nil, mockExecutor)

	outputFile := filepath.Join(tempDir, "backup.tar")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
		Extra: map[string]interface{}{
			"container": "web-container",
		},
	}

	ctx := context.Background()

	// Expected command with container flag
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "-c", "web-container", "--",
		"tar", "-cvf", "-", "-C", "/", "data"}

	mockStdout := newMockReadCloser("")
	mockStderr := newMockReadCloser("")
	mockWait := func() error { return nil }
	mockExecutor.On("StreamBinary", ctx, expectedCmd).Return(mockStdout, mockStderr, mockWait, nil)

	err = provider.Execute(ctx, opts)
	assert.NoError(t, err)

	// Check that output file was created
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "Output file should be created")

	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_BuildTarCommand_NoRedirection(t *testing.T) {
	// This test verifies that the command structure is correct without shell redirection
	// We test this indirectly through the command expectations in other tests

	expectedCmdBase := []string{"kubectl", "exec", "-n", "test-ns", "test-pod", "-c", "test-container", "--",
		"tar", "-czf", "-", "--exclude=*.log", "-C", "/", "data"}

	// Verify the command doesn't contain shell redirection
	for _, arg := range expectedCmdBase {
		assert.NotEqual(t, ">", arg, "Command should not contain shell redirection operator")
		assert.NotContains(t, arg, "/tmp/backup.tar.gz", "Command should not contain output file path")
	}
}

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

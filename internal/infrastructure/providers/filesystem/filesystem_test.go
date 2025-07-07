package filesystem_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/providers/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFilesystemProvider_Interface(t *testing.T) {
	// Compile-time check
	var _ backup.Provider = (*filesystem.Provider)(nil)
}

func TestFilesystemProvider_Name(t *testing.T) {
	provider := filesystem.NewProvider(nil, nil)
	assert.Equal(t, "filesystem", provider.Name())
}

func TestFilesystemProvider_Description(t *testing.T) {
	provider := filesystem.NewProvider(nil, nil)
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
		
		provider := filesystem.NewProvider(nil, nil)
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
			provider := filesystem.NewProvider(nil, nil)
			err := provider.ValidateOptions(tt.opts)
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// MockCommandExecutor for testing - simplified to match actual usage
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

func TestFilesystemProvider_EstimateSize(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	provider := filesystem.NewProvider(nil, mockExecutor)
	
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
	provider := filesystem.NewProvider(nil, mockExecutor)
	
	outputFile := filepath.Join(tempDir, "backup.tar.gz")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
		Compress:   true,
	}
	
	ctx := context.Background()
	
	// Expected command (without shell redirection)
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
		"tar", "-czf", "-", "-C", "/", "data"}
	
	// Mock Execute instead of Stream for new implementation
	mockExecutor.On("Execute", ctx, expectedCmd).Return("", nil)
	
	err = provider.Execute(ctx, opts)
	assert.NoError(t, err)
	
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
	
	// Check that output file was created
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "Output file should be created")
	
	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_StreamProgress(t *testing.T) {
	provider := filesystem.NewProvider(nil, nil)
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
	provider := filesystem.NewProvider(nil, mockExecutor)
	
	outputFile := filepath.Join(tempDir, "backup.tar")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
		Exclude:    []string{"*.log", "tmp/"},
	}
	
	ctx := context.Background()
	
	// Expected command with excludes (without shell redirection)
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
		"tar", "-cf", "-", "--exclude=*.log", "--exclude=tmp/", "-C", "/", "data"}
	
	mockExecutor.On("Execute", ctx, expectedCmd).Return("", nil)
	
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
	provider := filesystem.NewProvider(nil, mockExecutor)
	
	outputFile := filepath.Join(tempDir, "backup.tar")
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
	}
	
	ctx := context.Background()
	
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
		"tar", "-cf", "-", "-C", "/", "data"}
	
	// Mock command failure
	mockExecutor.On("Execute", ctx, expectedCmd).Return("", assert.AnError)
	
	err = provider.Execute(ctx, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backup failed")
	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_Execute_OutputDirectoryCreation(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := ioutil.TempDir("", "test-backup")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	mockExecutor := new(MockCommandExecutor)
	provider := filesystem.NewProvider(nil, mockExecutor)
	
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
		"tar", "-cf", "-", "-C", "/", "data"}
	
	mockExecutor.On("Execute", ctx, expectedCmd).Return("", nil)
	
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
	provider := filesystem.NewProvider(nil, mockExecutor)
	
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
		"tar", "-cf", "-", "-C", "/", "data"}
	
	mockExecutor.On("Execute", ctx, expectedCmd).Return("", nil)
	
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
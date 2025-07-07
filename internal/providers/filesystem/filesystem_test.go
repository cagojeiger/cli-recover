package filesystem_test

import (
	"context"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/providers/filesystem"
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

func TestFilesystemProvider_ValidateOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    backup.Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar",
			},
			wantErr: false,
		},
		{
			name: "missing namespace",
			opts: backup.Options{
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar",
			},
			wantErr: true,
			errMsg:  "namespace is required",
		},
		{
			name: "missing pod name",
			opts: backup.Options{
				Namespace:  "default",
				SourcePath: "/data",
				OutputFile: "backup.tar",
			},
			wantErr: true,
			errMsg:  "pod name is required",
		},
		{
			name: "missing source path",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				OutputFile: "backup.tar",
			},
			wantErr: true,
			errMsg:  "source path is required",
		},
		{
			name: "missing output file",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
			},
			wantErr: true,
			errMsg:  "output file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := filesystem.NewProvider(nil, nil)
			err := provider.ValidateOptions(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockCommandExecutor for testing
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
	mockExecutor := new(MockCommandExecutor)
	provider := filesystem.NewProvider(nil, mockExecutor)
	
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "/tmp/backup.tar.gz",
		Compress:   true,
	}
	
	ctx := context.Background()
	
	// Mock tar command
	outputCh := make(chan string)
	errorCh := make(chan error, 1)
	
	// Expected command
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
		"tar", "-czf", "-", "-C", "/", "data", ">", "/tmp/backup.tar.gz"}
	
	mockExecutor.On("Stream", ctx, expectedCmd).Return(
		(<-chan string)(outputCh), (<-chan error)(errorCh))
	
	// Simulate tar output and completion
	go func() {
		outputCh <- "tar: data/file1.txt"
		outputCh <- "tar: data/file2.txt" 
		outputCh <- "tar: data/dir1/"
		close(outputCh)
		close(errorCh) // No error
	}()
	
	// Start monitoring progress
	progressCh := provider.StreamProgress()
	progressReceived := false
	
	go func() {
		for range progressCh {
			progressReceived = true
		}
	}()
	
	err := provider.Execute(ctx, opts)
	assert.NoError(t, err)
	assert.True(t, progressReceived, "Should receive progress updates")
	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_StreamProgress(t *testing.T) {
	provider := filesystem.NewProvider(nil, nil)
	progressCh := provider.StreamProgress()
	
	// Progress channel should be non-nil
	assert.NotNil(t, progressCh)
}

func TestFilesystemProvider_Execute_WithExcludes(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	provider := filesystem.NewProvider(nil, mockExecutor)
	
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar",
		Exclude:    []string{"*.log", "tmp/"},
	}
	
	ctx := context.Background()
	
	// Mock channels
	outputCh := make(chan string)
	errorCh := make(chan error, 1)
	
	// Expected command with excludes
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
		"tar", "-cf", "-", "--exclude=*.log", "--exclude=tmp/", "-C", "/", "data", ">", "backup.tar"}
	
	mockExecutor.On("Stream", ctx, expectedCmd).Return(
		(<-chan string)(outputCh), (<-chan error)(errorCh))
	
	go func() {
		close(outputCh)
		close(errorCh)
	}()
	
	err := provider.Execute(ctx, opts)
	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
}

func TestFilesystemProvider_Execute_Cancellation(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	provider := filesystem.NewProvider(nil, mockExecutor)
	
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar",
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Mock channels
	outputCh := make(chan string)
	errorCh := make(chan error, 1)
	
	expectedCmd := []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
		"tar", "-cf", "-", "-C", "/", "data", ">", "backup.tar"}
	
	mockExecutor.On("Stream", ctx, expectedCmd).Return(
		(<-chan string)(outputCh), (<-chan error)(errorCh))
	
	// Cancel context immediately
	cancel()
	
	// Don't close channels to simulate ongoing operation
	
	err := provider.Execute(ctx, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}
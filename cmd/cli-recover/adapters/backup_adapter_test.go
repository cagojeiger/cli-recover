package adapters

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

// NoOpLogger for testing
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(msg string, fields ...logger.Field) {}
func (n *NoOpLogger) Info(msg string, fields ...logger.Field) {}
func (n *NoOpLogger) Warn(msg string, fields ...logger.Field) {}
func (n *NoOpLogger) Error(msg string, fields ...logger.Field) {}
func (n *NoOpLogger) Fatal(msg string, fields ...logger.Field) {}
func (n *NoOpLogger) WithContext(ctx context.Context) logger.Logger { return n }
func (n *NoOpLogger) WithField(key string, value interface{}) logger.Logger { return n }
func (n *NoOpLogger) WithFields(fields ...logger.Field) logger.Logger { return n }
func (n *NoOpLogger) SetLevel(level logger.Level) {}
func (n *NoOpLogger) GetLevel() logger.Level { return logger.InfoLevel }

// MockProvider for testing
type MockProvider struct {
	mock.Mock
	progressCh chan backup.Progress
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		progressCh: make(chan backup.Progress, 100),
	}
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Execute(ctx context.Context, opts backup.Options) error {
	args := m.Called(ctx, opts)
	// Simulate some progress
	m.progressCh <- backup.Progress{Current: 50, Total: 100, Message: "Processing..."}
	m.progressCh <- backup.Progress{Current: 100, Total: 100, Message: "Complete"}
	close(m.progressCh)
	return args.Error(0)
}

func (m *MockProvider) EstimateSize(opts backup.Options) (int64, error) {
	args := m.Called(opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProvider) StreamProgress() <-chan backup.Progress {
	return m.progressCh
}

func (m *MockProvider) ValidateOptions(opts backup.Options) error {
	args := m.Called(opts)
	return args.Error(0)
}

func TestBackupAdapter_ExecuteBackup(t *testing.T) {
	// Create test registry
	testRegistry := backup.NewRegistry()
	
	// Create and register mock provider
	mockProvider := NewMockProvider()
	mockProvider.On("ValidateOptions", mock.Anything).Return(nil)
	mockProvider.On("EstimateSize", mock.Anything).Return(int64(1024*1024), nil)
	mockProvider.On("Execute", mock.Anything, mock.Anything).Return(nil)
	
	testRegistry.RegisterFactory("test", func() backup.Provider {
		return mockProvider
	})
	
	// Create adapter with test registry
	adapter := &BackupAdapter{
		registry: testRegistry,
		logger:   &NoOpLogger{},
	}
	
	// Create test command
	cmd := &cobra.Command{}
	cmd.Flags().String("namespace", "default", "")
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("verbose", false, "")
	
	// Test dry-run
	t.Run("dry-run", func(t *testing.T) {
		cmd.Flags().Set("dry-run", "true")
		
		err := adapter.ExecuteBackup("test", cmd, []string{})
		assert.NoError(t, err)
		
		// Should not call Execute in dry-run mode
		mockProvider.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	})
	
	// Test normal execution
	t.Run("normal execution", func(t *testing.T) {
		cmd.Flags().Set("dry-run", "false")
		
		// Create a temporary output file
		tmpFile, err := os.CreateTemp("", "test-backup-*.tar")
		assert.NoError(t, err)
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())
		
		// Mock buildOptions by using test provider name
		// The adapter will use default buildOptions which we'll test separately
		
		err = adapter.ExecuteBackup("test", cmd, []string{})
		assert.NoError(t, err)
		
		// Verify provider methods were called
		mockProvider.AssertCalled(t, "ValidateOptions", mock.Anything)
		mockProvider.AssertCalled(t, "EstimateSize", mock.Anything)
		mockProvider.AssertCalled(t, "Execute", mock.Anything, mock.Anything)
	})
}

func TestBackupAdapter_buildOptions(t *testing.T) {
	adapter := NewBackupAdapter()
	
	// Create test command with filesystem flags
	cmd := &cobra.Command{}
	cmd.Flags().String("namespace", "test-ns", "")
	cmd.Flags().String("compression", "gzip", "")
	cmd.Flags().StringSlice("exclude", []string{}, "")
	cmd.Flags().Bool("exclude-vcs", false, "")
	cmd.Flags().String("container", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("totals", false, "")
	cmd.Flags().Bool("preserve-perms", false, "")
	
	t.Run("filesystem provider", func(t *testing.T) {
		args := []string{"test-pod", "/data"}
		
		opts, err := adapter.buildOptions("filesystem", cmd, args)
		assert.NoError(t, err)
		
		assert.Equal(t, "test-ns", opts.Namespace)
		assert.Equal(t, "test-pod", opts.PodName)
		assert.Equal(t, "/data", opts.SourcePath)
		assert.True(t, opts.Compress)
		assert.Equal(t, "gzip", opts.Extra["compression"])
		
		// Output file should be auto-generated
		assert.Contains(t, opts.OutputFile, "backup-test-ns-test-pod-data")
		assert.Contains(t, opts.OutputFile, ".tar.gz")
	})
	
	t.Run("filesystem with custom output", func(t *testing.T) {
		cmd.Flags().Set("output", "custom-backup.tar")
		args := []string{"test-pod", "/data"}
		
		opts, err := adapter.buildOptions("filesystem", cmd, args)
		assert.NoError(t, err)
		
		assert.Equal(t, "custom-backup.tar", opts.OutputFile)
	})
	
	t.Run("filesystem with exclude patterns", func(t *testing.T) {
		cmd.Flags().Set("exclude", "*.log,*.tmp")
		cmd.Flags().Set("exclude-vcs", "true")
		args := []string{"test-pod", "/data"}
		
		opts, err := adapter.buildOptions("filesystem", cmd, args)
		assert.NoError(t, err)
		
		assert.Contains(t, opts.Exclude, "*.log")
		assert.Contains(t, opts.Exclude, "*.tmp")
		assert.Contains(t, opts.Exclude, ".git")
		assert.Contains(t, opts.Exclude, ".svn")
	})
	
	t.Run("unknown provider", func(t *testing.T) {
		_, err := adapter.buildOptions("unknown", cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown provider")
	})
	
	t.Run("insufficient args", func(t *testing.T) {
		_, err := adapter.buildOptions("filesystem", cmd, []string{"pod-only"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires [pod] [path] arguments")
	})
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/", "root"},
		{"/data", "data"},
		{"/var/log", "var-log"},
		{"/path with spaces", "path-with-spaces"},
		{"/file.name.txt", "file-name-txt"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		compression string
		expected    string
	}{
		{"gzip", ".tar.gz"},
		{"bzip2", ".tar.bz2"},
		{"xz", ".tar.xz"},
		{"none", ".tar"},
		{"unknown", ".tar.gz"},
	}
	
	for _, tt := range tests {
		t.Run(tt.compression, func(t *testing.T) {
			result := getFileExtension(tt.compression)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := humanizeBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
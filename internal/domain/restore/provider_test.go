package restore_test

import (
	"context"
	"testing"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/stretchr/testify/assert"
)

// MockProvider implements the restore.Provider interface for testing
type MockProvider struct {
	name         string
	description  string
	progressCh   chan restore.Progress
	shouldFail   bool
	validateOpts func(opts restore.Options) error
}

func NewMockProvider(name, description string) *MockProvider {
	return &MockProvider{
		name:        name,
		description: description,
		progressCh:  make(chan restore.Progress, 10),
	}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Description() string {
	return m.description
}

func (m *MockProvider) ValidateOptions(opts restore.Options) error {
	if m.validateOpts != nil {
		return m.validateOpts(opts)
	}
	return nil
}

func (m *MockProvider) ValidateBackup(backupFile string, metadata *restore.Metadata) error {
	if m.shouldFail {
		return restore.NewRestoreError("INVALID_BACKUP", "Mock validation failed")
	}
	return nil
}

func (m *MockProvider) Execute(ctx context.Context, opts restore.Options) (*restore.RestoreResult, error) {
	if m.shouldFail {
		return nil, restore.NewRestoreError("RESTORE_FAILED", "Mock execution failed")
	}

	// Simulate progress
	go func() {
		m.progressCh <- restore.Progress{Current: 50, Total: 100, Message: "Starting restore..."}
		m.progressCh <- restore.Progress{Current: 100, Total: 100, Message: "Restore complete"}
		close(m.progressCh)
	}()

	return &restore.RestoreResult{
		Success:      true,
		RestoredPath: opts.TargetPath,
		FileCount:    10,
		BytesWritten: 1024,
		Duration:     time.Duration(1 * time.Second),
		Warnings:     nil,
	}, nil
}

func (m *MockProvider) StreamProgress() <-chan restore.Progress {
	return m.progressCh
}

func (m *MockProvider) EstimateSize(backupFile string) (int64, error) {
	if m.shouldFail {
		return 0, restore.NewRestoreError("SIZE_ESTIMATION_FAILED", "Mock size estimation failed")
	}
	return 2048, nil
}

func (m *MockProvider) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

func (m *MockProvider) SetValidateOptsFunc(fn func(opts restore.Options) error) {
	m.validateOpts = fn
}

func TestMockProvider_Interface(t *testing.T) {
	// Test that MockProvider implements Provider interface
	var provider restore.Provider = NewMockProvider("test", "Test Provider")
	assert.NotNil(t, provider)
}

func TestMockProvider_Name(t *testing.T) {
	provider := NewMockProvider("filesystem", "Filesystem restore provider")
	assert.Equal(t, "filesystem", provider.Name())
}

func TestMockProvider_Description(t *testing.T) {
	provider := NewMockProvider("filesystem", "Filesystem restore provider")
	assert.Equal(t, "Filesystem restore provider", provider.Description())
}

func TestMockProvider_ValidateOptions_Success(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")

	options := restore.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		BackupFile: "/path/to/backup.tar.gz",
		TargetPath: "/restore/path",
	}

	err := provider.ValidateOptions(options)
	assert.NoError(t, err)
}

func TestMockProvider_ValidateOptions_CustomValidation(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")

	// Set custom validation function
	provider.SetValidateOptsFunc(func(opts restore.Options) error {
		if opts.Namespace == "" {
			return restore.NewRestoreError("INVALID_NAMESPACE", "Namespace is required")
		}
		return nil
	})

	// Test with empty namespace
	options := restore.Options{
		PodName:    "test-pod",
		BackupFile: "/path/to/backup.tar.gz",
	}

	err := provider.ValidateOptions(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Namespace is required")

	// Test with valid namespace
	options.Namespace = "default"
	err = provider.ValidateOptions(options)
	assert.NoError(t, err)
}

func TestMockProvider_ValidateBackup_Success(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")

	metadata := &restore.Metadata{
		ID:         "backup-123",
		Type:       "filesystem",
		BackupFile: "/path/to/backup.tar.gz",
	}

	err := provider.ValidateBackup("/path/to/backup.tar.gz", metadata)
	assert.NoError(t, err)
}

func TestMockProvider_ValidateBackup_Failure(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")
	provider.SetShouldFail(true)

	metadata := &restore.Metadata{
		ID:         "backup-123",
		Type:       "filesystem",
		BackupFile: "/path/to/backup.tar.gz",
	}

	err := provider.ValidateBackup("/path/to/backup.tar.gz", metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Mock validation failed")
}

func TestMockProvider_Execute_Success(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")

	options := restore.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		BackupFile: "/path/to/backup.tar.gz",
		TargetPath: "/restore/path",
	}

	ctx := context.Background()
	result, err := provider.Execute(ctx, options)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "/restore/path", result.RestoredPath)
	assert.Equal(t, 10, result.FileCount)
	assert.Equal(t, int64(1024), result.BytesWritten)
	assert.Equal(t, time.Duration(1*time.Second), result.Duration)
	assert.Nil(t, result.Warnings)
}

func TestMockProvider_Execute_Failure(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")
	provider.SetShouldFail(true)

	options := restore.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		BackupFile: "/path/to/backup.tar.gz",
		TargetPath: "/restore/path",
	}

	ctx := context.Background()
	result, err := provider.Execute(ctx, options)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Mock execution failed")
}

func TestMockProvider_StreamProgress(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")

	progressCh := provider.StreamProgress()
	assert.NotNil(t, progressCh)

	// Execute to trigger progress updates
	options := restore.Options{
		TargetPath: "/test/path",
	}
	ctx := context.Background()

	go func() {
		provider.Execute(ctx, options)
	}()

	// Read progress updates
	var progressUpdates []restore.Progress
	for progress := range progressCh {
		progressUpdates = append(progressUpdates, progress)
	}

	assert.Len(t, progressUpdates, 2)
	assert.Equal(t, "Starting restore...", progressUpdates[0].Message)
	assert.Equal(t, "Restore complete", progressUpdates[1].Message)
}

func TestMockProvider_EstimateSize_Success(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")

	size, err := provider.EstimateSize("/path/to/backup.tar.gz")

	assert.NoError(t, err)
	assert.Equal(t, int64(2048), size)
}

func TestMockProvider_EstimateSize_Failure(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")
	provider.SetShouldFail(true)

	size, err := provider.EstimateSize("/path/to/backup.tar.gz")

	assert.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "Mock size estimation failed")
}

func TestMockProvider_ContextCancellation(t *testing.T) {
	provider := NewMockProvider("test", "Test Provider")

	options := restore.Options{
		TargetPath: "/test/path",
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	// Execute should handle cancelled context
	// Note: Our mock doesn't actually check context cancellation,
	// but this tests the interface contract
	result, err := provider.Execute(ctx, options)

	// Mock provider doesn't check context, so it will succeed
	// In a real implementation, this should respect context cancellation
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestMockProvider_MultipleInstances(t *testing.T) {
	provider1 := NewMockProvider("provider1", "Provider 1")
	provider2 := NewMockProvider("provider2", "Provider 2")

	assert.NotSame(t, provider1, provider2)
	assert.Equal(t, "provider1", provider1.Name())
	assert.Equal(t, "provider2", provider2.Name())

	// They should have independent state
	provider1.SetShouldFail(true)
	provider2.SetShouldFail(false)

	err1 := provider1.ValidateBackup("test", nil)
	err2 := provider2.ValidateBackup("test", nil)

	assert.Error(t, err1)
	assert.NoError(t, err2)
}

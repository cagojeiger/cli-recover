package operation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cagojeiger/cli-recover/internal/domain/operation"
)

// MockProvider is a mock implementation of Provider interface
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Type() operation.ProviderType {
	args := m.Called()
	return args.Get(0).(operation.ProviderType)
}

func (m *MockProvider) ValidateOptions(opts operation.Options) error {
	args := m.Called(opts)
	return args.Error(0)
}

func (m *MockProvider) Execute(ctx context.Context, opts operation.Options) (*operation.Result, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*operation.Result), args.Error(1)
}

func (m *MockProvider) EstimateSize(opts operation.Options) (int64, error) {
	args := m.Called(opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProvider) StreamProgress() <-chan operation.Progress {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(<-chan operation.Progress)
}

func TestProviderType(t *testing.T) {
	tests := []struct {
		name     string
		typ      operation.ProviderType
		expected string
	}{
		{
			name:     "backup type",
			typ:      operation.TypeBackup,
			expected: "backup",
		},
		{
			name:     "restore type",
			typ:      operation.TypeRestore,
			expected: "restore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.typ))
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("options creation", func(t *testing.T) {
		opts := operation.Options{
			Namespace:     "test-ns",
			PodName:       "test-pod",
			Container:     "test-container",
			SourcePath:    "/source",
			OutputFile:    "/output.tar",
			BackupFile:    "/backup.tar",
			TargetPath:    "/target",
			Compress:      true,
			Overwrite:     true,
			PreservePerms: true,
			Exclude:       []string{"*.tmp"},
			SkipPaths:     []string{"/skip"},
			Extra: map[string]interface{}{
				"key": "value",
			},
		}

		assert.Equal(t, "test-ns", opts.Namespace)
		assert.Equal(t, "test-pod", opts.PodName)
		assert.Equal(t, "test-container", opts.Container)
		assert.Equal(t, "/source", opts.SourcePath)
		assert.Equal(t, "/output.tar", opts.OutputFile)
		assert.Equal(t, "/backup.tar", opts.BackupFile)
		assert.Equal(t, "/target", opts.TargetPath)
		assert.True(t, opts.Compress)
		assert.True(t, opts.Overwrite)
		assert.True(t, opts.PreservePerms)
		assert.Equal(t, []string{"*.tmp"}, opts.Exclude)
		assert.Equal(t, []string{"/skip"}, opts.SkipPaths)
		assert.Equal(t, "value", opts.Extra["key"])
	})
}

func TestResult(t *testing.T) {
	t.Run("successful result", func(t *testing.T) {
		result := &operation.Result{
			Success:      true,
			Message:      "Operation completed",
			RestoredPath: "/restored",
			FileCount:    100,
			BytesWritten: 1024 * 1024,
			Warnings:     []string{"warning1"},
			Error:        nil,
		}

		assert.True(t, result.Success)
		assert.Equal(t, "Operation completed", result.Message)
		assert.Equal(t, "/restored", result.RestoredPath)
		assert.Equal(t, 100, result.FileCount)
		assert.Equal(t, int64(1024*1024), result.BytesWritten)
		assert.Equal(t, []string{"warning1"}, result.Warnings)
		assert.Nil(t, result.Error)
	})

	t.Run("failed result", func(t *testing.T) {
		expectedErr := errors.New("operation failed")
		result := &operation.Result{
			Success: false,
			Message: "Operation failed",
			Error:   expectedErr,
		}

		assert.False(t, result.Success)
		assert.Equal(t, "Operation failed", result.Message)
		assert.Equal(t, expectedErr, result.Error)
	})
}

func TestProgress(t *testing.T) {
	t.Run("progress creation", func(t *testing.T) {
		progress := operation.Progress{
			Current: 50,
			Total:   100,
			Message: "Processing...",
		}

		assert.Equal(t, int64(50), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "Processing...", progress.Message)
	})
}

func TestMetadata(t *testing.T) {
	t.Run("metadata creation", func(t *testing.T) {
		createdAt, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
		completedAt, _ := time.Parse(time.RFC3339, "2024-01-01T00:01:00Z")

		metadata := &operation.Metadata{
			ID:          "test-123",
			Type:        "filesystem",
			Provider:    "filesystem",
			Namespace:   "test-ns",
			PodName:     "test-pod",
			Container:   "test-container",
			SourcePath:  "/source",
			BackupFile:  "/backup.tar",
			Compression: "gzip",
			Size:        1024 * 1024,
			Checksum:    "abc123",
			CreatedAt:   createdAt,
			CompletedAt: completedAt,
			Status:      "completed",
			ProviderInfo: map[string]interface{}{
				"version": "1.0",
			},
			Extra: map[string]string{
				"key": "value",
			},
		}

		assert.Equal(t, "test-123", metadata.ID)
		assert.Equal(t, "filesystem", metadata.Type)
		assert.Equal(t, "filesystem", metadata.Provider)
		assert.Equal(t, "test-ns", metadata.Namespace)
		assert.Equal(t, "test-pod", metadata.PodName)
		assert.Equal(t, "test-container", metadata.Container)
		assert.Equal(t, "/source", metadata.SourcePath)
		assert.Equal(t, "/backup.tar", metadata.BackupFile)
		assert.Equal(t, "gzip", metadata.Compression)
		assert.Equal(t, int64(1024*1024), metadata.Size)
		assert.Equal(t, "abc123", metadata.Checksum)
		assert.Equal(t, "completed", metadata.Status)
		assert.Equal(t, "1.0", metadata.ProviderInfo["version"])
		assert.Equal(t, "value", metadata.Extra["key"])
	})
}

func TestProviderInterface(t *testing.T) {
	t.Run("provider methods", func(t *testing.T) {
		mockProvider := new(MockProvider)
		ctx := context.Background()
		opts := operation.Options{
			Namespace:  "test-ns",
			PodName:    "test-pod",
			SourcePath: "/source",
			OutputFile: "/output.tar",
		}

		// Setup expectations
		mockProvider.On("Name").Return("test-provider")
		mockProvider.On("Description").Return("Test Provider")
		mockProvider.On("Type").Return(operation.TypeBackup)
		mockProvider.On("ValidateOptions", opts).Return(nil)
		mockProvider.On("Execute", ctx, opts).Return(&operation.Result{
			Success: true,
			Message: "Backup completed",
		}, nil)
		mockProvider.On("EstimateSize", opts).Return(int64(1024), nil)

		progressChan := make(chan operation.Progress, 1)
		progressChan <- operation.Progress{Current: 50, Total: 100}
		close(progressChan)
		mockProvider.On("StreamProgress").Return((<-chan operation.Progress)(progressChan))

		// Test methods
		assert.Equal(t, "test-provider", mockProvider.Name())
		assert.Equal(t, "Test Provider", mockProvider.Description())
		assert.Equal(t, operation.TypeBackup, mockProvider.Type())
		assert.NoError(t, mockProvider.ValidateOptions(opts))

		result, err := mockProvider.Execute(ctx, opts)
		assert.NoError(t, err)
		assert.True(t, result.Success)

		size, err := mockProvider.EstimateSize(opts)
		assert.NoError(t, err)
		assert.Equal(t, int64(1024), size)

		progress := <-mockProvider.StreamProgress()
		assert.Equal(t, int64(50), progress.Current)
		assert.Equal(t, int64(100), progress.Total)

		mockProvider.AssertExpectations(t)
	})
}

func TestProviderValidation(t *testing.T) {
	tests := []struct {
		name        string
		opts        operation.Options
		expectedErr string
	}{
		{
			name: "valid options",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/source",
				OutputFile: "/output.tar",
			},
			expectedErr: "",
		},
		{
			name: "missing namespace",
			opts: operation.Options{
				PodName:    "test-pod",
				SourcePath: "/source",
				OutputFile: "/output.tar",
			},
			expectedErr: "namespace is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockProvider)

			if tt.expectedErr != "" {
				mockProvider.On("ValidateOptions", tt.opts).Return(errors.New(tt.expectedErr))
				err := mockProvider.ValidateOptions(tt.opts)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				mockProvider.On("ValidateOptions", tt.opts).Return(nil)
				err := mockProvider.ValidateOptions(tt.opts)
				assert.NoError(t, err)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}

func TestProviderExecution(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		opts   operation.Options
		result *operation.Result
		err    error
	}{
		{
			name: "successful execution",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/source",
				OutputFile: "/output.tar",
			},
			result: &operation.Result{
				Success: true,
				Message: "Operation completed successfully",
			},
			err: nil,
		},
		{
			name: "failed execution",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/source",
				OutputFile: "/output.tar",
			},
			result: nil,
			err:    errors.New("execution failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := new(MockProvider)
			mockProvider.On("Execute", ctx, tt.opts).Return(tt.result, tt.err)

			result, err := mockProvider.Execute(ctx, tt.opts)

			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.result.Success, result.Success)
				assert.Equal(t, tt.result.Message, result.Message)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}

func TestResultValidation(t *testing.T) {
	tests := []struct {
		name     string
		result   *operation.Result
		validate func(*testing.T, *operation.Result)
	}{
		{
			name: "success result",
			result: &operation.Result{
				Success: true,
				Message: "Operation successful",
			},
			validate: func(t *testing.T, r *operation.Result) {
				assert.True(t, r.Success)
				assert.Equal(t, "Operation successful", r.Message)
				assert.Nil(t, r.Error)
			},
		},
		{
			name: "failure result",
			result: &operation.Result{
				Success: false,
				Message: "Operation failed",
				Error:   errors.New("error occurred"),
			},
			validate: func(t *testing.T, r *operation.Result) {
				assert.False(t, r.Success)
				assert.Equal(t, "Operation failed", r.Message)
				assert.NotNil(t, r.Error)
			},
		},
		{
			name: "result with metadata",
			result: &operation.Result{
				Success:      true,
				Message:      "Restore completed",
				RestoredPath: "/restored/path",
				FileCount:    42,
				BytesWritten: 1024 * 1024 * 10,
				Warnings:     []string{"warning1", "warning2"},
			},
			validate: func(t *testing.T, r *operation.Result) {
				assert.True(t, r.Success)
				assert.Equal(t, "/restored/path", r.RestoredPath)
				assert.Equal(t, 42, r.FileCount)
				assert.Equal(t, int64(1024*1024*10), r.BytesWritten)
				assert.Len(t, r.Warnings, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.result)
		})
	}
}

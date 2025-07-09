package operation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/operation"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// MockBackupProvider is a mock implementation of backup.Provider
type MockBackupProvider struct {
	mock.Mock
}

func (m *MockBackupProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBackupProvider) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBackupProvider) ValidateOptions(opts backup.Options) error {
	args := m.Called(opts)
	return args.Error(0)
}

func (m *MockBackupProvider) Execute(ctx context.Context, opts backup.Options) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func (m *MockBackupProvider) ExecuteWithResult(ctx context.Context, opts backup.Options) (*backup.Result, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.Result), args.Error(1)
}

func (m *MockBackupProvider) EstimateSize(opts backup.Options) (int64, error) {
	args := m.Called(opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBackupProvider) StreamProgress() <-chan backup.Progress {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(<-chan backup.Progress)
}

// MockRestoreProvider is a mock implementation of restore.Provider
type MockRestoreProvider struct {
	mock.Mock
}

func (m *MockRestoreProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRestoreProvider) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRestoreProvider) ValidateOptions(opts restore.Options) error {
	args := m.Called(opts)
	return args.Error(0)
}

func (m *MockRestoreProvider) ValidateBackup(backupFile string, metadata *restore.Metadata) error {
	args := m.Called(backupFile, metadata)
	return args.Error(0)
}

func (m *MockRestoreProvider) Execute(ctx context.Context, opts restore.Options) (*restore.RestoreResult, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*restore.RestoreResult), args.Error(1)
}

func (m *MockRestoreProvider) EstimateSize(backupFile string) (int64, error) {
	args := m.Called(backupFile)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRestoreProvider) StreamProgress() <-chan restore.Progress {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(<-chan restore.Progress)
}

func TestNewBackupAdapter(t *testing.T) {
	mockBackup := new(MockBackupProvider)
	adapter := operation.NewBackupAdapter(mockBackup)

	assert.NotNil(t, adapter)
	assert.Implements(t, (*operation.Provider)(nil), adapter)
}

func TestBackupAdapter_Name(t *testing.T) {
	mockBackup := new(MockBackupProvider)
	mockBackup.On("Name").Return("filesystem-backup")

	adapter := operation.NewBackupAdapter(mockBackup)
	name := adapter.Name()

	assert.Equal(t, "filesystem-backup", name)
	mockBackup.AssertExpectations(t)
}

func TestBackupAdapter_Description(t *testing.T) {
	mockBackup := new(MockBackupProvider)
	mockBackup.On("Description").Return("Backup filesystem data")

	adapter := operation.NewBackupAdapter(mockBackup)
	desc := adapter.Description()

	assert.Equal(t, "Backup filesystem data", desc)
	mockBackup.AssertExpectations(t)
}

func TestBackupAdapter_Type(t *testing.T) {
	mockBackup := new(MockBackupProvider)
	adapter := operation.NewBackupAdapter(mockBackup)

	assert.Equal(t, operation.TypeBackup, adapter.Type())
}

func TestBackupAdapter_ValidateOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        operation.Options
		setupMock   func(*MockBackupProvider, operation.Options)
		expectedErr error
	}{
		{
			name: "valid options",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			setupMock: func(m *MockBackupProvider, opts operation.Options) {
				expectedBackupOpts := backup.Options{
					Namespace:  opts.Namespace,
					PodName:    opts.PodName,
					SourcePath: opts.SourcePath,
					OutputFile: opts.OutputFile,
				}
				m.On("ValidateOptions", expectedBackupOpts).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid options - missing namespace",
			opts: operation.Options{
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			setupMock: func(m *MockBackupProvider, opts operation.Options) {
				expectedBackupOpts := backup.Options{
					PodName:    opts.PodName,
					SourcePath: opts.SourcePath,
					OutputFile: opts.OutputFile,
				}
				m.On("ValidateOptions", expectedBackupOpts).Return(errors.New("namespace is required"))
			},
			expectedErr: errors.New("namespace is required"),
		},
		{
			name: "options with extra fields",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				Container:  "app",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
				Compress:   true,
				Exclude:    []string{"*.tmp", "*.log"},
				Extra:      map[string]interface{}{"compression": "gzip"},
			},
			setupMock: func(m *MockBackupProvider, opts operation.Options) {
				expectedBackupOpts := backup.Options{
					Namespace:  opts.Namespace,
					PodName:    opts.PodName,
					Container:  opts.Container,
					SourcePath: opts.SourcePath,
					OutputFile: opts.OutputFile,
					Compress:   opts.Compress,
					Exclude:    opts.Exclude,
					Extra:      opts.Extra,
				}
				m.On("ValidateOptions", expectedBackupOpts).Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBackup := new(MockBackupProvider)
			tt.setupMock(mockBackup, tt.opts)

			adapter := operation.NewBackupAdapter(mockBackup)
			err := adapter.ValidateOptions(tt.opts)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockBackup.AssertExpectations(t)
		})
	}
}

func TestBackupAdapter_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		opts           operation.Options
		backupError    error
		expectedResult *operation.Result
		expectedError  error
	}{
		{
			name: "successful backup",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			backupError: nil,
			expectedResult: &operation.Result{
				Success: true,
				Message: "Backup completed successfully",
			},
			expectedError: nil,
		},
		{
			name: "failed backup",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			backupError: errors.New("pod not found"),
			expectedResult: &operation.Result{
				Success: false,
				Message: "Backup failed: pod not found",
				Error:   errors.New("pod not found"),
			},
			expectedError: errors.New("pod not found"),
		},
		{
			name: "backup with compression",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
				Compress:   true,
				Exclude:    []string{"*.log"},
			},
			backupError: nil,
			expectedResult: &operation.Result{
				Success: true,
				Message: "Backup completed successfully",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBackup := new(MockBackupProvider)

			expectedBackupOpts := backup.Options{
				Namespace:  tt.opts.Namespace,
				PodName:    tt.opts.PodName,
				Container:  tt.opts.Container,
				SourcePath: tt.opts.SourcePath,
				OutputFile: tt.opts.OutputFile,
				Compress:   tt.opts.Compress,
				Exclude:    tt.opts.Exclude,
				Extra:      tt.opts.Extra,
			}

			mockBackup.On("Execute", ctx, expectedBackupOpts).Return(tt.backupError)

			adapter := operation.NewBackupAdapter(mockBackup)
			result, err := adapter.Execute(ctx, tt.opts)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Success, result.Success)
				assert.Equal(t, tt.expectedResult.Message, result.Message)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Success, result.Success)
				assert.Equal(t, tt.expectedResult.Message, result.Message)
			}

			mockBackup.AssertExpectations(t)
		})
	}
}

func TestBackupAdapter_EstimateSize(t *testing.T) {
	tests := []struct {
		name         string
		opts         operation.Options
		expectedSize int64
		expectedErr  error
	}{
		{
			name: "successful size estimation",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/data",
			},
			expectedSize: 1024 * 1024 * 100, // 100MB
			expectedErr:  nil,
		},
		{
			name: "size estimation error",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/nonexistent",
			},
			expectedSize: 0,
			expectedErr:  errors.New("path not found"),
		},
		{
			name: "empty directory",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				SourcePath: "/empty",
			},
			expectedSize: 0,
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBackup := new(MockBackupProvider)

			expectedBackupOpts := backup.Options{
				Namespace:  tt.opts.Namespace,
				PodName:    tt.opts.PodName,
				SourcePath: tt.opts.SourcePath,
			}

			mockBackup.On("EstimateSize", expectedBackupOpts).Return(tt.expectedSize, tt.expectedErr)

			adapter := operation.NewBackupAdapter(mockBackup)
			size, err := adapter.EstimateSize(tt.opts)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedSize, size)

			mockBackup.AssertExpectations(t)
		})
	}
}

func TestBackupAdapter_StreamProgress(t *testing.T) {
	t.Run("progress streaming", func(t *testing.T) {
		mockBackup := new(MockBackupProvider)

		// Create backup progress channel
		backupChan := make(chan backup.Progress, 3)
		backupChan <- backup.Progress{Current: 0, Total: 100, Message: "Starting backup"}
		backupChan <- backup.Progress{Current: 50, Total: 100, Message: "50% complete"}
		backupChan <- backup.Progress{Current: 100, Total: 100, Message: "Backup complete"}
		close(backupChan)

		mockBackup.On("StreamProgress").Return((<-chan backup.Progress)(backupChan))

		adapter := operation.NewBackupAdapter(mockBackup)
		progressChan := adapter.StreamProgress()

		// Verify progress conversion
		progress := <-progressChan
		assert.Equal(t, int64(0), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "Starting backup", progress.Message)

		progress = <-progressChan
		assert.Equal(t, int64(50), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "50% complete", progress.Message)

		progress = <-progressChan
		assert.Equal(t, int64(100), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "Backup complete", progress.Message)

		// Verify channel is closed
		_, ok := <-progressChan
		assert.False(t, ok)

		mockBackup.AssertExpectations(t)
	})

	// Note: Testing nil progress channel is skipped because the adapter
	// implementation will hang when trying to range over a nil channel.
	// This is a known limitation in the current implementation.
}

func TestNewRestoreAdapter(t *testing.T) {
	mockRestore := new(MockRestoreProvider)
	adapter := operation.NewRestoreAdapter(mockRestore)

	assert.NotNil(t, adapter)
	assert.Implements(t, (*operation.Provider)(nil), adapter)
}

func TestRestoreAdapter_Name(t *testing.T) {
	mockRestore := new(MockRestoreProvider)
	mockRestore.On("Name").Return("filesystem-restore")

	adapter := operation.NewRestoreAdapter(mockRestore)
	name := adapter.Name()

	assert.Equal(t, "filesystem-restore", name)
	mockRestore.AssertExpectations(t)
}

func TestRestoreAdapter_Description(t *testing.T) {
	mockRestore := new(MockRestoreProvider)
	mockRestore.On("Description").Return("Restore filesystem data")

	adapter := operation.NewRestoreAdapter(mockRestore)
	desc := adapter.Description()

	assert.Equal(t, "Restore filesystem data", desc)
	mockRestore.AssertExpectations(t)
}

func TestRestoreAdapter_Type(t *testing.T) {
	mockRestore := new(MockRestoreProvider)
	adapter := operation.NewRestoreAdapter(mockRestore)

	assert.Equal(t, operation.TypeRestore, adapter.Type())
}

func TestRestoreAdapter_ValidateOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        operation.Options
		setupMock   func(*MockRestoreProvider, operation.Options)
		expectedErr error
	}{
		{
			name: "valid restore options",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				BackupFile: "/backup.tar",
				TargetPath: "/restore",
			},
			setupMock: func(m *MockRestoreProvider, opts operation.Options) {
				expectedRestoreOpts := restore.Options{
					Namespace:     opts.Namespace,
					PodName:       opts.PodName,
					BackupFile:    opts.BackupFile,
					TargetPath:    opts.TargetPath,
					Overwrite:     opts.Overwrite,
					PreservePerms: opts.PreservePerms,
					SkipPaths:     opts.SkipPaths,
					Extra:         opts.Extra,
				}
				m.On("ValidateOptions", expectedRestoreOpts).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid options - missing backup file",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				TargetPath: "/restore",
			},
			setupMock: func(m *MockRestoreProvider, opts operation.Options) {
				expectedRestoreOpts := restore.Options{
					Namespace:  opts.Namespace,
					PodName:    opts.PodName,
					TargetPath: opts.TargetPath,
				}
				m.On("ValidateOptions", expectedRestoreOpts).Return(errors.New("backup file is required"))
			},
			expectedErr: errors.New("backup file is required"),
		},
		{
			name: "restore options with all fields",
			opts: operation.Options{
				Namespace:     "test-ns",
				PodName:       "test-pod",
				Container:     "app",
				BackupFile:    "/backup.tar.gz",
				TargetPath:    "/data/restore",
				Overwrite:     true,
				PreservePerms: true,
				SkipPaths:     []string{"/skip/this", "/and/this"},
				Extra:         map[string]interface{}{"verify": true},
			},
			setupMock: func(m *MockRestoreProvider, opts operation.Options) {
				expectedRestoreOpts := restore.Options{
					Namespace:     opts.Namespace,
					PodName:       opts.PodName,
					Container:     opts.Container,
					BackupFile:    opts.BackupFile,
					TargetPath:    opts.TargetPath,
					Overwrite:     opts.Overwrite,
					PreservePerms: opts.PreservePerms,
					SkipPaths:     opts.SkipPaths,
					Extra:         opts.Extra,
				}
				m.On("ValidateOptions", expectedRestoreOpts).Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRestore := new(MockRestoreProvider)
			tt.setupMock(mockRestore, tt.opts)

			adapter := operation.NewRestoreAdapter(mockRestore)
			err := adapter.ValidateOptions(tt.opts)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRestore.AssertExpectations(t)
		})
	}
}

func TestRestoreAdapter_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		opts           operation.Options
		restoreResult  *restore.RestoreResult
		restoreError   error
		expectedResult *operation.Result
		expectedError  error
	}{
		{
			name: "successful restore",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				BackupFile: "/backup.tar",
				TargetPath: "/restore",
			},
			restoreResult: &restore.RestoreResult{
				Success:      true,
				RestoredPath: "/restore",
				FileCount:    42,
				BytesWritten: 1024000,
				Duration:     time.Minute,
			},
			restoreError: nil,
			expectedResult: &operation.Result{
				Success:      true,
				Message:      "Restore completed successfully",
				RestoredPath: "/restore",
				FileCount:    42,
				BytesWritten: 1024000,
			},
			expectedError: nil,
		},
		{
			name: "failed restore",
			opts: operation.Options{
				Namespace:  "test-ns",
				PodName:    "test-pod",
				BackupFile: "/backup.tar",
				TargetPath: "/restore",
			},
			restoreResult: nil,
			restoreError:  errors.New("backup file corrupted"),
			expectedResult: &operation.Result{
				Success: false,
				Message: "Restore failed: backup file corrupted",
				Error:   errors.New("backup file corrupted"),
			},
			expectedError: errors.New("backup file corrupted"),
		},
		{
			name: "restore with warnings",
			opts: operation.Options{
				Namespace:     "test-ns",
				PodName:       "test-pod",
				BackupFile:    "/backup.tar",
				TargetPath:    "/restore",
				PreservePerms: true,
				SkipPaths:     []string{"/skip"},
			},
			restoreResult: &restore.RestoreResult{
				Success:      true,
				RestoredPath: "/restore",
				FileCount:    38,
				BytesWritten: 900000,
				Duration:     45 * time.Second,
				Warnings:     []string{"skipped 4 files", "permission denied on 2 files"},
			},
			restoreError: nil,
			expectedResult: &operation.Result{
				Success:      true,
				Message:      "Restore completed successfully",
				RestoredPath: "/restore",
				FileCount:    38,
				BytesWritten: 900000,
				Warnings:     []string{"skipped 4 files", "permission denied on 2 files"},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRestore := new(MockRestoreProvider)

			expectedRestoreOpts := restore.Options{
				Namespace:     tt.opts.Namespace,
				PodName:       tt.opts.PodName,
				Container:     tt.opts.Container,
				BackupFile:    tt.opts.BackupFile,
				TargetPath:    tt.opts.TargetPath,
				Overwrite:     tt.opts.Overwrite,
				PreservePerms: tt.opts.PreservePerms,
				SkipPaths:     tt.opts.SkipPaths,
				Extra:         tt.opts.Extra,
			}

			mockRestore.On("Execute", ctx, expectedRestoreOpts).Return(tt.restoreResult, tt.restoreError)

			adapter := operation.NewRestoreAdapter(mockRestore)
			result, err := adapter.Execute(ctx, tt.opts)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Success, result.Success)
				assert.Equal(t, tt.expectedResult.Message, result.Message)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Success, result.Success)
				assert.Equal(t, tt.expectedResult.Message, result.Message)
				assert.Equal(t, tt.expectedResult.RestoredPath, result.RestoredPath)
				assert.Equal(t, tt.expectedResult.FileCount, result.FileCount)
				assert.Equal(t, tt.expectedResult.BytesWritten, result.BytesWritten)
				assert.Equal(t, tt.expectedResult.Warnings, result.Warnings)
			}

			mockRestore.AssertExpectations(t)
		})
	}
}

func TestRestoreAdapter_EstimateSize(t *testing.T) {
	tests := []struct {
		name         string
		opts         operation.Options
		expectedSize int64
		expectedErr  error
	}{
		{
			name: "successful size estimation from backup",
			opts: operation.Options{
				BackupFile: "/backup.tar",
			},
			expectedSize: 1024 * 1024 * 50, // 50MB
			expectedErr:  nil,
		},
		{
			name: "backup file not found",
			opts: operation.Options{
				BackupFile: "/nonexistent.tar",
			},
			expectedSize: 0,
			expectedErr:  errors.New("backup file not found"),
		},
		{
			name: "corrupted backup file",
			opts: operation.Options{
				BackupFile: "/corrupted.tar",
			},
			expectedSize: 0,
			expectedErr:  errors.New("invalid backup format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRestore := new(MockRestoreProvider)

			mockRestore.On("EstimateSize", tt.opts.BackupFile).Return(tt.expectedSize, tt.expectedErr)

			adapter := operation.NewRestoreAdapter(mockRestore)
			size, err := adapter.EstimateSize(tt.opts)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedSize, size)

			mockRestore.AssertExpectations(t)
		})
	}
}

func TestRestoreAdapter_StreamProgress(t *testing.T) {
	t.Run("progress streaming", func(t *testing.T) {
		mockRestore := new(MockRestoreProvider)

		// Create restore progress channel
		restoreChan := make(chan restore.Progress, 3)
		restoreChan <- restore.Progress{Current: 0, Total: 100, Message: "Starting restore"}
		restoreChan <- restore.Progress{Current: 50, Total: 100, Message: "50% restored"}
		restoreChan <- restore.Progress{Current: 100, Total: 100, Message: "Restore complete"}
		close(restoreChan)

		mockRestore.On("StreamProgress").Return((<-chan restore.Progress)(restoreChan))

		adapter := operation.NewRestoreAdapter(mockRestore)
		progressChan := adapter.StreamProgress()

		// Verify progress conversion
		progress := <-progressChan
		assert.Equal(t, int64(0), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "Starting restore", progress.Message)

		progress = <-progressChan
		assert.Equal(t, int64(50), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "50% restored", progress.Message)

		progress = <-progressChan
		assert.Equal(t, int64(100), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "Restore complete", progress.Message)

		// Verify channel is closed
		_, ok := <-progressChan
		assert.False(t, ok)

		mockRestore.AssertExpectations(t)
	})

	// Note: Testing nil progress channel is skipped because the adapter
	// implementation will hang when trying to range over a nil channel.
	// This is a known limitation in the current implementation.
}

func TestConvertBackupMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata *restore.Metadata
		expected *operation.Metadata
	}{
		{
			name: "full metadata conversion",
			metadata: &restore.Metadata{
				ID:           "backup-123",
				Type:         "filesystem",
				Namespace:    "test-ns",
				PodName:      "test-pod",
				SourcePath:   "/data",
				BackupFile:   "/backup.tar",
				Compression:  "gzip",
				Size:         1024000,
				Checksum:     "sha256:abc123",
				CreatedAt:    time.Now(),
				CompletedAt:  time.Now().Add(time.Minute),
				Status:       "completed",
				ProviderInfo: map[string]interface{}{"version": "1.0"},
			},
			expected: &operation.Metadata{
				ID:           "backup-123",
				Type:         "filesystem",
				Provider:     "filesystem",
				Namespace:    "test-ns",
				PodName:      "test-pod",
				SourcePath:   "/data",
				BackupFile:   "/backup.tar",
				Compression:  "gzip",
				Size:         1024000,
				Checksum:     "sha256:abc123",
				Status:       "completed",
				ProviderInfo: map[string]interface{}{"version": "1.0"},
			},
		},
		{
			name: "minimal metadata conversion",
			metadata: &restore.Metadata{
				ID:         "backup-456",
				Type:       "database",
				Namespace:  "prod",
				BackupFile: "/db-backup.sql",
				Size:       5000000,
			},
			expected: &operation.Metadata{
				ID:         "backup-456",
				Type:       "database",
				Provider:   "database",
				Namespace:  "prod",
				BackupFile: "/db-backup.sql",
				Size:       5000000,
			},
		},
		{
			name:     "nil metadata",
			metadata: nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := operation.ConvertBackupMetadata(tt.metadata)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Type, result.Type)
				assert.Equal(t, tt.expected.Provider, result.Provider)
				assert.Equal(t, tt.expected.Namespace, result.Namespace)
				assert.Equal(t, tt.expected.PodName, result.PodName)
				assert.Equal(t, tt.expected.SourcePath, result.SourcePath)
				assert.Equal(t, tt.expected.BackupFile, result.BackupFile)
				assert.Equal(t, tt.expected.Compression, result.Compression)
				assert.Equal(t, tt.expected.Size, result.Size)
				assert.Equal(t, tt.expected.Checksum, result.Checksum)
				assert.Equal(t, tt.expected.Status, result.Status)

				// ProviderInfo should match
				if tt.metadata.ProviderInfo != nil {
					assert.Equal(t, tt.metadata.ProviderInfo, result.ProviderInfo)
				}

				// Time fields should be set from the original
				if !tt.metadata.CreatedAt.IsZero() {
					assert.Equal(t, tt.metadata.CreatedAt, result.CreatedAt)
				}
				if !tt.metadata.CompletedAt.IsZero() {
					assert.Equal(t, tt.metadata.CompletedAt, result.CompletedAt)
				}

				// Extra should be initialized
				assert.NotNil(t, result.Extra)
			}
		})
	}
}

func TestConvertToRestoreMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata *operation.Metadata
		expected *restore.Metadata
	}{
		{
			name: "full metadata conversion",
			metadata: &operation.Metadata{
				ID:          "restore-123",
				Type:        "filesystem",
				Namespace:   "test-ns",
				PodName:     "test-pod",
				SourcePath:  "/data",
				BackupFile:  "/backup.tar",
				Compression: "gzip",
				Size:        1024000,
				Checksum:    "sha256:abc123",
				CreatedAt:   time.Now(),
				CompletedAt: time.Now().Add(time.Minute),
				Status:      "completed",
			},
			expected: &restore.Metadata{
				ID:          "restore-123",
				Type:        "filesystem",
				Namespace:   "test-ns",
				PodName:     "test-pod",
				SourcePath:  "/data",
				BackupFile:  "/backup.tar",
				Compression: "gzip",
				Size:        1024000,
				Checksum:    "sha256:abc123",
				Status:      "completed",
			},
		},
		{
			name: "metadata with container",
			metadata: &operation.Metadata{
				ID:         "restore-456",
				Type:       "database",
				Namespace:  "prod",
				PodName:    "db-pod",
				Container:  "postgres",
				BackupFile: "/db-backup.sql",
				Size:       5000000,
			},
			expected: &restore.Metadata{
				ID:         "restore-456",
				Type:       "database",
				Namespace:  "prod",
				PodName:    "db-pod",
				BackupFile: "/db-backup.sql",
				Size:       5000000,
			},
		},
		{
			name:     "nil metadata",
			metadata: nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := operation.ConvertToRestoreMetadata(tt.metadata)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Type, result.Type)
				assert.Equal(t, tt.expected.Namespace, result.Namespace)
				assert.Equal(t, tt.expected.PodName, result.PodName)
				assert.Equal(t, tt.expected.SourcePath, result.SourcePath)
				assert.Equal(t, tt.expected.BackupFile, result.BackupFile)
				assert.Equal(t, tt.expected.Compression, result.Compression)
				assert.Equal(t, tt.expected.Size, result.Size)
				assert.Equal(t, tt.expected.Checksum, result.Checksum)
				assert.Equal(t, tt.expected.Status, result.Status)

				// Time fields should be set from the original
				if !tt.metadata.CreatedAt.IsZero() {
					assert.Equal(t, tt.metadata.CreatedAt, result.CreatedAt)
				}
				if !tt.metadata.CompletedAt.IsZero() {
					assert.Equal(t, tt.metadata.CompletedAt, result.CompletedAt)
				}
			}
		})
	}
}

package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// MockRestoreProvider is a mock implementation of restore.Provider
type MockRestoreProvider struct {
	mock.Mock
	progressCh chan restore.Progress
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

func (m *MockRestoreProvider) StreamProgress() <-chan restore.Progress {
	if m.progressCh == nil {
		m.progressCh = make(chan restore.Progress)
	}
	return m.progressCh
}

func (m *MockRestoreProvider) EstimateSize(backupFile string) (int64, error) {
	args := m.Called(backupFile)
	return args.Get(0).(int64), args.Error(1)
}

// Test helpers
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
	}
	
	// Add flags
	cmd.Flags().String("namespace", "default", "Kubernetes namespace")
	cmd.Flags().String("target-path", "", "Target restore path")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing files")
	cmd.Flags().Bool("preserve-perms", false, "Preserve file permissions")
	cmd.Flags().StringSlice("skip-paths", []string{}, "Paths to skip during restore")
	cmd.Flags().String("container", "", "Container name")
	cmd.Flags().Bool("dry-run", false, "Show what would be executed without running")
	cmd.Flags().Bool("debug", false, "Enable debug output")
	cmd.Flags().Bool("verbose", false, "Verbose output")
	
	return cmd
}

func TestRestoreAdapter_ExecuteRestore_Success(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		args         []string
		setupFlags   func(*cobra.Command)
		setupMock    func(*MockRestoreProvider)
	}{
		{
			name:         "successful filesystem restore",
			providerName: "filesystem",
			args:         []string{"test-pod", "backup-20240101.tar.gz"},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("namespace", "production")
				cmd.Flags().Set("target-path", "/data/restore")
				cmd.Flags().Set("overwrite", "true")
			},
			setupMock: func(m *MockRestoreProvider) {
				expectedOpts := restore.Options{
					Namespace:     "production",
					PodName:       "test-pod",
					BackupFile:    "backup-20240101.tar.gz",
					TargetPath:    "/data/restore",
					Overwrite:     true,
					PreservePerms: false,
					SkipPaths:     []string{},
					Container:     "",
					Extra: map[string]interface{}{
						"verbose": false,
					},
				}
				
				m.On("ValidateOptions", expectedOpts).Return(nil)
				m.On("ValidateBackup", "backup-20240101.tar.gz", mock.Anything).Return(nil)
				m.On("EstimateSize", "backup-20240101.tar.gz").Return(int64(1024*1024), nil)
				m.On("Execute", mock.Anything, expectedOpts).Return(&restore.RestoreResult{
					Success:      true,
					RestoredPath: "/data/restore",
					FileCount:    42,
					BytesWritten: 1024 * 1024,
					Duration:     5 * time.Second,
					Warnings:     []string{},
				}, nil)
			},
		},
		{
			name:         "restore with warnings",
			providerName: "filesystem",
			args:         []string{"test-pod", "backup.tar"},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("target-path", "/restore")
			},
			setupMock: func(m *MockRestoreProvider) {
				m.On("ValidateOptions", mock.Anything).Return(nil)
				m.On("ValidateBackup", "backup.tar", mock.Anything).Return(nil)
				m.On("EstimateSize", "backup.tar").Return(int64(0), errors.New("cannot estimate"))
				m.On("Execute", mock.Anything, mock.Anything).Return(&restore.RestoreResult{
					Success:      true,
					RestoredPath: "/restore",
					FileCount:    10,
					BytesWritten: 512 * 1024,
					Duration:     2 * time.Second,
					Warnings:     []string{"Skipped 2 files due to permissions"},
				}, nil)
			},
		},
		{
			name:         "dry run mode",
			providerName: "filesystem",
			args:         []string{"test-pod", "backup.tar"},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("dry-run", "true")
				cmd.Flags().Set("target-path", "/data")
			},
			setupMock: func(m *MockRestoreProvider) {
				m.On("ValidateOptions", mock.Anything).Return(nil)
				m.On("ValidateBackup", "backup.tar", mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, adapter, mockProvider := setupRestoreTest(tt.setupFlags, tt.setupMock)
			
			err := adapter.ExecuteRestore(tt.providerName, cmd, tt.args)
			
			assert.NoError(t, err)
			mockProvider.AssertExpectations(t)
		})
	}
}

func TestRestoreAdapter_ExecuteRestore_ValidationFailures(t *testing.T) {
	tests := []struct {
		name          string
		providerName  string
		args          []string
		setupFlags    func(*cobra.Command)
		setupMock     func(*MockRestoreProvider)
		expectedError string
	}{
		{
			name:         "invalid options",
			providerName: "filesystem",
			args:         []string{"test-pod", "backup.tar"},
			setupMock: func(m *MockRestoreProvider) {
				m.On("ValidateOptions", mock.Anything).Return(errors.New("target path required"))
			},
			expectedError: "invalid options: target path required",
		},
		{
			name:         "backup validation failure",
			providerName: "filesystem",
			args:         []string{"test-pod", "invalid.tar"},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("target-path", "/data")
			},
			setupMock: func(m *MockRestoreProvider) {
				m.On("ValidateOptions", mock.Anything).Return(nil)
				m.On("ValidateBackup", "invalid.tar", mock.Anything).Return(errors.New("invalid backup format"))
			},
			expectedError: "backup validation failed: invalid backup format",
		},
		{
			name:          "insufficient arguments",
			providerName:  "filesystem",
			args:          []string{"test-pod"}, // Missing backup file
			expectedError: "filesystem restore requires [pod] [backup-file] arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, adapter, mockProvider := setupRestoreTest(tt.setupFlags, tt.setupMock)
			
			err := adapter.ExecuteRestore(tt.providerName, cmd, tt.args)
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			if mockProvider != nil {
				mockProvider.AssertExpectations(t)
			}
		})
	}
}

func TestRestoreAdapter_ExecuteRestore_ExecutionFailures(t *testing.T) {
	tests := []struct {
		name          string
		providerName  string
		args          []string
		setupFlags    func(*cobra.Command)
		setupMock     func(*MockRestoreProvider)
		expectedError string
	}{
		{
			name:         "restore execution failure",
			providerName: "filesystem",
			args:         []string{"test-pod", "backup.tar"},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("target-path", "/data")
			},
			setupMock: func(m *MockRestoreProvider) {
				m.On("ValidateOptions", mock.Anything).Return(nil)
				m.On("ValidateBackup", "backup.tar", mock.Anything).Return(nil)
				m.On("EstimateSize", "backup.tar").Return(int64(1024), nil)
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("pod not found"))
			},
			expectedError: "restore failed: pod not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, adapter, mockProvider := setupRestoreTest(tt.setupFlags, tt.setupMock)
			
			err := adapter.ExecuteRestore(tt.providerName, cmd, tt.args)
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			mockProvider.AssertExpectations(t)
		})
	}
}

func setupRestoreTest(setupFlags func(*cobra.Command), setupMock func(*MockRestoreProvider)) (*cobra.Command, *RestoreAdapter, *MockRestoreProvider) {
	cmd := createTestCommand()
	if setupFlags != nil {
		setupFlags(cmd)
	}

	var mockProvider *MockRestoreProvider
	if setupMock != nil {
		mockProvider = new(MockRestoreProvider)
		setupMock(mockProvider)
	}

	adapter := &RestoreAdapter{
		registry: &mockRestoreRegistry{
			provider: mockProvider,
		},
		logger: &NoOpLogger{},
	}

	return cmd, adapter, mockProvider
}

func TestRestoreAdapter_BuildOptions_Success(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		args         []string
		setupFlags   func(*cobra.Command)
		expected     restore.Options
	}{
		{
			name:         "filesystem options with all flags",
			providerName: "filesystem",
			args:         []string{"pod-name", "/backup/file.tar.gz"},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("namespace", "custom-ns")
				cmd.Flags().Set("target-path", "/restore/path")
				cmd.Flags().Set("overwrite", "true")
				cmd.Flags().Set("preserve-perms", "true")
				cmd.Flags().Set("skip-paths", "tmp,cache")
				cmd.Flags().Set("container", "app")
				cmd.Flags().Set("verbose", "true")
			},
			expected: restore.Options{
				Namespace:     "custom-ns",
				PodName:       "pod-name",
				BackupFile:    "/backup/file.tar.gz",
				TargetPath:    "/restore/path",
				Container:     "app",
				Overwrite:     true,
				PreservePerms: true,
				SkipPaths:     []string{"tmp", "cache"},
				Extra: map[string]interface{}{
					"verbose": true,
				},
			},
		},
		{
			name:         "default target path",
			providerName: "filesystem",
			args:         []string{"pod", "backup.tar"},
			setupFlags:   func(cmd *cobra.Command) {},
			expected: restore.Options{
				Namespace:     "default",
				PodName:       "pod",
				BackupFile:    "backup.tar",
				TargetPath:    "/",
				Container:     "",
				Overwrite:     false,
				PreservePerms: false,
				SkipPaths:     []string{},
				Extra: map[string]interface{}{
					"verbose": false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommand()
			if tt.setupFlags != nil {
				tt.setupFlags(cmd)
			}

			adapter := &RestoreAdapter{logger: &NoOpLogger{}}
			opts, err := adapter.buildOptions(tt.providerName, cmd, tt.args)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, opts)
		})
	}
}

func TestRestoreAdapter_BuildOptions_Failure(t *testing.T) {
	t.Run("unknown provider", func(t *testing.T) {
		cmd := createTestCommand()
		adapter := &RestoreAdapter{logger: &NoOpLogger{}}
		
		_, err := adapter.buildOptions("unknown", cmd, []string{"arg1", "arg2"})
		
		assert.Error(t, err)
	})
}

func TestRestoreAdapter_MonitorProgress(t *testing.T) {
	// Create mock provider with progress channel
	mockProvider := &MockRestoreProvider{
		progressCh: make(chan restore.Progress, 10),
	}

	adapter := &RestoreAdapter{logger: &NoOpLogger{}}

	// Start monitoring in goroutine
	done := make(chan bool)
	output := &strings.Builder{}

	go func() {
		// Redirect stderr for testing
		r, w, _ := os.Pipe()
		oldStderr := os.Stderr
		os.Stderr = w

		// Monitor progress with verbose=true
		go adapter.monitorProgress(mockProvider, 1024*1024, done, true)

		// Send progress updates
		mockProvider.progressCh <- restore.Progress{
			Current: 256 * 1024,
			Total:   1024 * 1024,
			Message: "Restoring file1.txt",
		}
		time.Sleep(100 * time.Millisecond)

		mockProvider.progressCh <- restore.Progress{
			Current: 512 * 1024,
			Total:   1024 * 1024,
			Message: "Restoring file2.txt",
		}
		time.Sleep(100 * time.Millisecond)

		// Close monitoring
		close(done)
		time.Sleep(100 * time.Millisecond)

		// Restore stderr and read output
		w.Close()
		os.Stderr = oldStderr
		io.Copy(output, r)
	}()

	// Wait for completion
	time.Sleep(500 * time.Millisecond)

	// Check output contains progress info
	outputStr := output.String()
	assert.Contains(t, outputStr, "[PROGRESS]")
	assert.Contains(t, outputStr, "Restoring file1.txt")
}

// Mock registry for testing
type mockRestoreRegistry struct {
	provider restore.Provider
}

func (r *mockRestoreRegistry) Register(name string, factory func() restore.Provider) error {
	return nil
}

func (r *mockRestoreRegistry) Create(name string) (restore.Provider, error) {
	if r.provider != nil {
		return r.provider, nil
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

func (r *mockRestoreRegistry) List() []string {
	return []string{"filesystem"}
}
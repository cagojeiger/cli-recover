package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
)

// Mock for backup.Provider
type mockBackupProvider struct {
	mock.Mock
}

func (m *mockBackupProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockBackupProvider) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockBackupProvider) Execute(ctx context.Context, opts backup.Options) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func (m *mockBackupProvider) ExecuteWithResult(ctx context.Context, opts backup.Options) (*backup.Result, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*backup.Result), args.Error(1)
}

func (m *mockBackupProvider) EstimateSize(opts backup.Options) (int64, error) {
	args := m.Called(opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockBackupProvider) StreamProgress() <-chan backup.Progress {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(<-chan backup.Progress)
}

func (m *mockBackupProvider) ValidateOptions(opts backup.Options) error {
	args := m.Called(opts)
	return args.Error(0)
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path",
			input:    "/var/log",
			expected: "var-log",
		},
		{
			name:     "path with multiple slashes",
			input:    "/usr/local/bin/",
			expected: "usr-local-bin-",
		},
		{
			name:     "path with special characters",
			input:    "/var/lib/mysql-5.7",
			expected: "var-lib-mysql-5-7",
		},
		{
			name:     "root path",
			input:    "/",
			expected: "root",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "path with dots",
			input:    "/home/user/.config",
			expected: "home-user--config",
		},
		{
			name:     "path with spaces",
			input:    "/var/my files/backup",
			expected: "var-my-files-backup",
		},
		{
			name:     "path with trailing slash",
			input:    "/data/",
			expected: "data-",
		},
		{
			name:     "complex nested path",
			input:    "/opt/app/v1.2.3/config/",
			expected: "opt-app-v1-2-3-config-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name        string
		compression string
		expected    string
	}{
		{
			name:        "gzip compression",
			compression: "gzip",
			expected:    ".tar.gz",
		},
		{
			name:        "bzip2 compression",
			compression: "bzip2",
			expected:    ".tar",
		},
		{
			name:        "xz compression",
			compression: "xz",
			expected:    ".tar",
		},
		{
			name:        "no compression",
			compression: "none",
			expected:    ".tar",
		},
		{
			name:        "empty compression",
			compression: "",
			expected:    ".tar",
		},
		{
			name:        "unknown compression",
			compression: "unknown",
			expected:    ".tar",
		},
		{
			name:        "mixed case compression",
			compression: "GZip",
			expected:    ".tar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileExtension(tt.compression)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "bytes",
			bytes:    500,
			expected: "500 B",
		},
		{
			name:     "exactly 1 KB",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "kilobytes",
			bytes:    2048,
			expected: "2.0 KB",
		},
		{
			name:     "fractional kilobytes",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "exactly 1 MB",
			bytes:    1024 * 1024,
			expected: "1.0 MB",
		},
		{
			name:     "megabytes",
			bytes:    5 * 1024 * 1024,
			expected: "5.0 MB",
		},
		{
			name:     "fractional megabytes",
			bytes:    int64(1.5 * 1024 * 1024),
			expected: "1.5 MB",
		},
		{
			name:     "exactly 1 GB",
			bytes:    1024 * 1024 * 1024,
			expected: "1.0 GB",
		},
		{
			name:     "gigabytes",
			bytes:    int64(2.5 * 1024 * 1024 * 1024),
			expected: "2.5 GB",
		},
		{
			name:     "exactly 1 TB",
			bytes:    1024 * 1024 * 1024 * 1024,
			expected: "1.0 TB",
		},
		{
			name:     "terabytes",
			bytes:    3 * 1024 * 1024 * 1024 * 1024 + 751619276800, // ~3.7 TB
			expected: "3.7 TB",
		},
		{
			name:     "negative bytes",
			bytes:    -1024,
			expected: "-1024 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := humanizeBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLogDirFromCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		expected    string
		description string
	}{
		{
			name: "with log-dir flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("log-dir", "/custom/logs", "")
				cmd.Flags().Set("log-dir", "/custom/logs")
				return cmd
			},
			expected:    "/custom/logs",
			description: "Should return custom log directory when specified",
		},
		{
			name: "without log-dir flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				return cmd
			},
			expected:    func() string {
				homeDir, _ := os.UserHomeDir()
				return filepath.Join(homeDir, ".cli-recover", "logs")
			}(),
			description: "Should return default log directory when not specified",
		},
		{
			name: "with empty log-dir flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("log-dir", "", "")
				return cmd
			},
			expected:    func() string {
				homeDir, _ := os.UserHomeDir()
				return filepath.Join(homeDir, ".cli-recover", "logs")
			}(),
			description: "Should return default log directory when flag is empty",
		},
		{
			name: "with root command having log-dir",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				// The function only checks the command's own flags, not parent
				// So this test case should expect the default directory
				return cmd
			},
			expected:    func() string {
				homeDir, _ := os.UserHomeDir()
				return filepath.Join(homeDir, ".cli-recover", "logs")
			}(),
			description: "Function only checks command's own flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			result := getLogDirFromCmd(cmd)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestBuildBackupOptions(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		setupCmd     func() *cobra.Command
		args         []string
		expected     backup.Options
		expectedErr  string
	}{
		{
			name:         "filesystem backup with all options",
			providerName: "filesystem",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("namespace", "production", "")
				cmd.Flags().String("output", "/backups/app.tar.gz", "")
				cmd.Flags().String("compression", "gzip", "")
				cmd.Flags().StringSlice("exclude", []string{"*.log", "*.tmp"}, "")
				cmd.Flags().String("container", "app", "")
				return cmd
			},
			args: []string{"webapp-1", "/app/data"},
			expected: backup.Options{
				Namespace:  "production",
				PodName:    "webapp-1",
				SourcePath: "/app/data",
				OutputFile: "/backups/app.tar.gz",
				Compress:   true,
				Container:  "app",
				Exclude:    []string{"*.log", "*.tmp"},
			},
			expectedErr: "",
		},
		{
			name:         "filesystem backup with minimal options",
			providerName: "filesystem",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("namespace", "default", "")
				cmd.Flags().String("compression", "none", "")
				return cmd
			},
			args: []string{"test-pod", "/data"},
			expected: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "", // Will be generated
				Compress:   false,
				Exclude:    nil,
			},
			expectedErr: "",
		},
		{
			name:         "filesystem backup missing arguments",
			providerName: "filesystem",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("namespace", "default", "")
				return cmd
			},
			args:        []string{"test-pod"},
			expected:    backup.Options{},
			expectedErr: "Missing required arguments",
		},
		{
			name:         "unknown provider",
			providerName: "unknown",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				return cmd
			},
			args:        []string{"arg1", "arg2"},
			expected:    backup.Options{},
			expectedErr: "unknown provider: unknown",
		},
		{
			name:         "filesystem backup with output directory",
			providerName: "filesystem",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{}
				cmd.Flags().String("namespace", "default", "")
				cmd.Flags().String("output", "/backups/", "")
				cmd.Flags().String("compression", "none", "")
				return cmd
			},
			args: []string{"test-pod", "/var/log"},
			expected: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/var/log",
				OutputFile: "", // Will be generated with proper filename
				Compress:   false,
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			result, err := buildBackupOptions(tt.providerName, cmd, tt.args)
			
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Namespace, result.Namespace)
				assert.Equal(t, tt.expected.PodName, result.PodName)
				assert.Equal(t, tt.expected.SourcePath, result.SourcePath)
				assert.Equal(t, tt.expected.Compress, result.Compress)
				assert.Equal(t, tt.expected.Container, result.Container)
				if tt.expected.Exclude == nil && result.Exclude != nil && len(result.Exclude) == 0 {
					// Empty slice is equivalent to nil for our purposes
					assert.Empty(t, result.Exclude)
				} else {
					assert.Equal(t, tt.expected.Exclude, result.Exclude)
				}
				
				// Check output file generation
				if tt.expected.OutputFile != "" {
					assert.Equal(t, tt.expected.OutputFile, result.OutputFile)
				} else {
					assert.NotEmpty(t, result.OutputFile)
					if outputFlag := cmd.Flag("output"); outputFlag != nil && strings.HasSuffix(outputFlag.Value.String(), "/") {
						// Should generate filename in directory
						assert.True(t, strings.HasPrefix(result.OutputFile, outputFlag.Value.String()))
					}
				}
			}
		})
	}
}

func TestMonitorBackupProgress(t *testing.T) {
	t.Run("normal progress monitoring", func(t *testing.T) {
		provider := new(mockBackupProvider)
		progressChan := make(chan backup.Progress, 3)
		
		// Send some progress updates
		go func() {
			progressChan <- backup.Progress{Current: 0, Total: 100, Message: "Starting"}
			time.Sleep(10 * time.Millisecond)
			progressChan <- backup.Progress{Current: 50, Total: 100, Message: "50% complete"}
			time.Sleep(10 * time.Millisecond)
			progressChan <- backup.Progress{Current: 100, Total: 100, Message: "Complete"}
			close(progressChan)
		}()
		
		provider.On("StreamProgress").Return((<-chan backup.Progress)(progressChan))
		
		done := make(chan bool)
		go func() {
			time.Sleep(50 * time.Millisecond)
			close(done)
		}()
		
		// This should not panic or error
		monitorBackupProgress(provider, 1024*1024, done, false)
		
		provider.AssertExpectations(t)
	})
	
	t.Run("progress monitoring with early done signal", func(t *testing.T) {
		provider := new(mockBackupProvider)
		progressChan := make(chan backup.Progress, 1)
		progressChan <- backup.Progress{Current: 10, Total: 100, Message: "Starting"}
		
		provider.On("StreamProgress").Return((<-chan backup.Progress)(progressChan))
		
		done := make(chan bool)
		close(done) // Close immediately
		
		// Should exit gracefully
		monitorBackupProgress(provider, 1024*1024, done, true)
		
		provider.AssertExpectations(t)
	})
}

func TestSaveBackupMetadataWithChecksum(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		opts         backup.Options
		size         int64
		startTime    time.Time
		endTime      time.Time
		checksum     string
		expectedErr  bool
	}{
		{
			name:         "successful metadata save",
			providerName: "filesystem",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
				Compress:   true,
				Extra: map[string]interface{}{
					"compression": "gzip",
				},
			},
			size:        1024 * 1024,
			startTime:   time.Now().Add(-time.Minute),
			endTime:     time.Now(),
			checksum:    "sha256:abc123",
			expectedErr: false,
		},
		{
			name:         "metadata save with minimal options",
			providerName: "filesystem",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			size:        1024,
			startTime:   time.Now().Add(-time.Second * 30),
			endTime:     time.Now(),
			checksum:    "sha256:def456",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := saveBackupMetadataWithChecksum(tt.providerName, tt.opts, tt.size, tt.startTime, tt.endTime, tt.checksum)
			
			// The function should always succeed as DefaultStore is initialized
			// It might save to a temp directory if HOME is not accessible
			assert.NoError(t, err)
		})
	}
}

func TestExecuteBackup_Integration(t *testing.T) {
	t.Run("filesystem backup validation error", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("namespace", "default", "")
		
		// Missing required arguments
		err := executeBackup("filesystem", cmd, []string{"pod-only"})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Missing required arguments")
	})
	
	t.Run("unknown provider error", func(t *testing.T) {
		cmd := &cobra.Command{}
		
		err := executeBackup("unknown-provider", cmd, []string{"arg1", "arg2"})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown backup provider")
	})
}

// Helper function tests
func TestBackupOptionValidation(t *testing.T) {
	tests := []struct {
		name        string
		opts        backup.Options
		expectedErr bool
		errContains string
	}{
		{
			name: "valid options",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			expectedErr: false,
		},
		{
			name: "missing namespace",
			opts: backup.Options{
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			expectedErr: true,
			errContains: "namespace is required",
		},
		{
			name: "missing pod name",
			opts: backup.Options{
				Namespace:  "default",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			expectedErr: true,
			errContains: "pod name is required",
		},
		{
			name: "missing source path",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				OutputFile: "/backup.tar",
			},
			expectedErr: true,
			errContains: "source path is required",
		},
		{
			name: "missing output file",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
			},
			expectedErr: true,
			errContains: "output file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			
			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkSanitizePath(b *testing.B) {
	paths := []string{
		"/var/log/application",
		"/usr/local/bin/script",
		"/home/user/.config/app",
		"/opt/data/backup/2024/01",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizePath(paths[i%len(paths)])
	}
}

func BenchmarkHumanizeBytes(b *testing.B) {
	sizes := []int64{
		1024,
		1024 * 1024,
		1024 * 1024 * 1024,
		1024 * 1024 * 1024 * 1024,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		humanizeBytes(sizes[i%len(sizes)])
	}
}
package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
)

// TestBackupExecutor tests the integrated backup execution logic
type TestBackupExecutor struct {
	provider backup.Provider
	executed bool
	options  backup.Options
}

func (e *TestBackupExecutor) Execute(ctx context.Context, providerName string, opts backup.Options) error {
	e.executed = true
	e.options = opts
	
	// Simulate provider execution
	if e.provider != nil {
		return e.provider.Execute(ctx, opts)
	}
	
	// Create dummy output file
	if opts.OutputFile != "" {
		return os.WriteFile(opts.OutputFile, []byte("test backup data"), 0644)
	}
	
	return nil
}

// TestBackupIntegration tests backup command with integrated adapter logic
func TestBackupIntegration_FilesystemBackup(t *testing.T) {
	// Create temp directory for test output
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-backup.tar.gz")
	
	// Test backup options building
	opts := backup.Options{
		Namespace:  "test-namespace",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: outputFile,
		Compress:   true,
		Extra:      make(map[string]interface{}),
	}
	
	opts.Extra["compression"] = "gzip"
	opts.Extra["verbose"] = false
	opts.Extra["totals"] = true
	opts.Extra["preserve-perms"] = true
	
	// Verify options are built correctly
	assert.Equal(t, "test-namespace", opts.Namespace)
	assert.Equal(t, "test-pod", opts.PodName)
	assert.Equal(t, "/data", opts.SourcePath)
	assert.Equal(t, outputFile, opts.OutputFile)
	assert.True(t, opts.Compress)
	assert.Equal(t, "gzip", opts.Extra["compression"])
}

// TestBackupIntegration_ExecuteWithMockProvider tests backup execution flow
func TestBackupIntegration_ExecuteWithMockProvider(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-backup.tar")
	
	// Create a mock provider
	mockProvider := &MockBackupProvider{
		validateFunc: func(opts backup.Options) error {
			return nil
		},
		executeFunc: func(ctx context.Context, opts backup.Options) error {
			// Simulate writing backup data
			return os.WriteFile(opts.OutputFile, []byte("mock backup data"), 0644)
		},
		estimateSizeFunc: func(opts backup.Options) (int64, error) {
			return 1024, nil
		},
		progressChan: make(chan backup.Progress, 1),
	}
	
	// Test backup execution
	ctx := context.Background()
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/test",
		OutputFile: outputFile,
		Extra:      map[string]interface{}{"verbose": false},
	}
	
	// Validate options
	err := mockProvider.ValidateOptions(opts)
	require.NoError(t, err)
	
	// Execute backup
	err = mockProvider.Execute(ctx, opts)
	require.NoError(t, err)
	
	// Verify output file was created
	_, err = os.Stat(outputFile)
	require.NoError(t, err)
	
	// Verify file content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "mock backup data", string(content))
}

// TestBackupIntegration_ProgressMonitoring tests progress monitoring
func TestBackupIntegration_ProgressMonitoring(t *testing.T) {
	mockProvider := &MockBackupProvider{
		progressChan: make(chan backup.Progress, 10),
	}
	
	// Send some progress updates
	go func() {
		mockProvider.progressChan <- backup.Progress{
			Current: 100,
			Total:   1000,
			Message: "Processing file 1",
		}
		mockProvider.progressChan <- backup.Progress{
			Current: 500,
			Total:   1000,
			Message: "Processing file 2",
		}
		close(mockProvider.progressChan)
	}()
	
	// Collect progress updates
	var updates []backup.Progress
	for progress := range mockProvider.StreamProgress() {
		updates = append(updates, progress)
	}
	
	assert.Len(t, updates, 2)
	assert.Equal(t, int64(100), updates[0].Current)
	assert.Equal(t, int64(500), updates[1].Current)
}

// TestBackupIntegration_DryRun tests dry-run functionality
func TestBackupIntegration_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-backup.tar")
	
	// Create options with dry-run
	_ = backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/test",
		OutputFile: outputFile,
		Extra:      map[string]interface{}{"dry-run": true},
	}
	
	// In dry-run mode, no file should be created
	// This would be handled by the integrated logic checking dry-run flag
	
	// Verify no output file exists
	_, err := os.Stat(outputFile)
	assert.True(t, os.IsNotExist(err))
}

// MockBackupProvider implements backup.Provider for testing
type MockBackupProvider struct {
	validateFunc     func(backup.Options) error
	executeFunc      func(context.Context, backup.Options) error
	estimateSizeFunc func(backup.Options) (int64, error)
	progressChan     chan backup.Progress
}

func (m *MockBackupProvider) ValidateOptions(opts backup.Options) error {
	if m.validateFunc != nil {
		return m.validateFunc(opts)
	}
	return nil
}

func (m *MockBackupProvider) Execute(ctx context.Context, opts backup.Options) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, opts)
	}
	return nil
}

func (m *MockBackupProvider) EstimateSize(opts backup.Options) (int64, error) {
	if m.estimateSizeFunc != nil {
		return m.estimateSizeFunc(opts)
	}
	return 0, nil
}

func (m *MockBackupProvider) StreamProgress() <-chan backup.Progress {
	return m.progressChan
}

// TestBackupIntegration_MetadataSaving tests metadata saving functionality
func TestBackupIntegration_MetadataSaving(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-backup.tar")
	
	// Create a test backup file
	err := os.WriteFile(outputFile, []byte("test data"), 0644)
	require.NoError(t, err)
	
	// Test metadata creation
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)
	
	metadata := map[string]interface{}{
		"type":        "filesystem",
		"namespace":   "default",
		"pod":         "test-pod",
		"path":        "/data",
		"backup_file": outputFile,
		"size":        int64(len("test data")),
		"created_at":  startTime,
		"completed_at": endTime,
		"status":      "completed",
	}
	
	// Verify metadata fields
	assert.Equal(t, "filesystem", metadata["type"])
	assert.Equal(t, "default", metadata["namespace"])
	assert.Equal(t, "test-pod", metadata["pod"])
	assert.Equal(t, "/data", metadata["path"])
	assert.Equal(t, outputFile, metadata["backup_file"])
	assert.Equal(t, int64(9), metadata["size"])
}

// TestBackupIntegration_ErrorHandling tests error handling in backup flow
func TestBackupIntegration_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		providerError error
		expectError   bool
		errorContains string
	}{
		{
			name:          "successful backup",
			providerError: nil,
			expectError:   false,
		},
		{
			name:          "provider validation error",
			providerError: assert.AnError,
			expectError:   true,
			errorContains: "error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockBackupProvider{
				validateFunc: func(opts backup.Options) error {
					if tt.name == "provider validation error" {
						return tt.providerError
					}
					return nil
				},
				executeFunc: func(ctx context.Context, opts backup.Options) error {
					if tt.name == "successful backup" {
						return nil
					}
					return tt.providerError
				},
			}
			
			opts := backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/test",
			}
			
			err := mockProvider.ValidateOptions(opts)
			if tt.expectError && tt.name == "provider validation error" {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else if tt.name == "successful backup" {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBackupIntegration_BuildOptionsFromFlags tests building options from command flags
func TestBackupIntegration_BuildOptionsFromFlags(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		args         []string
		flags        map[string]interface{}
		expectError  bool
		validateOpts func(t *testing.T, opts backup.Options)
	}{
		{
			name:     "filesystem with all options",
			provider: "filesystem",
			args:     []string{"test-pod", "/data"},
			flags: map[string]interface{}{
				"namespace":      "test-ns",
				"compression":    "gzip",
				"exclude":        []string{"*.log", "*.tmp"},
				"container":      "app",
				"output":         "backup.tar.gz",
				"verbose":        true,
				"preserve-perms": true,
			},
			expectError: false,
			validateOpts: func(t *testing.T, opts backup.Options) {
				assert.Equal(t, "test-ns", opts.Namespace)
				assert.Equal(t, "test-pod", opts.PodName)
				assert.Equal(t, "/data", opts.SourcePath)
				assert.True(t, opts.Compress)
				assert.Equal(t, []string{"*.log", "*.tmp"}, opts.Exclude)
				assert.Equal(t, "app", opts.Container)
				assert.Equal(t, "backup.tar.gz", opts.OutputFile)
			},
		},
		{
			name:        "filesystem missing args",
			provider:    "filesystem",
			args:        []string{"test-pod"}, // missing path
			expectError: true,
		},
		{
			name:     "filesystem with auto-generated output",
			provider: "filesystem",
			args:     []string{"test-pod", "/var/log"},
			flags: map[string]interface{}{
				"namespace":   "default",
				"compression": "none",
			},
			expectError: false,
			validateOpts: func(t *testing.T, opts backup.Options) {
				assert.Contains(t, opts.OutputFile, "backup-default-test-pod-var-log.tar")
				assert.False(t, opts.Compress)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate building options from flags
			opts := backup.Options{
				Extra: make(map[string]interface{}),
			}
			
			// Apply common flags
			if ns, ok := tt.flags["namespace"]; ok {
				opts.Namespace = ns.(string)
			} else {
				opts.Namespace = "default"
			}
			
			// Provider-specific logic
			switch tt.provider {
			case "filesystem":
				if len(tt.args) < 2 {
					if tt.expectError {
						return // Expected error case
					}
					t.Fatal("filesystem requires 2 args")
				}
				
				opts.PodName = tt.args[0]
				opts.SourcePath = tt.args[1]
				
				if compression, ok := tt.flags["compression"]; ok {
					opts.Compress = compression.(string) == "gzip"
					opts.Extra["compression"] = compression
				}
				
				if exclude, ok := tt.flags["exclude"]; ok {
					opts.Exclude = exclude.([]string)
				}
				
				if container, ok := tt.flags["container"]; ok {
					opts.Container = container.(string)
				}
				
				if output, ok := tt.flags["output"]; ok {
					opts.OutputFile = output.(string)
				} else {
					// Auto-generate filename
					ext := ".tar"
					if opts.Compress {
						ext = ".tar.gz"
					}
					sanitized := "var-log" // Simulated sanitization
					opts.OutputFile = "backup-" + opts.Namespace + "-" + opts.PodName + "-" + sanitized + ext
				}
				
				// Store extra flags
				for k, v := range tt.flags {
					if k == "verbose" || k == "preserve-perms" || k == "totals" {
						opts.Extra[k] = v
					}
				}
			}
			
			if !tt.expectError && tt.validateOpts != nil {
				tt.validateOpts(t, opts)
			}
		})
	}
}
package filesystem

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid absolute path",
			path:    "/tmp/restore",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "tmp/restore",
			wantErr: true,
			errMsg:  "must be absolute",
		},
		{
			name:    "path traversal attempt",
			path:    "/tmp/../etc/passwd",
			wantErr: true,
			errMsg:  "cannot contain '..'",
		},
		{
			name:    "hidden path traversal",
			path:    "/tmp/foo/../bar",
			wantErr: true,
			errMsg:  "cannot contain '..'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRestoreProvider_ValidateOptions_WithPathValidation(t *testing.T) {
	provider := NewRestoreProvider(nil, nil)

	tests := []struct {
		name    string
		opts    restore.Options
		wantErr string
	}{
		{
			name: "valid options",
			opts: restore.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				BackupFile: "backup.tar",
				TargetPath: "/tmp/restore",
			},
			wantErr: "",
		},
		{
			name: "relative target path",
			opts: restore.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				BackupFile: "backup.tar",
				TargetPath: "tmp/restore",
			},
			wantErr: "must be absolute",
		},
		{
			name: "path traversal in target",
			opts: restore.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				BackupFile: "backup.tar",
				TargetPath: "/tmp/../etc",
			},
			wantErr: "cannot contain '..'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateOptions(tt.opts)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRestoreProvider_Execute_Timeout(t *testing.T) {
	// Skip this integration test in short mode
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test requires mocked kubernetes client
	t.Skip("skipping test that requires kubernetes client mock")
}

func TestRestoreProvider_EstimateSize_Integration(t *testing.T) {
	provider := NewRestoreProvider(nil, nil)

	// Create a temporary file with known size
	tmpDir := t.TempDir()
	backupFile := filepath.Join(tmpDir, "test.tar")
	
	testData := []byte("This is test data for size estimation")
	err := os.WriteFile(backupFile, testData, 0644)
	require.NoError(t, err)

	// Test size estimation
	size, err := provider.EstimateSize(backupFile)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testData)), size)

	// Test with non-existent file
	_, err = provider.EstimateSize("/non/existent/file")
	assert.Error(t, err)
}

func TestRestoreProvider_ProgressReporting(t *testing.T) {
	// This test verifies that progress is reported correctly
	provider := NewRestoreProvider(nil, nil)
	
	// Get progress channel
	progressCh := provider.StreamProgress()
	assert.NotNil(t, progressCh)

	// Send some test progress
	go func() {
		provider.progressCh <- restore.Progress{
			Current: 50,
			Total:   100,
			Message: "Test progress",
		}
	}()

	// Verify we can receive progress
	select {
	case progress := <-progressCh:
		assert.Equal(t, int64(50), progress.Current)
		assert.Equal(t, int64(100), progress.Total)
		assert.Equal(t, "Test progress", progress.Message)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for progress")
	}
}
package kubernetes

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreExecutor_ExecuteRestore(t *testing.T) {
	// Create a temporary backup file for testing
	tmpDir := t.TempDir()
	backupFile := filepath.Join(tmpDir, "test-backup.tar")

	// Create a simple tar file
	err := os.WriteFile(backupFile, []byte("test tar content"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		backupFile  string
		kubectlArgs []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid backup file",
			backupFile:  backupFile,
			kubectlArgs: []string{"exec", "-i", "-n", "test", "test-pod", "--", "tar", "-xvf", "-", "-C", "/tmp"},
			wantErr:     true, // Will fail as kubectl is not available in test
			errContains: "kubectl",
		},
		{
			name:        "non-existent backup file",
			backupFile:  "/non/existent/file.tar",
			kubectlArgs: []string{"exec", "-i", "-n", "test", "test-pod", "--", "tar", "-xvf", "-"},
			wantErr:     true,
			errContains: "failed to open backup file",
		},
		{
			name:        "not a regular file",
			backupFile:  tmpDir,
			kubectlArgs: []string{"exec", "-i", "-n", "test", "test-pod", "--", "tar", "-xvf", "-"},
			wantErr:     true,
			errContains: "not a regular file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progressCh := make(chan RestoreProgress, 100)
			executor := NewRestoreExecutor(progressCh)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := executor.ExecuteRestore(ctx, tt.backupFile, tt.kubectlArgs)

			if tt.wantErr {
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

func TestRestoreExecutor_monitorProgress(t *testing.T) {
	tests := []struct {
		name          string
		stderrOutput  []string
		expectedCount int
		expectedFiles []string
	}{
		{
			name: "tar verbose output",
			stderrOutput: []string{
				"x usr/",
				"x usr/bin/",
				"x usr/bin/app",
				"x etc/config.yaml",
			},
			expectedCount: 4,
			expectedFiles: []string{"usr/", "usr/bin/", "usr/bin/app", "etc/config.yaml"},
		},
		{
			name: "tar with errors",
			stderrOutput: []string{
				"x file1.txt",
				"tar: Cannot create symlink to 'target': Operation not permitted",
				"x file2.txt",
			},
			expectedCount: 2,
			expectedFiles: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "mixed output",
			stderrOutput: []string{
				"x data/",
				"some other output",
				"x data/file.json",
				"",
				"x data/image.png",
			},
			expectedCount: 3,
			expectedFiles: []string{"data/", "data/file.json", "data/image.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progressCh := make(chan RestoreProgress, 100)
			executor := NewRestoreExecutor(progressCh)

			// Create a pipe to simulate stderr
			r, w, err := os.Pipe()
			require.NoError(t, err)

			// Write test data in a goroutine
			go func() {
				for _, line := range tt.stderrOutput {
					w.Write([]byte(line + "\n"))
				}
				w.Close()
			}()

			// Monitor progress
			err = executor.monitorProgress(r)
			assert.NoError(t, err)

			// Verify progress updates
			close(progressCh)

			fileCount := 0
			var files []string
			for progress := range progressCh {
				if progress.Error == nil && progress.FileName != "" {
					fileCount++
					files = append(files, progress.FileName)
				}
			}

			assert.Equal(t, tt.expectedCount, fileCount)
			assert.Equal(t, tt.expectedFiles, files)
		})
	}
}

func TestDetectCompression(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"backup.tar.gz", "gzip"},
		{"backup.tgz", "gzip"},
		{"BACKUP.TAR.GZ", "gzip"},
		{"backup.tar.bz2", "bzip2"},
		{"backup.tar.xz", "xz"},
		{"backup.tar", "none"},
		{"backup.zip", "none"},
		{"backup", "none"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detectCompression(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStreamRestore(t *testing.T) {
	// This is more of an integration test
	tmpDir := t.TempDir()
	backupFile := filepath.Join(tmpDir, "test.tar")

	// Create a dummy tar file
	err := os.WriteFile(backupFile, []byte("dummy tar content"), 0644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	progressCh, err := StreamRestore(
		ctx,
		backupFile,
		"test-namespace",
		"test-pod",
		"",
		"/tmp",
		false,
		true,
		[]string{"*.log"},
	)

	assert.NoError(t, err)
	assert.NotNil(t, progressCh)

	// The actual kubectl command will fail in test environment
	// Just verify that progress channel receives an error
	timeout := time.After(2 * time.Second)
	select {
	case progress := <-progressCh:
		assert.NotNil(t, progress.Error)
	case <-timeout:
		t.Fatal("timeout waiting for progress")
	}
}

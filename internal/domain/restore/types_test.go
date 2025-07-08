package restore_test

import (
	"testing"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	progress := restore.Progress{
		Current: 1024,
		Total:   2048,
		Message: "Processing files...",
	}

	assert.Equal(t, int64(1024), progress.Current)
	assert.Equal(t, int64(2048), progress.Total)
	assert.Equal(t, "Processing files...", progress.Message)
}

func TestOptions(t *testing.T) {
	options := restore.Options{
		Namespace:     "default",
		PodName:       "test-pod",
		BackupFile:    "/path/to/backup.tar.gz",
		TargetPath:    "/restore/path",
		Container:     "main",
		Extra:         map[string]interface{}{"key": "value"},
		Overwrite:     true,
		PreservePerms: false,
		SkipPaths:     []string{"/tmp", "/cache"},
	}

	assert.Equal(t, "default", options.Namespace)
	assert.Equal(t, "test-pod", options.PodName)
	assert.Equal(t, "/path/to/backup.tar.gz", options.BackupFile)
	assert.Equal(t, "/restore/path", options.TargetPath)
	assert.Equal(t, "main", options.Container)
	assert.Equal(t, "value", options.Extra["key"])
	assert.True(t, options.Overwrite)
	assert.False(t, options.PreservePerms)
	assert.Contains(t, options.SkipPaths, "/tmp")
	assert.Contains(t, options.SkipPaths, "/cache")
}

func TestRestoreError_Error(t *testing.T) {
	err := restore.RestoreError{
		Code:    "RESTORE_FAILED",
		Message: "Failed to restore backup",
		Details: map[string]interface{}{"reason": "file not found"},
	}

	assert.Equal(t, "Failed to restore backup", err.Error())
	assert.Equal(t, "RESTORE_FAILED", err.Code)
	assert.Equal(t, "file not found", err.Details["reason"])
}

func TestNewRestoreError(t *testing.T) {
	err := restore.NewRestoreError("INVALID_BACKUP", "Backup file is corrupted")

	assert.Equal(t, "INVALID_BACKUP", err.Code)
	assert.Equal(t, "Backup file is corrupted", err.Message)
	assert.NotNil(t, err.Details)
	assert.Empty(t, err.Details)
}

func TestRestoreError_WithDetail(t *testing.T) {
	err := restore.NewRestoreError("PERMISSION_ERROR", "Permission denied")

	// Test method chaining
	err = err.WithDetail("file", "/etc/hosts").
		WithDetail("user", "root").
		WithDetail("retryable", true)

	assert.Equal(t, "PERMISSION_ERROR", err.Code)
	assert.Equal(t, "Permission denied", err.Message)
	assert.Equal(t, "/etc/hosts", err.Details["file"])
	assert.Equal(t, "root", err.Details["user"])
	assert.Equal(t, true, err.Details["retryable"])
}

func TestRestoreError_WithDetail_Fluent(t *testing.T) {
	err := restore.NewRestoreError("NETWORK_ERROR", "Connection timeout").
		WithDetail("host", "backup-server").
		WithDetail("timeout", 30).
		WithDetail("attempts", 3)

	assert.Equal(t, "NETWORK_ERROR", err.Code)
	assert.Equal(t, "Connection timeout", err.Message)
	assert.Len(t, err.Details, 3)
	assert.Equal(t, "backup-server", err.Details["host"])
	assert.Equal(t, 30, err.Details["timeout"])
	assert.Equal(t, 3, err.Details["attempts"])
}

func TestMetadata(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	metadata := restore.Metadata{
		ID:          "backup-123",
		Type:        "filesystem",
		Namespace:   "production",
		PodName:     "web-server",
		SourcePath:  "/var/www",
		BackupFile:  "/backups/web-server-20240101.tar.gz",
		Size:        1048576,
		Checksum:    "sha256:abcd1234",
		CreatedAt:   now,
		CompletedAt: later,
		Status:      "completed",
		Compression: "gzip",
		ProviderInfo: map[string]interface{}{
			"compression_level": 6,
			"excludes":          []string{".cache", ".tmp"},
		},
	}

	assert.Equal(t, "backup-123", metadata.ID)
	assert.Equal(t, "filesystem", metadata.Type)
	assert.Equal(t, "production", metadata.Namespace)
	assert.Equal(t, "web-server", metadata.PodName)
	assert.Equal(t, "/var/www", metadata.SourcePath)
	assert.Equal(t, "/backups/web-server-20240101.tar.gz", metadata.BackupFile)
	assert.Equal(t, int64(1048576), metadata.Size)
	assert.Equal(t, "sha256:abcd1234", metadata.Checksum)
	assert.Equal(t, now, metadata.CreatedAt)
	assert.Equal(t, later, metadata.CompletedAt)
	assert.Equal(t, "completed", metadata.Status)
	assert.Equal(t, "gzip", metadata.Compression)
	assert.Equal(t, 6, metadata.ProviderInfo["compression_level"])
}

func TestRestoreResult(t *testing.T) {
	duration := time.Duration(5 * time.Minute)
	warnings := []string{"File permission changed", "Symlink target not found"}

	result := restore.RestoreResult{
		Success:      true,
		RestoredPath: "/app/data",
		FileCount:    150,
		BytesWritten: 2048000,
		Duration:     duration,
		Warnings:     warnings,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "/app/data", result.RestoredPath)
	assert.Equal(t, 150, result.FileCount)
	assert.Equal(t, int64(2048000), result.BytesWritten)
	assert.Equal(t, duration, result.Duration)
	assert.Len(t, result.Warnings, 2)
	assert.Contains(t, result.Warnings, "File permission changed")
	assert.Contains(t, result.Warnings, "Symlink target not found")
}

func TestRestoreResult_Failed(t *testing.T) {
	result := restore.RestoreResult{
		Success:      false,
		RestoredPath: "",
		FileCount:    0,
		BytesWritten: 0,
		Duration:     time.Duration(30 * time.Second),
		Warnings:     nil,
	}

	assert.False(t, result.Success)
	assert.Empty(t, result.RestoredPath)
	assert.Equal(t, 0, result.FileCount)
	assert.Equal(t, int64(0), result.BytesWritten)
	assert.Equal(t, time.Duration(30*time.Second), result.Duration)
	assert.Nil(t, result.Warnings)
}

func TestOptions_Empty(t *testing.T) {
	options := restore.Options{}

	// Test zero values
	assert.Empty(t, options.Namespace)
	assert.Empty(t, options.PodName)
	assert.Empty(t, options.BackupFile)
	assert.Empty(t, options.TargetPath)
	assert.Empty(t, options.Container)
	assert.Nil(t, options.Extra)
	assert.False(t, options.Overwrite)
	assert.False(t, options.PreservePerms)
	assert.Nil(t, options.SkipPaths)
}

func TestProgress_Percentage(t *testing.T) {
	// Test progress calculation (if needed in the future)
	progress := restore.Progress{
		Current: 50,
		Total:   100,
		Message: "50% complete",
	}

	// Manual percentage calculation for testing
	percentage := float64(progress.Current) / float64(progress.Total) * 100
	assert.Equal(t, 50.0, percentage)
}

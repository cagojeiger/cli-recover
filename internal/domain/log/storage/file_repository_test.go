package storage_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/log"
	"github.com/cagojeiger/cli-recover/internal/domain/log/storage"
)

func TestFileRepository_Save(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	require.NoError(t, err)

	// Create log
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)
	l.SetMetadata("test", "value")

	// Save log
	err = repo.Save(l)
	assert.NoError(t, err)

	// Verify metadata file exists
	metadataPath := filepath.Join(tempDir, "metadata", l.ID+".json")
	assert.FileExists(t, metadataPath)

	// Verify content
	data, err := os.ReadFile(metadataPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), l.ID)
	assert.Contains(t, string(data), "backup")
	assert.Contains(t, string(data), "filesystem")
}

func TestFileRepository_Get(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	require.NoError(t, err)

	// Create and save log
	original, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)
	original.SetMetadata("namespace", "default")
	err = repo.Save(original)
	require.NoError(t, err)

	// Get log
	retrieved, err := repo.Get(original.ID)
	assert.NoError(t, err)
	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Type, retrieved.Type)
	assert.Equal(t, original.Provider, retrieved.Provider)

	// Check metadata
	ns, ok := retrieved.GetMetadata("namespace")
	assert.True(t, ok)
	assert.Equal(t, "default", ns)

	// Get non-existent log
	_, err = repo.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileRepository_List(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	require.NoError(t, err)

	// Create multiple logs
	var logs []*log.Log
	for i := 0; i < 5; i++ {
		l, err := log.NewLog(log.TypeBackup, "filesystem")
		require.NoError(t, err)
		if i%2 == 0 {
			l.Complete()
		}
		err = repo.Save(l)
		require.NoError(t, err)
		logs = append(logs, l)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps and IDs
	}

	// Test list all
	allLogs, err := repo.List(log.ListFilter{})
	assert.NoError(t, err)
	assert.Len(t, allLogs, 5)

	// Test filter by status
	completedLogs, err := repo.List(log.ListFilter{
		Status: log.StatusCompleted,
	})
	assert.NoError(t, err)
	assert.Len(t, completedLogs, 3)

	// Test limit
	limitedLogs, err := repo.List(log.ListFilter{
		Limit: 2,
	})
	assert.NoError(t, err)
	assert.Len(t, limitedLogs, 2)

	// Verify order (newest first)
	if len(limitedLogs) >= 2 {
		assert.True(t, limitedLogs[0].StartTime.After(limitedLogs[1].StartTime) ||
			limitedLogs[0].StartTime.Equal(limitedLogs[1].StartTime))
	}
}

func TestFileRepository_Update(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	require.NoError(t, err)

	// Create and save log
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)
	err = repo.Save(l)
	require.NoError(t, err)

	// Update log
	l.Complete()
	l.SetMetadata("updated", "true")
	err = repo.Update(l)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.Get(l.ID)
	assert.NoError(t, err)
	assert.Equal(t, log.StatusCompleted, updated.Status)
	assert.NotNil(t, updated.EndTime)

	val, ok := updated.GetMetadata("updated")
	assert.True(t, ok)
	assert.Equal(t, "true", val)

	// Update non-existent log
	fake := &log.Log{
		ID:       "fake-id",
		Type:     log.TypeBackup,
		Provider: "test",
	}
	err = repo.Update(fake)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileRepository_Delete(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	require.NoError(t, err)

	// Create and save log
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)
	err = repo.Save(l)
	require.NoError(t, err)

	// Verify it exists
	_, err = repo.Get(l.ID)
	assert.NoError(t, err)

	// Delete log
	err = repo.Delete(l.ID)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = repo.Get(l.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Delete non-existent log (should not error)
	err = repo.Delete("nonexistent")
	assert.NoError(t, err)
}

func TestFileRepository_GetLatest(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	require.NoError(t, err)

	// Create logs with different types
	backup1, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)
	err = repo.Save(backup1)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	restore1, err := log.NewLog(log.TypeRestore, "filesystem")
	require.NoError(t, err)
	err = repo.Save(restore1)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	backup2, err := log.NewLog(log.TypeBackup, "minio")
	require.NoError(t, err)
	err = repo.Save(backup2)
	require.NoError(t, err)

	// Get latest backup
	latest, err := repo.GetLatest(log.ListFilter{
		Type: log.TypeBackup,
	})
	assert.NoError(t, err)
	assert.Equal(t, backup2.ID, latest.ID)

	// Get latest filesystem restore
	latest, err = repo.GetLatest(log.ListFilter{
		Type:     log.TypeRestore,
		Provider: "filesystem",
	})
	assert.NoError(t, err)
	assert.Equal(t, restore1.ID, latest.ID)

	// No matching logs
	_, err = repo.GetLatest(log.ListFilter{
		Type:     log.TypeRestore,
		Provider: "mongodb",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no logs found")
}

func TestFileRepository_CleanupOldLogs(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	repo, err := storage.NewFileRepository(tempDir)
	require.NoError(t, err)

	// Create old log
	oldLog, _ := log.NewLog(log.TypeBackup, "filesystem")
	oldLog.StartTime = time.Now().Add(-48 * time.Hour)
	repo.Save(oldLog)

	// Create recent log
	recentLog, _ := log.NewLog(log.TypeBackup, "filesystem")
	repo.Save(recentLog)

	// Cleanup logs older than 24 hours
	err = repo.CleanupOldLogs(24 * time.Hour)
	assert.NoError(t, err)

	// Verify old log is gone
	_, err = repo.Get(oldLog.ID)
	assert.Error(t, err)

	// Verify recent log still exists
	_, err = repo.Get(recentLog.ID)
	assert.NoError(t, err)
}

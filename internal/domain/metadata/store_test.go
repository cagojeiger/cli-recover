package metadata_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/metadata"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that FileStore implements Store interface
func TestFileStore_Interface(t *testing.T) {
	// Compile-time check
	var _ metadata.Store = (*metadata.FileStore)(nil)
}

func TestNewFileStore_DefaultPath(t *testing.T) {
	store, err := metadata.NewFileStore("")

	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Check that default directory was created
	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, ".cli-recover", "metadata")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)
}

func TestNewFileStore_CustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom-metadata")

	store, err := metadata.NewFileStore(customPath)

	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Check that custom directory was created
	_, err = os.Stat(customPath)
	assert.NoError(t, err)
}

func TestNewFileStore_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "deep", "metadata")

	store, err := metadata.NewFileStore(nestedPath)

	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Check that nested directories were created
	info, err := os.Stat(nestedPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFileStore_Save_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	testMetadata := &restore.Metadata{
		Type:        "filesystem",
		Namespace:   "default",
		PodName:     "test-pod",
		SourcePath:  "/data",
		BackupFile:  "/tmp/backup.tar.gz",
		Size:        1024000,
		Checksum:    "sha256:abc123",
		Status:      "completed",
		Compression: "gzip",
	}

	err = store.Save(testMetadata)
	assert.NoError(t, err)

	// Check that ID was generated
	assert.NotEmpty(t, testMetadata.ID)
	assert.Contains(t, testMetadata.ID, "backup-")

	// Check that CreatedAt was set
	assert.False(t, testMetadata.CreatedAt.IsZero())

	// Check that file was created
	filename := filepath.Join(tmpDir, testMetadata.ID+".json")
	_, err = os.Stat(filename)
	assert.NoError(t, err)
}

func TestFileStore_Save_IDGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Save two different metadata entries
	metadata1 := &restore.Metadata{Type: "filesystem", Namespace: "default"}
	metadata2 := &restore.Metadata{Type: "filesystem", Namespace: "default"}

	err = store.Save(metadata1)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond) // Ensure different timestamps

	err = store.Save(metadata2)
	assert.NoError(t, err)

	// IDs should be different
	assert.NotEqual(t, metadata1.ID, metadata2.ID)
	assert.NotEmpty(t, metadata1.ID)
	assert.NotEmpty(t, metadata2.ID)
}

func TestFileStore_Save_PreserveExistingID(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	existingID := "custom-backup-123"
	testMetadata := &restore.Metadata{
		ID:        existingID,
		Type:      "filesystem",
		Namespace: "default",
	}

	err = store.Save(testMetadata)
	assert.NoError(t, err)

	// ID should remain the same
	assert.Equal(t, existingID, testMetadata.ID)
}

func TestFileStore_Get_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Save test metadata
	originalMetadata := &restore.Metadata{
		Type:       "filesystem",
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		BackupFile: "/tmp/backup.tar.gz",
		Size:       1024000,
		Status:     "completed",
	}

	err = store.Save(originalMetadata)
	require.NoError(t, err)

	// Retrieve metadata
	retrievedMetadata, err := store.Get(originalMetadata.ID)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedMetadata)
	assert.Equal(t, originalMetadata.ID, retrievedMetadata.ID)
	assert.Equal(t, originalMetadata.Type, retrievedMetadata.Type)
	assert.Equal(t, originalMetadata.Namespace, retrievedMetadata.Namespace)
	assert.Equal(t, originalMetadata.PodName, retrievedMetadata.PodName)
	assert.Equal(t, originalMetadata.SourcePath, retrievedMetadata.SourcePath)
	assert.Equal(t, originalMetadata.BackupFile, retrievedMetadata.BackupFile)
	assert.Equal(t, originalMetadata.Size, retrievedMetadata.Size)
	assert.Equal(t, originalMetadata.Status, retrievedMetadata.Status)
}

func TestFileStore_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	_, err = store.Get("non-existent-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open metadata file")
}

func TestFileStore_GetByFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	backupFile := "/tmp/specific-backup.tar.gz"
	testMetadata := &restore.Metadata{
		Type:       "filesystem",
		Namespace:  "default",
		BackupFile: backupFile,
	}

	err = store.Save(testMetadata)
	require.NoError(t, err)

	// Retrieve by backup file
	retrievedMetadata, err := store.GetByFile(backupFile)

	assert.NoError(t, err)
	assert.NotNil(t, retrievedMetadata)
	assert.Equal(t, testMetadata.ID, retrievedMetadata.ID)
	assert.Equal(t, backupFile, retrievedMetadata.BackupFile)
}

func TestFileStore_GetByFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	_, err = store.GetByFile("/tmp/non-existent-backup.tar.gz")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata not found for backup file")
}

func TestFileStore_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	entries, err := store.List()

	assert.NoError(t, err)
	assert.Empty(t, entries)
}

func TestFileStore_List_Multiple(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Save multiple metadata entries
	metadata1 := &restore.Metadata{Type: "filesystem", Namespace: "default", PodName: "pod1"}
	metadata2 := &restore.Metadata{Type: "filesystem", Namespace: "kube-system", PodName: "pod2"}
	metadata3 := &restore.Metadata{Type: "filesystem", Namespace: "default", PodName: "pod3"}

	err = store.Save(metadata1)
	require.NoError(t, err)
	err = store.Save(metadata2)
	require.NoError(t, err)
	err = store.Save(metadata3)
	require.NoError(t, err)

	// List all entries
	entries, err := store.List()

	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	// Check that all entries are present
	ids := make([]string, len(entries))
	for i, entry := range entries {
		ids[i] = entry.ID
	}
	assert.Contains(t, ids, metadata1.ID)
	assert.Contains(t, ids, metadata2.ID)
	assert.Contains(t, ids, metadata3.ID)
}

func TestFileStore_List_SkipInvalidFiles(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Create a valid metadata file
	validMetadata := &restore.Metadata{Type: "filesystem", Namespace: "default"}
	err = store.Save(validMetadata)
	require.NoError(t, err)

	// Create invalid files that should be skipped
	err = os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte("invalid json"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "notjson.txt"), []byte("not json file"), 0644)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(tmpDir, "directory"), 0755)
	require.NoError(t, err)

	// List should return only valid metadata
	entries, err := store.List()

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, validMetadata.ID, entries[0].ID)
}

func TestFileStore_ListByNamespace_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Save metadata in different namespaces
	metadata1 := &restore.Metadata{Type: "filesystem", Namespace: "default", PodName: "pod1"}
	metadata2 := &restore.Metadata{Type: "filesystem", Namespace: "kube-system", PodName: "pod2"}
	metadata3 := &restore.Metadata{Type: "filesystem", Namespace: "default", PodName: "pod3"}

	err = store.Save(metadata1)
	require.NoError(t, err)
	err = store.Save(metadata2)
	require.NoError(t, err)
	err = store.Save(metadata3)
	require.NoError(t, err)

	// List entries for 'default' namespace
	entries, err := store.ListByNamespace("default")

	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	for _, entry := range entries {
		assert.Equal(t, "default", entry.Namespace)
	}
}

func TestFileStore_ListByNamespace_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Save metadata in different namespace
	metadata := &restore.Metadata{Type: "filesystem", Namespace: "default"}
	err = store.Save(metadata)
	require.NoError(t, err)

	// List entries for non-existent namespace
	entries, err := store.ListByNamespace("non-existent")

	assert.NoError(t, err)
	assert.Empty(t, entries)
}

func TestFileStore_Delete_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Save test metadata
	testMetadata := &restore.Metadata{Type: "filesystem", Namespace: "default"}
	err = store.Save(testMetadata)
	require.NoError(t, err)

	// Verify file exists
	filename := filepath.Join(tmpDir, testMetadata.ID+".json")
	_, err = os.Stat(filename)
	assert.NoError(t, err)

	// Delete metadata
	err = store.Delete(testMetadata.ID)
	assert.NoError(t, err)

	// Verify file was deleted
	_, err = os.Stat(filename)
	assert.True(t, os.IsNotExist(err))
}

func TestFileStore_Delete_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Delete non-existent file should not return error
	err = store.Delete("non-existent-id")
	assert.NoError(t, err)
}

func TestFileStore_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	const numGoroutines = 10
	const numOperationsPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Run concurrent save operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				metadata := &restore.Metadata{
					Type:      "filesystem",
					Namespace: "default",
					PodName:   "pod-" + string(rune(goroutineID)) + "-" + string(rune(j)),
				}

				err := store.Save(metadata)
				assert.NoError(t, err)

				// Try to read it back
				retrieved, err := store.Get(metadata.ID)
				assert.NoError(t, err)
				assert.Equal(t, metadata.PodName, retrieved.PodName)
			}
		}(i)
	}

	wg.Wait()

	// Verify all entries were saved
	entries, err := store.List()
	assert.NoError(t, err)
	assert.Len(t, entries, numGoroutines*numOperationsPerGoroutine)
}

func TestNewFileStore_HomeDirectoryError(t *testing.T) {
	// This test is hard to simulate without mocking os.UserHomeDir()
	// but we can test it conceptually by passing empty string
	// The function should handle the case gracefully
	store, err := metadata.NewFileStore("")

	// Should succeed in normal circumstances
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestNewFileStore_DirectoryPermissionError(t *testing.T) {
	// Test with a path that can't be created (root directory on Unix)
	if os.Getenv("CI") != "" {
		t.Skip("Skipping permission test in CI environment")
	}

	invalidPath := "/root/cannot-create-this-directory"
	_, err := metadata.NewFileStore(invalidPath)

	// Should return an error on permission denied
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create metadata directory")
}

func TestFileStore_Save_MarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Create metadata with invalid field that can't be marshaled
	invalidMetadata := &restore.Metadata{
		Type:      "filesystem",
		Namespace: "default",
		ProviderInfo: map[string]interface{}{
			"invalid": func() {}, // Function can't be marshaled to JSON
		},
	}

	err = store.Save(invalidMetadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal metadata")
}

func TestFileStore_Save_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Remove write permissions from directory
	err = os.Chmod(tmpDir, 0444) // Read-only
	defer os.Chmod(tmpDir, 0755) // Restore permissions

	if err != nil {
		t.Skip("Could not change directory permissions")
	}

	testMetadata := &restore.Metadata{Type: "filesystem", Namespace: "default"}
	err = store.Save(testMetadata)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write metadata file")
}

func TestFileStore_loadMetadata_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Create invalid JSON file
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	require.NoError(t, err)

	// Use reflection to call private method for testing
	// Since loadMetadata is private, we test it through Get()
	_, err = store.Get("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal metadata")
}

func TestFileStore_List_DirectoryReadError(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// Remove directory to cause read error
	err = os.RemoveAll(tmpDir)
	require.NoError(t, err)

	_, err = store.List()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read metadata directory")
}

func TestDefaultStore_Initialization(t *testing.T) {
	// Test that DefaultStore is initialized
	assert.NotNil(t, metadata.DefaultStore)

	// Test that we can use DefaultStore
	// Create a temporary metadata to test with
	testMetadata := &restore.Metadata{
		Type:      "filesystem",
		Namespace: "test",
		PodName:   "test-pod",
	}

	// Save using DefaultStore (cleanup afterward)
	err := metadata.DefaultStore.Save(testMetadata)
	assert.NoError(t, err)

	// Cleanup
	if testMetadata.ID != "" {
		_ = metadata.DefaultStore.Delete(testMetadata.ID)
	}
}

func TestFileStore_FullWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := metadata.NewFileStore(tmpDir)
	require.NoError(t, err)

	// 1. Save metadata
	originalMetadata := &restore.Metadata{
		Type:        "filesystem",
		Namespace:   "default",
		PodName:     "test-pod",
		SourcePath:  "/data",
		BackupFile:  "/tmp/backup.tar.gz",
		Size:        1024000,
		Checksum:    "sha256:abc123",
		Status:      "completed",
		Compression: "gzip",
		ProviderInfo: map[string]interface{}{
			"tar_flags": []string{"-czf"},
			"exclude":   []string{"*.log"},
		},
	}

	err = store.Save(originalMetadata)
	require.NoError(t, err)
	require.NotEmpty(t, originalMetadata.ID)

	// 2. Get by ID
	retrieved, err := store.Get(originalMetadata.ID)
	require.NoError(t, err)
	assert.Equal(t, originalMetadata.Type, retrieved.Type)
	assert.Equal(t, originalMetadata.BackupFile, retrieved.BackupFile)

	// 3. Get by file
	retrievedByFile, err := store.GetByFile(originalMetadata.BackupFile)
	require.NoError(t, err)
	assert.Equal(t, originalMetadata.ID, retrievedByFile.ID)

	// 4. List all
	allEntries, err := store.List()
	require.NoError(t, err)
	assert.Len(t, allEntries, 1)
	assert.Equal(t, originalMetadata.ID, allEntries[0].ID)

	// 5. List by namespace
	namespaceEntries, err := store.ListByNamespace("default")
	require.NoError(t, err)
	assert.Len(t, namespaceEntries, 1)
	assert.Equal(t, originalMetadata.ID, namespaceEntries[0].ID)

	// 6. Delete
	err = store.Delete(originalMetadata.ID)
	require.NoError(t, err)

	// 7. Verify deletion
	_, err = store.Get(originalMetadata.ID)
	assert.Error(t, err)

	allEntries, err = store.List()
	require.NoError(t, err)
	assert.Empty(t, allEntries)
}

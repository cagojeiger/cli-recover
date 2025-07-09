package metadata_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/metadata"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Edge case and concurrent access tests

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
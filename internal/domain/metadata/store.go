package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// Store defines the interface for metadata storage
type Store interface {
	// Save saves backup metadata
	Save(metadata *restore.Metadata) error

	// Get retrieves metadata by ID
	Get(id string) (*restore.Metadata, error)

	// GetByFile retrieves metadata by backup file path
	GetByFile(backupFile string) (*restore.Metadata, error)

	// List returns all metadata entries
	List() ([]*restore.Metadata, error)

	// ListByNamespace returns metadata for a specific namespace
	ListByNamespace(namespace string) ([]*restore.Metadata, error)

	// Delete removes metadata by ID
	Delete(id string) error
}

// FileStore implements Store using local filesystem
type FileStore struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFileStore creates a new file-based metadata store
func NewFileStore(baseDir string) (*FileStore, error) {
	// Expand home directory if needed
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, ".cli-recover", "metadata")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	return &FileStore{
		baseDir: baseDir,
	}, nil
}

// Save saves backup metadata
func (s *FileStore) Save(metadata *restore.Metadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if metadata.ID == "" {
		metadata.ID = generateID()
	}

	// Update timestamps
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write to file
	filename := filepath.Join(s.baseDir, metadata.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// Get retrieves metadata by ID
func (s *FileStore) Get(id string) (*restore.Metadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.baseDir, id+".json")
	return s.loadMetadata(filename)
}

// GetByFile retrieves metadata by backup file path
func (s *FileStore) GetByFile(backupFile string) (*restore.Metadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := s.list()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.BackupFile == backupFile {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("metadata not found for backup file: %s", backupFile)
}

// List returns all metadata entries
func (s *FileStore) List() ([]*restore.Metadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.list()
}

// ListByNamespace returns metadata for a specific namespace
func (s *FileStore) ListByNamespace(namespace string) ([]*restore.Metadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allEntries, err := s.list()
	if err != nil {
		return nil, err
	}

	var filtered []*restore.Metadata
	for _, entry := range allEntries {
		if entry.Namespace == namespace {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// Delete removes metadata by ID
func (s *FileStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.baseDir, id+".json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata file: %w", err)
	}

	return nil
}

// list reads all metadata files (internal helper)
func (s *FileStore) list() ([]*restore.Metadata, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %w", err)
	}

	var result []*restore.Metadata
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filename := filepath.Join(s.baseDir, entry.Name())
		metadata, err := s.loadMetadata(filename)
		if err != nil {
			// Log error but continue
			continue
		}

		result = append(result, metadata)
	}

	return result, nil
}

// loadMetadata loads metadata from a file
func (s *FileStore) loadMetadata(filename string) (*restore.Metadata, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata restore.Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// generateID generates a unique ID for metadata
func generateID() string {
	return fmt.Sprintf("backup-%d", time.Now().UnixNano())
}

// DefaultStore is the default metadata store instance
var DefaultStore Store

func init() {
	// Initialize default store
	store, err := NewFileStore("")
	if err != nil {
		// Fallback to temporary directory
		tmpDir := filepath.Join(os.TempDir(), "cli-recover-metadata")
		store, _ = NewFileStore(tmpDir)
	}
	DefaultStore = store
}

package storage

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/log"
)

// FileRepository implements log.Repository using filesystem
type FileRepository struct {
	mu      sync.RWMutex
	baseDir string
	index   map[string]*log.Log // In-memory index for fast lookups
}

// NewFileRepository creates a new file-based log repository
func NewFileRepository(baseDir string) (*FileRepository, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	repo := &FileRepository{
		baseDir: baseDir,
		index:   make(map[string]*log.Log),
	}

	// Load existing logs into index
	if err := repo.loadIndex(); err != nil {
		return nil, fmt.Errorf("failed to load log index: %w", err)
	}

	return repo, nil
}

// Save saves a log entry
func (r *FileRepository) Save(l *log.Log) error {
	if err := l.Validate(); err != nil {
		return fmt.Errorf("invalid log: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate metadata file path
	metadataPath := r.getMetadataPath(l.ID)
	
	// Create directory if needed
	dir := filepath.Dir(metadataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Save metadata
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Update index
	r.index[l.ID] = l

	return nil
}

// Get retrieves a log by ID
func (r *FileRepository) Get(id string) (*log.Log, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check index first
	if l, ok := r.index[id]; ok {
		return l, nil
	}

	// Load from file if not in index
	metadataPath := r.getMetadataPath(id)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("log not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var l log.Log
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("failed to unmarshal log: %w", err)
	}

	// Update index
	r.index[id] = &l

	return &l, nil
}

// List returns all logs with optional filters
func (r *FileRepository) List(filter log.ListFilter) ([]*log.Log, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*log.Log

	// Iterate through all logs
	for _, l := range r.index {
		// Apply filters
		if filter.Type != "" && l.Type != filter.Type {
			continue
		}
		if filter.Provider != "" && l.Provider != filter.Provider {
			continue
		}
		if filter.Status != "" && l.Status != filter.Status {
			continue
		}
		if filter.StartDate != nil && l.StartTime.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && l.StartTime.After(*filter.EndDate) {
			continue
		}

		result = append(result, l)
	}

	// Sort by start time (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.After(result[j].StartTime)
	})

	// Apply limit
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result, nil
}

// Update updates an existing log
func (r *FileRepository) Update(l *log.Log) error {
	if err := l.Validate(); err != nil {
		return fmt.Errorf("invalid log: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if log exists
	if _, ok := r.index[l.ID]; !ok {
		return fmt.Errorf("log not found: %s", l.ID)
	}

	// Save updated metadata
	metadataPath := r.getMetadataPath(l.ID)
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Update index
	r.index[l.ID] = l

	return nil
}

// Delete removes a log entry
func (r *FileRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove metadata file
	metadataPath := r.getMetadataPath(id)
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata: %w", err)
	}

	// Remove from index
	delete(r.index, id)

	return nil
}

// GetLatest returns the most recent log matching the filter
func (r *FileRepository) GetLatest(filter log.ListFilter) (*log.Log, error) {
	filter.Limit = 1
	logs, err := r.List(filter)
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return nil, fmt.Errorf("no logs found matching filter")
	}

	return logs[0], nil
}

// loadIndex loads all log metadata files into memory
func (r *FileRepository) loadIndex() error {
	metadataDir := filepath.Join(r.baseDir, "metadata")
	
	// Create metadata directory if it doesn't exist
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return err
	}

	// Walk through metadata files
	err := filepath.WalkDir(metadataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .json files
		if filepath.Ext(path) != ".json" {
			return nil
		}

		// Read and parse metadata
		data, err := os.ReadFile(path)
		if err != nil {
			// Log error but continue
			return nil
		}

		var l log.Log
		if err := json.Unmarshal(data, &l); err != nil {
			// Log error but continue
			return nil
		}

		// Add to index
		r.index[l.ID] = &l

		return nil
	})

	return err
}

// getMetadataPath returns the path to the metadata file for a log
func (r *FileRepository) getMetadataPath(id string) string {
	return filepath.Join(r.baseDir, "metadata", id+".json")
}

// GetLogDir returns the base directory for log files
func (r *FileRepository) GetLogDir() string {
	return filepath.Join(r.baseDir, "files")
}

// CleanupOldLogs removes logs older than the specified duration
func (r *FileRepository) CleanupOldLogs(maxAge time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var toDelete []string

	for id, l := range r.index {
		if l.StartTime.Before(cutoff) {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		// Get log to find associated log file
		l := r.index[id]
		
		// Remove log file if it exists
		if l.FilePath != "" {
			os.Remove(l.FilePath)
		}

		// Remove metadata
		metadataPath := r.getMetadataPath(id)
		os.Remove(metadataPath)

		// Remove from index
		delete(r.index, id)
	}

	return nil
}
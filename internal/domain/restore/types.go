package restore

import (
	"time"
)

// Progress represents restore operation progress
type Progress struct {
	Current int64  // Current bytes processed
	Total   int64  // Total bytes to process
	Message string // Progress message
}

// Options contains restore operation options
type Options struct {
	// Common fields
	Namespace  string                 // Kubernetes namespace
	PodName    string                 // Target pod name
	BackupFile string                 // Path to backup file
	TargetPath string                 // Target restore path
	Container  string                 // Container name (optional)
	Extra      map[string]interface{} // Provider-specific options

	// Restore-specific options
	Overwrite     bool     // Overwrite existing files
	PreservePerms bool     // Preserve file permissions
	SkipPaths     []string // Paths to skip during restore
}

// RestoreError represents a restore-specific error
type RestoreError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (e *RestoreError) Error() string {
	return e.Message
}

// NewRestoreError creates a new restore error
func NewRestoreError(code, message string) *RestoreError {
	return &RestoreError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the error
func (e *RestoreError) WithDetail(key string, value interface{}) *RestoreError {
	e.Details[key] = value
	return e
}

// Metadata represents backup metadata for restore operations
type Metadata struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Namespace    string                 `json:"namespace"`
	PodName      string                 `json:"pod_name"`
	SourcePath   string                 `json:"source_path"`
	BackupFile   string                 `json:"backup_file"`
	Size         int64                  `json:"size"`
	Checksum     string                 `json:"checksum"`
	CreatedAt    time.Time              `json:"created_at"`
	CompletedAt  time.Time              `json:"completed_at"`
	Status       string                 `json:"status"`
	Compression  string                 `json:"compression"`
	ProviderInfo map[string]interface{} `json:"provider_info"`
}

// RestoreResult contains the result of a restore operation
type RestoreResult struct {
	Success      bool
	RestoredPath string
	FileCount    int
	BytesWritten int64
	Duration     time.Duration
	Warnings     []string
}

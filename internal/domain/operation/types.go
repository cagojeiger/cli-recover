package operation

import (
	"time"
)

// Options contains configuration for backup/restore operations
type Options struct {
	// Type of operation
	Type ProviderType
	
	// Common options
	Namespace  string
	PodName    string
	Container  string
	
	// Backup-specific options
	SourcePath string
	OutputFile string
	Compress   bool
	Exclude    []string
	
	// Restore-specific options
	BackupFile    string
	TargetPath    string
	Overwrite     bool
	PreservePerms bool
	SkipPaths     []string
	
	// Extra provider-specific options
	Extra map[string]interface{}
}

// Result represents the outcome of an operation
type Result struct {
	// Success indicates if the operation completed successfully
	Success bool
	
	// Message provides a summary of the operation
	Message string
	
	// Error contains any error that occurred
	Error error
	
	// Common result fields
	BytesWritten int64
	FileCount    int
	Duration     time.Duration
	
	// Restore-specific fields
	RestoredPath string
	
	// Warnings collected during operation
	Warnings []string
}

// Metadata represents stored information about a completed operation
type Metadata struct {
	// Unique identifier
	ID string `json:"id" yaml:"id"`
	
	// Type of operation (backup/restore)
	Type string `json:"type" yaml:"type"`
	
	// Provider name
	Provider string `json:"provider" yaml:"provider"`
	
	// Kubernetes info
	Namespace string `json:"namespace" yaml:"namespace"`
	PodName   string `json:"pod_name" yaml:"pod_name"`
	Container string `json:"container,omitempty" yaml:"container,omitempty"`
	
	// Backup-specific metadata
	SourcePath  string `json:"source_path,omitempty" yaml:"source_path,omitempty"`
	BackupFile  string `json:"backup_file,omitempty" yaml:"backup_file,omitempty"`
	Compression string `json:"compression,omitempty" yaml:"compression,omitempty"`
	
	// Restore-specific metadata
	TargetPath   string `json:"target_path,omitempty" yaml:"target_path,omitempty"`
	RestoredFrom string `json:"restored_from,omitempty" yaml:"restored_from,omitempty"`
	
	// Common metadata
	Size        int64     `json:"size" yaml:"size"`
	FileCount   int       `json:"file_count" yaml:"file_count"`
	Checksum    string    `json:"checksum,omitempty" yaml:"checksum,omitempty"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	CompletedAt time.Time `json:"completed_at" yaml:"completed_at"`
	Status      string    `json:"status" yaml:"status"`
	
	// Provider-specific information
	ProviderInfo map[string]interface{} `json:"provider_info,omitempty" yaml:"provider_info,omitempty"`
	
	// Additional metadata
	Extra map[string]string `json:"extra,omitempty" yaml:"extra,omitempty"`
}

// SetMetadata sets a metadata key-value pair
func (m *Metadata) SetMetadata(key, value string) {
	if m.Extra == nil {
		m.Extra = make(map[string]string)
	}
	m.Extra[key] = value
}
package metadata

import (
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

// DefaultStore is the default metadata store instance
var DefaultStore Store

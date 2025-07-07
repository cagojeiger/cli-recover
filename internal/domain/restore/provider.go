package restore

import (
	"context"
)

// Provider defines the interface for restore providers
type Provider interface {
	// Name returns the provider name
	Name() string
	
	// Description returns a human-readable description
	Description() string
	
	// ValidateOptions validates the restore options
	ValidateOptions(opts Options) error
	
	// ValidateBackup validates that the backup file is compatible
	ValidateBackup(backupFile string, metadata *Metadata) error
	
	// Execute performs the restore operation
	Execute(ctx context.Context, opts Options) (*RestoreResult, error)
	
	// StreamProgress returns a channel for progress updates
	StreamProgress() <-chan Progress
	
	// EstimateSize estimates the size of data to be restored
	EstimateSize(backupFile string) (int64, error)
}
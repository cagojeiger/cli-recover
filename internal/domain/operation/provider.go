package operation

import (
	"context"
)

// ProviderType defines the type of operation a provider performs
type ProviderType string

const (
	// TypeBackup indicates a backup provider
	TypeBackup ProviderType = "backup"
	// TypeRestore indicates a restore provider
	TypeRestore ProviderType = "restore"
)

// Provider defines the unified interface for both backup and restore providers
type Provider interface {
	// Name returns the unique name of the provider
	Name() string
	
	// Description returns a human-readable description
	Description() string
	
	// Type returns the provider type (backup or restore)
	Type() ProviderType
	
	// ValidateOptions validates provider-specific options
	ValidateOptions(opts Options) error
	
	// Execute performs the operation (backup or restore)
	Execute(ctx context.Context, opts Options) (*Result, error)
	
	// EstimateSize estimates the size of the operation in bytes
	EstimateSize(opts Options) (int64, error)
	
	// StreamProgress returns a channel that streams progress updates
	StreamProgress() <-chan Progress
}

// Progress represents the current progress of an operation
type Progress struct {
	// Current bytes processed
	Current int64
	// Total bytes to process (0 if unknown)
	Total int64
	// Human-readable status message
	Message string
}
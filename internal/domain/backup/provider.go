package backup

import (
	"context"
)

// Provider defines the interface for backup providers
type Provider interface {
	// Name returns the unique name of the provider
	Name() string

	// Description returns a human-readable description
	Description() string

	// Execute performs the backup operation
	Execute(ctx context.Context, opts Options) error

	// EstimateSize estimates the size of the backup in bytes
	EstimateSize(opts Options) (int64, error)

	// StreamProgress returns a channel that streams progress updates
	StreamProgress() <-chan Progress

	// ValidateOptions validates provider-specific options
	ValidateOptions(opts Options) error
}

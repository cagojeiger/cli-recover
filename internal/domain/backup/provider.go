package backup

import (
	"context"
	"fmt"
	"sync"
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

// ProviderRegistry manages registered backup providers
type ProviderRegistry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *ProviderRegistry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves a provider by name
func (r *ProviderRegistry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// List returns all registered provider names
func (r *ProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	return names
}

// DefaultRegistry is the global provider registry
var DefaultRegistry = NewProviderRegistry()
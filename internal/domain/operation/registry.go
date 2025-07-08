package operation

import (
	"fmt"
	"sync"
)

// ProviderRegistry manages registered providers for both backup and restore
type ProviderRegistry struct {
	providers map[string]map[ProviderType]Provider
	mu        sync.RWMutex
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]map[ProviderType]Provider),
	}
}

// Register adds a provider to the registry
func (r *ProviderRegistry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	providerType := provider.Type()
	
	// Initialize the inner map if needed
	if r.providers[name] == nil {
		r.providers[name] = make(map[ProviderType]Provider)
	}
	
	// Check if provider already exists
	if _, exists := r.providers[name][providerType]; exists {
		return fmt.Errorf("provider %s of type %s already registered", name, providerType)
	}

	r.providers[name][providerType] = provider
	return nil
}

// Get retrieves a provider by name and type
func (r *ProviderRegistry) Get(name string, providerType ProviderType) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if providers, exists := r.providers[name]; exists {
		if provider, exists := providers[providerType]; exists {
			return provider, nil
		}
	}
	
	return nil, fmt.Errorf("provider %s of type %s not found", name, providerType)
}

// List returns all registered providers
func (r *ProviderRegistry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []Provider
	for _, typeMap := range r.providers {
		for _, provider := range typeMap {
			providers = append(providers, provider)
		}
	}

	return providers
}

// ListByType returns all registered providers of a specific type
func (r *ProviderRegistry) ListByType(providerType ProviderType) []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []Provider
	for _, typeMap := range r.providers {
		if provider, exists := typeMap[providerType]; exists {
			providers = append(providers, provider)
		}
	}

	return providers
}

// DefaultRegistry is the global provider registry
var DefaultRegistry = NewProviderRegistry()
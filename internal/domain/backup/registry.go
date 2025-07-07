package backup

import (
	"fmt"
	"sort"
	"sync"
)

// ProviderFactory is a function that creates a new Provider instance
type ProviderFactory func() Provider

// Registry manages provider factories
type Registry struct {
	factories map[string]ProviderFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]ProviderFactory),
	}
}

// RegisterFactory registers a provider factory
func (r *Registry) RegisterFactory(name string, factory ProviderFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("provider factory %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Create creates a new provider instance by name
func (r *Registry) Create(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[name]
	if !exists {
		return nil, fmt.Errorf("provider factory %s not found", name)
	}

	return factory(), nil
}

// Available returns sorted list of available provider names
func (r *Registry) Available() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// ProviderInfo contains information about a provider
type ProviderInfo struct {
	Name        string
	Description string
}

// Info returns information about all available providers
func (r *Registry) Info() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(r.factories))
	for name, factory := range r.factories {
		provider := factory()
		infos = append(infos, ProviderInfo{
			Name:        name,
			Description: provider.Description(),
		})
	}

	// Sort by name for consistent output
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// GlobalRegistry is the default registry instance
var GlobalRegistry = NewRegistry()
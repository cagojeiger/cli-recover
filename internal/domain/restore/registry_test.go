package restore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockProvider for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string                                      { return m.name }
func (m *mockProvider) Description() string                               { return "Mock provider" }
func (m *mockProvider) ValidateOptions(opts Options) error                { return nil }
func (m *mockProvider) ValidateBackup(backupFile string, metadata *Metadata) error { return nil }
func (m *mockProvider) Execute(ctx context.Context, opts Options) (*RestoreResult, error) {
	return &RestoreResult{Success: true}, nil
}
func (m *mockProvider) StreamProgress() <-chan Progress                   { return make(chan Progress) }
func (m *mockProvider) EstimateSize(backupFile string) (int64, error)     { return 0, nil }

func TestRegistry_RegisterFactory(t *testing.T) {
	t.Run("register new provider", func(t *testing.T) {
		registry := NewRegistry()
		
		err := registry.RegisterFactory("test", func() Provider {
			return &mockProvider{name: "test"}
		})
		
		assert.NoError(t, err)
		assert.Contains(t, registry.Available(), "test")
	})

	t.Run("register duplicate provider", func(t *testing.T) {
		registry := NewRegistry()
		
		// First registration
		err := registry.RegisterFactory("test", func() Provider {
			return &mockProvider{name: "test"}
		})
		assert.NoError(t, err)
		
		// Duplicate registration
		err = registry.RegisterFactory("test", func() Provider {
			return &mockProvider{name: "test2"}
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("register with nil factory", func(t *testing.T) {
		registry := NewRegistry()
		
		err := registry.RegisterFactory("test", nil)
		
		// The current implementation doesn't check for nil factory, so no error
		assert.NoError(t, err)
	})
}

func TestRegistry_Create(t *testing.T) {
	t.Run("create registered provider", func(t *testing.T) {
		registry := NewRegistry()
		
		// Register provider
		registry.RegisterFactory("test", func() Provider {
			return &mockProvider{name: "test"}
		})
		
		// Create provider
		provider, err := registry.Create("test")
		
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "test", provider.Name())
	})

	t.Run("create unregistered provider", func(t *testing.T) {
		registry := NewRegistry()
		
		provider, err := registry.Create("unknown")
		
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("factory returns nil", func(t *testing.T) {
		registry := NewRegistry()
		
		// Register factory that returns nil
		registry.RegisterFactory("bad", func() Provider {
			return nil
		})
		
		provider, err := registry.Create("bad")
		
		// The current implementation doesn't check for nil provider
		assert.NoError(t, err)
		assert.Nil(t, provider)
	})
}

func TestRegistry_Available(t *testing.T) {
	t.Run("list empty registry", func(t *testing.T) {
		registry := NewRegistry()
		
		list := registry.Available()
		
		assert.Empty(t, list)
	})

	t.Run("list with providers", func(t *testing.T) {
		registry := NewRegistry()
		
		registry.RegisterFactory("provider1", func() Provider {
			return &mockProvider{name: "provider1"}
		})
		registry.RegisterFactory("provider2", func() Provider {
			return &mockProvider{name: "provider2"}
		})
		
		list := registry.Available()
		
		assert.Len(t, list, 2)
		assert.Contains(t, list, "provider1")
		assert.Contains(t, list, "provider2")
	})
}

func TestGlobalRegistry(t *testing.T) {
	t.Run("global registry is initialized", func(t *testing.T) {
		assert.NotNil(t, GlobalRegistry)
	})

	t.Run("register and create via global registry", func(t *testing.T) {
		// Clear any existing registrations
		GlobalRegistry = NewRegistry()
		
		// Register provider
		err := GlobalRegistry.RegisterFactory("global-test", func() Provider {
			return &mockProvider{name: "global-test"}
		})
		assert.NoError(t, err)
		
		// Create provider
		provider, err := GlobalRegistry.Create("global-test")
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "global-test", provider.Name())
		
		// List providers
		list := GlobalRegistry.Available()
		assert.Contains(t, list, "global-test")
	})
}

func TestRegistry_Info(t *testing.T) {
	t.Run("get provider info", func(t *testing.T) {
		registry := NewRegistry()
		
		registry.RegisterFactory("test1", func() Provider {
			return &mockProvider{name: "test1"}
		})
		registry.RegisterFactory("test2", func() Provider {
			return &mockProvider{name: "test2"}
		})
		
		infos := registry.Info()
		
		assert.Len(t, infos, 2)
		// Info should be sorted by name
		assert.Equal(t, "test1", infos[0].Name)
		assert.Equal(t, "test2", infos[1].Name)
		assert.Equal(t, "Mock provider", infos[0].Description)
	})
}
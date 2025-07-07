package backup_test

import (
	"context"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/stretchr/testify/assert"
)

// Mock provider for testing
type mockProvider struct {
	name        string
	description string
}

func (m *mockProvider) Name() string        { return m.name }
func (m *mockProvider) Description() string { return m.description }
func (m *mockProvider) Execute(ctx context.Context, opts backup.Options) error { return nil }
func (m *mockProvider) EstimateSize(opts backup.Options) (int64, error)    { return 0, nil }
func (m *mockProvider) StreamProgress() <-chan backup.Progress             { return nil }
func (m *mockProvider) ValidateOptions(opts backup.Options) error          { return nil }

func newMockProvider(name, description string) backup.Provider {
	return &mockProvider{name: name, description: description}
}

func TestNewRegistry(t *testing.T) {
	registry := backup.NewRegistry()
	
	assert.NotNil(t, registry)
	assert.Empty(t, registry.Available())
}

func TestRegistry_RegisterFactory(t *testing.T) {
	registry := backup.NewRegistry()
	
	// Register a provider factory
	err := registry.RegisterFactory("test-provider", func() backup.Provider {
		return newMockProvider("test-provider", "Test provider")
	})
	
	assert.NoError(t, err)
	assert.Contains(t, registry.Available(), "test-provider")
}

func TestRegistry_RegisterFactory_Duplicate(t *testing.T) {
	registry := backup.NewRegistry()
	
	factory := func() backup.Provider {
		return newMockProvider("test-provider", "Test provider")
	}
	
	// Register first time - should succeed
	err := registry.RegisterFactory("test-provider", factory)
	assert.NoError(t, err)
	
	// Register same name again - should fail
	err = registry.RegisterFactory("test-provider", factory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_Create(t *testing.T) {
	registry := backup.NewRegistry()
	
	// Register a provider factory
	err := registry.RegisterFactory("test-provider", func() backup.Provider {
		return newMockProvider("test-provider", "Test provider")
	})
	assert.NoError(t, err)
	
	// Create provider instance
	provider, err := registry.Create("test-provider")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "test-provider", provider.Name())
	assert.Equal(t, "Test provider", provider.Description())
}

func TestRegistry_Create_NotFound(t *testing.T) {
	registry := backup.NewRegistry()
	
	// Try to create non-existent provider
	provider, err := registry.Create("non-existent")
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_Create_MultipleInstances(t *testing.T) {
	registry := backup.NewRegistry()
	
	// Register a provider factory
	err := registry.RegisterFactory("test-provider", func() backup.Provider {
		return newMockProvider("test-provider", "Test provider")
	})
	assert.NoError(t, err)
	
	// Create multiple instances
	provider1, err := registry.Create("test-provider")
	assert.NoError(t, err)
	
	provider2, err := registry.Create("test-provider")
	assert.NoError(t, err)
	
	// Should be different instances
	assert.NotSame(t, provider1, provider2)
	assert.Equal(t, provider1.Name(), provider2.Name())
}

func TestRegistry_Available(t *testing.T) {
	registry := backup.NewRegistry()
	
	// Empty registry
	available := registry.Available()
	assert.Empty(t, available)
	
	// Register multiple providers
	registry.RegisterFactory("provider-a", func() backup.Provider {
		return newMockProvider("provider-a", "Provider A")
	})
	registry.RegisterFactory("provider-b", func() backup.Provider {
		return newMockProvider("provider-b", "Provider B")
	})
	registry.RegisterFactory("provider-c", func() backup.Provider {
		return newMockProvider("provider-c", "Provider C")
	})
	
	// Check available providers
	available = registry.Available()
	assert.Len(t, available, 3)
	assert.Contains(t, available, "provider-a")
	assert.Contains(t, available, "provider-b")
	assert.Contains(t, available, "provider-c")
	
	// Should be sorted
	assert.Equal(t, []string{"provider-a", "provider-b", "provider-c"}, available)
}

func TestRegistry_Info(t *testing.T) {
	registry := backup.NewRegistry()
	
	// Empty registry
	info := registry.Info()
	assert.Empty(t, info)
	
	// Register providers
	registry.RegisterFactory("provider-z", func() backup.Provider {
		return newMockProvider("provider-z", "Provider Z")
	})
	registry.RegisterFactory("provider-a", func() backup.Provider {
		return newMockProvider("provider-a", "Provider A")
	})
	registry.RegisterFactory("provider-m", func() backup.Provider {
		return newMockProvider("provider-m", "Provider M")
	})
	
	// Check provider info
	info = registry.Info()
	assert.Len(t, info, 3)
	
	// Should be sorted by name
	assert.Equal(t, "provider-a", info[0].Name)
	assert.Equal(t, "Provider A", info[0].Description)
	assert.Equal(t, "provider-m", info[1].Name)
	assert.Equal(t, "Provider M", info[1].Description)
	assert.Equal(t, "provider-z", info[2].Name)
	assert.Equal(t, "Provider Z", info[2].Description)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := backup.NewRegistry()
	
	// Register a provider
	err := registry.RegisterFactory("test-provider", func() backup.Provider {
		return newMockProvider("test-provider", "Test provider")
	})
	assert.NoError(t, err)
	
	// Test concurrent reads
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// Concurrent reads should not cause issues
			available := registry.Available()
			assert.Contains(t, available, "test-provider")
			
			provider, err := registry.Create("test-provider")
			assert.NoError(t, err)
			assert.NotNil(t, provider)
			
			info := registry.Info()
			assert.Len(t, info, 1)
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Test that global registry exists and works
	assert.NotNil(t, backup.GlobalRegistry)
	
	// Should be able to register and create providers
	err := backup.GlobalRegistry.RegisterFactory("global-test", func() backup.Provider {
		return newMockProvider("global-test", "Global test provider")
	})
	assert.NoError(t, err)
	
	provider, err := backup.GlobalRegistry.Create("global-test")
	assert.NoError(t, err)
	assert.Equal(t, "global-test", provider.Name())
	
	// Clean up for other tests
	// Note: This is a bit hacky since there's no unregister method
	// In a real scenario, you might want to add an unregister method
}
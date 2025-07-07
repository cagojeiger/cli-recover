package backup_test

import (
	"context"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of Provider interface
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Execute(ctx context.Context, opts backup.Options) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func (m *MockProvider) EstimateSize(opts backup.Options) (int64, error) {
	args := m.Called(opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProvider) StreamProgress() <-chan backup.Progress {
	args := m.Called()
	if ch := args.Get(0); ch != nil {
		return ch.(<-chan backup.Progress)
	}
	return nil
}

func (m *MockProvider) ValidateOptions(opts backup.Options) error {
	args := m.Called(opts)
	return args.Error(0)
}

func TestMockProvider_Name(t *testing.T) {
	mockProvider := new(MockProvider)
	mockProvider.On("Name").Return("filesystem")

	name := mockProvider.Name()
	assert.Equal(t, "filesystem", name)
	mockProvider.AssertExpectations(t)
}

func TestMockProvider_Execute(t *testing.T) {
	mockProvider := new(MockProvider)
	ctx := context.Background()
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar.gz",
	}

	mockProvider.On("Execute", ctx, opts).Return(nil)

	err := mockProvider.Execute(ctx, opts)
	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
}

func TestMockProvider_EstimateSize(t *testing.T) {
	mockProvider := new(MockProvider)
	opts := backup.Options{
		Namespace:  "default",
		PodName:    "test-pod",
		SourcePath: "/data",
		OutputFile: "backup.tar.gz",
	}

	expectedSize := int64(1048576) // 1MB
	mockProvider.On("EstimateSize", opts).Return(expectedSize, nil)

	size, err := mockProvider.EstimateSize(opts)
	assert.NoError(t, err)
	assert.Equal(t, expectedSize, size)
	mockProvider.AssertExpectations(t)
}

func TestMockProvider_StreamProgress(t *testing.T) {
	mockProvider := new(MockProvider)
	
	progressCh := make(chan backup.Progress, 1)
	progressCh <- backup.Progress{
		Current: 50,
		Total:   100,
		Message: "50% complete",
	}
	close(progressCh)

	mockProvider.On("StreamProgress").Return((<-chan backup.Progress)(progressCh))

	ch := mockProvider.StreamProgress()
	assert.NotNil(t, ch)
	
	progress := <-ch
	assert.Equal(t, int64(50), progress.Current)
	assert.Equal(t, int64(100), progress.Total)
	assert.Equal(t, "50% complete", progress.Message)
	
	mockProvider.AssertExpectations(t)
}

func TestProviderRegistry_Register(t *testing.T) {
	registry := backup.NewProviderRegistry()
	mockProvider := new(MockProvider)
	mockProvider.On("Name").Return("test")

	err := registry.Register(mockProvider)
	assert.NoError(t, err)

	// Try to register again - should fail
	err = registry.Register(mockProvider)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestProviderRegistry_Get(t *testing.T) {
	registry := backup.NewProviderRegistry()
	mockProvider := new(MockProvider)
	mockProvider.On("Name").Return("test")

	registry.Register(mockProvider)

	// Get existing provider
	provider, err := registry.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, mockProvider, provider)

	// Get non-existing provider
	provider, err = registry.Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "not found")
}

func TestProviderRegistry_List(t *testing.T) {
	registry := backup.NewProviderRegistry()
	
	// Empty registry
	names := registry.List()
	assert.Empty(t, names)

	// Add providers
	mock1 := new(MockProvider)
	mock1.On("Name").Return("provider1")
	mock2 := new(MockProvider)
	mock2.On("Name").Return("provider2")

	registry.Register(mock1)
	registry.Register(mock2)

	names = registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "provider1")
	assert.Contains(t, names, "provider2")
}
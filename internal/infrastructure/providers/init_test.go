package providers_test

import (
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/providers"
	"github.com/stretchr/testify/assert"
)

func TestRegisterProviders_Success(t *testing.T) {
	// Clear existing registrations to ensure clean test
	backup.GlobalRegistry = backup.NewRegistry()
	restore.GlobalRegistry = restore.NewRegistry()
	
	// Create mock dependencies
	mockKubeClient := new(kubernetes.MockKubeClient)
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	// Register providers
	err := providers.RegisterProviders(mockKubeClient, mockExecutor)
	assert.NoError(t, err)
	
	// Verify backup provider is registered
	backupProvider, err := backup.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	assert.NotNil(t, backupProvider)
	assert.Equal(t, "filesystem", backupProvider.Name())
	
	// Verify restore provider is registered
	restoreProvider, err := restore.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	assert.NotNil(t, restoreProvider)
	assert.Equal(t, "filesystem", restoreProvider.Name())
}

func TestRegisterProviders_FilesystemProviderDetails(t *testing.T) {
	// Clear existing registrations
	backup.GlobalRegistry = backup.NewRegistry()
	restore.GlobalRegistry = restore.NewRegistry()
	
	// Create mock dependencies
	mockKubeClient := new(kubernetes.MockKubeClient)
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	// Register providers
	err := providers.RegisterProviders(mockKubeClient, mockExecutor)
	assert.NoError(t, err)
	
	// Test backup provider details
	backupProvider, err := backup.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	assert.Equal(t, "filesystem", backupProvider.Name())
	assert.Equal(t, "Backup filesystem from Kubernetes pods", backupProvider.Description())
	
	// Test restore provider details
	restoreProvider, err := restore.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	assert.Equal(t, "filesystem", restoreProvider.Name())
	assert.Equal(t, "Restore filesystem to Kubernetes pods", restoreProvider.Description())
}

func TestRegisterProviders_UnknownProvider(t *testing.T) {
	// Clear existing registrations
	backup.GlobalRegistry = backup.NewRegistry()
	restore.GlobalRegistry = restore.NewRegistry()
	
	// Create mock dependencies
	mockKubeClient := new(kubernetes.MockKubeClient)
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	// Register providers
	err := providers.RegisterProviders(mockKubeClient, mockExecutor)
	assert.NoError(t, err)
	
	// Try to create unknown provider
	_, err = backup.GlobalRegistry.Create("unknown-provider")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	_, err = restore.GlobalRegistry.Create("unknown-provider")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegisterProviders_ProviderFactories(t *testing.T) {
	// Clear existing registrations
	backup.GlobalRegistry = backup.NewRegistry()
	restore.GlobalRegistry = restore.NewRegistry()
	
	// Create mock dependencies
	mockKubeClient := new(kubernetes.MockKubeClient)
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	// Register providers
	err := providers.RegisterProviders(mockKubeClient, mockExecutor)
	assert.NoError(t, err)
	
	// Create multiple instances to ensure factory pattern works
	backupProvider1, err := backup.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	
	backupProvider2, err := backup.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	
	// Providers should be different instances but same type
	assert.NotSame(t, backupProvider1, backupProvider2)
	assert.Equal(t, backupProvider1.Name(), backupProvider2.Name())
	
	// Same for restore providers
	restoreProvider1, err := restore.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	
	restoreProvider2, err := restore.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	
	assert.NotSame(t, restoreProvider1, restoreProvider2)
	assert.Equal(t, restoreProvider1.Name(), restoreProvider2.Name())
}

func TestRegisterProviders_WithNilDependencies(t *testing.T) {
	// Clear existing registrations
	backup.GlobalRegistry = backup.NewRegistry()
	restore.GlobalRegistry = restore.NewRegistry()
	
	// Test with nil dependencies (should still work but providers might not function)
	err := providers.RegisterProviders(nil, nil)
	assert.NoError(t, err)
	
	// Verify providers are still registered
	backupProvider, err := backup.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	assert.NotNil(t, backupProvider)
	
	restoreProvider, err := restore.GlobalRegistry.Create("filesystem")
	assert.NoError(t, err)
	assert.NotNil(t, restoreProvider)
}
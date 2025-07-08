package infrastructure_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/infrastructure"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
)

func TestCreateBackupProvider(t *testing.T) {
	mockKubeClient := new(kubernetes.MockKubeClient)
	mockExecutor := new(kubernetes.MockCommandExecutor)

	tests := []struct {
		name        string
		provider    string
		expectError bool
	}{
		{
			name:        "filesystem provider",
			provider:    "filesystem",
			expectError: false,
		},
		{
			name:        "unknown provider",
			provider:    "unknown",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := infrastructure.CreateBackupProvider(tt.provider, mockKubeClient, mockExecutor)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, tt.provider, provider.Name())
			}
		})
	}
}

func TestCreateRestoreProvider(t *testing.T) {
	mockKubeClient := new(kubernetes.MockKubeClient)
	mockExecutor := new(kubernetes.MockCommandExecutor)

	tests := []struct {
		name        string
		provider    string
		expectError bool
	}{
		{
			name:        "filesystem provider",
			provider:    "filesystem",
			expectError: false,
		},
		{
			name:        "unknown provider",
			provider:    "unknown",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := infrastructure.CreateRestoreProvider(tt.provider, mockKubeClient, mockExecutor)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, tt.provider, provider.Name())
			}
		})
	}
}

func TestGetAvailableProviders(t *testing.T) {
	backupProviders := infrastructure.GetAvailableBackupProviders()
	assert.Contains(t, backupProviders, "filesystem")
	assert.Len(t, backupProviders, 1) // Only filesystem for now

	restoreProviders := infrastructure.GetAvailableRestoreProviders()
	assert.Contains(t, restoreProviders, "filesystem")
	assert.Len(t, restoreProviders, 1) // Only filesystem for now
}

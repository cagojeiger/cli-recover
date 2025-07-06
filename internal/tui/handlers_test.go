package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockRunner implements the runner.Runner interface for testing
type mockRunner struct {
	output string
	err    error
}

func (m *mockRunner) Run(cmd string, args ...string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(m.output), nil
}

func TestHandleBackupTypeEnter(t *testing.T) {
	// Create a mock runner
	mockRunner := &mockRunner{
		output: `{
			"items": [
				{"metadata": {"name": "default"}},
				{"metadata": {"name": "kube-system"}}
			]
		}`,
		err: nil,
	}

	// Initialize model
	m := InitialModel(mockRunner)
	m.screen = ScreenBackupType
	
	tests := []struct {
		name           string
		selected       int
		expectedType   string
		expectedScreen Screen
	}{
		{
			name:           "Select filesystem backup",
			selected:       0,
			expectedType:   "filesystem",
			expectedScreen: ScreenNamespaceList,
		},
		{
			name:           "Select minio backup",
			selected:       1,
			expectedType:   "minio",
			expectedScreen: ScreenNamespaceList,
		},
		{
			name:           "Select mongodb backup",
			selected:       2,
			expectedType:   "mongodb",
			expectedScreen: ScreenNamespaceList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.selected = tt.selected
			result := handleBackupTypeEnter(m)
			
			assert.Equal(t, tt.expectedType, result.selectedBackupType)
			assert.Equal(t, tt.expectedScreen, result.screen)
			assert.NotEmpty(t, result.namespaces)
			assert.Equal(t, 0, result.selected) // Should reset selection
		})
	}
}

func TestHandleBack_BackupTypeScreen(t *testing.T) {
	mockRunner := &mockRunner{}
	m := InitialModel(mockRunner)
	
	// Test backing out from backup type screen
	m.screen = ScreenBackupType
	result := handleBack(m)
	
	assert.Equal(t, ScreenMain, result.screen)
	assert.Equal(t, 0, result.selected)
	
	// Test backing out from namespace list when coming from backup type
	m.screen = ScreenNamespaceList
	m.selectedBackupType = "filesystem"
	result = handleBack(m)
	
	assert.Equal(t, ScreenBackupType, result.screen)
	assert.Equal(t, 0, result.selected)
}

func TestGetMaxItems_BackupTypeScreen(t *testing.T) {
	mockRunner := &mockRunner{}
	m := InitialModel(mockRunner)
	m.screen = ScreenBackupType
	
	maxItems := getMaxItems(m)
	assert.Equal(t, 3, maxItems) // filesystem, minio, mongodb
}
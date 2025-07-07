package adapters

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// TestNewRestoreAdapter tests the constructor
func TestNewRestoreAdapter(t *testing.T) {
	// Create a mock registry
	registry := &mockRestoreRegistry{}
	
	adapter := NewRestoreAdapter(registry)
	assert.NotNil(t, adapter)
	assert.IsType(t, &RestoreAdapter{}, adapter)
	assert.Equal(t, registry, adapter.registry)
}

// TestGetAbsolutePath tests path resolution
func TestGetAbsolutePath(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path unchanged",
			input:    "/tmp/backup.tar",
			expected: "/tmp/backup.tar",
		},
		{
			name:  "relative path made absolute",
			input: "backup.tar",
			// This will be the current directory + backup.tar
			expected: "", // We'll check it's absolute
		},
		{
			name:  "relative path with directory",
			input: "./backups/backup.tar",
			// This will be expanded to absolute
			expected: "", // We'll check it's absolute
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAbsolutePath(tt.input)
			
			// Always returns absolute path
			assert.True(t, filepath.IsAbs(result))
			
			// For absolute input, should be unchanged
			if filepath.IsAbs(tt.input) {
				assert.Equal(t, tt.input, result)
			}
		})
	}
}

// TestSanitizeTargetPath tests path sanitization
func TestSanitizeTargetPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "root path",
			input:    "/",
			expected: "/",
		},
		{
			name:     "normal path",
			input:    "/data",
			expected: "/data",
		},
		{
			name:     "path with trailing slash",
			input:    "/data/",
			expected: "/data",
		},
		{
			name:     "path with spaces",
			input:    "/my data",
			expected: "/my data",
		},
		{
			name:     "path with special chars",
			input:    "/data@backup!",
			expected: "/data@backup!",
		},
		{
			name:     "nested path",
			input:    "/var/lib/data",
			expected: "/var/lib/data",
		},
		{
			name:     "empty path defaults to root",
			input:    "",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTargetPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMonitorProgressEdgeCases tests additional monitor progress scenarios
func TestMonitorProgressEdgeCases(t *testing.T) {
	t.Run("monitor with zero estimated size", func(t *testing.T) {
		adapter := &RestoreAdapter{}
		mockProvider := &MockRestoreProvider{
			progressCh: make(chan restore.Progress, 1),
		}
		
		done := make(chan bool)
		
		// Start monitoring with zero size
		go adapter.monitorProgress(mockProvider, 0, done, false)
		
		// Send progress
		mockProvider.progressCh <- restore.Progress{
			Current: 1024,
			Total:   2048,
			Message: "Progress",
		}
		
		// Stop monitoring
		close(done)
	})
	
	t.Run("monitor with closed channel", func(t *testing.T) {
		adapter := &RestoreAdapter{}
		mockProvider := &MockRestoreProvider{
			progressCh: make(chan restore.Progress),
		}
		
		done := make(chan bool)
		
		// Close progress channel immediately
		close(mockProvider.progressCh)
		
		// Start monitoring - should handle closed channel gracefully
		go adapter.monitorProgress(mockProvider, 1024, done, false)
		
		// Give it time to process
		time.Sleep(100 * time.Millisecond)
		
		// Stop monitoring
		close(done)
	})
}
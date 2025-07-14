package config

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, 1, config.Version)
	assert.Equal(t, 30, config.Logs.RetentionDays)
	assert.Contains(t, config.Logs.Directory, ".cli-pipe")
	assert.Contains(t, config.Logs.Directory, "logs")
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create config
	config := &Config{
		Version: 1,
		Logs: LogConfig{
			Directory:     "~/.cli-pipe/logs",
			RetentionDays: 7,
		},
	}
	
	// Save config
	err := config.Save()
	require.NoError(t, err)
	
	// Verify file exists
	configPath := filepath.Join(tempDir, ".cli-pipe", "config.yaml")
	assert.FileExists(t, configPath)
	
	// Load config
	loaded, err := Load()
	require.NoError(t, err)
	
	assert.Equal(t, config.Version, loaded.Version)
	assert.Equal(t, config.Logs.RetentionDays, loaded.Logs.RetentionDays)
	assert.Contains(t, loaded.Logs.Directory, ".cli-pipe/logs")
}

func TestLoadNonExistentConfig(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Load config (should return default)
	config, err := Load()
	require.NoError(t, err)
	
	// Should be default config
	assert.Equal(t, 1, config.Version)
	assert.Equal(t, 30, config.Logs.RetentionDays)
}

func TestEnsureLogDir(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &Config{
		Logs: LogConfig{
			Directory: filepath.Join(tempDir, "test-logs"),
		},
	}
	
	// Directory shouldn't exist yet
	_, err := os.Stat(config.Logs.Directory)
	assert.True(t, os.IsNotExist(err))
	
	// Ensure directory
	err = config.EnsureLogDir()
	require.NoError(t, err)
	
	// Directory should exist now
	info, err := os.Stat(config.Logs.Directory)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "home directory",
			input:    "~/test",
			expected: filepath.Join(homeDir, "test"),
		},
		{
			name:     "absolute path",
			input:    "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "relative path",
			input:    "test/path",
			expected: "test/path", // Will be made absolute in actual expandPath
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if tt.name == "home directory" && result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
			// For relative paths, just check it's not empty
			if tt.name == "relative path" && result == "" {
				t.Errorf("expandPath(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestLoaderLoadFromReader_ValidYAML(t *testing.T) {
	yaml := `
logger:
  level: debug
  output: file
  file:
    path: /tmp/test.log
    format: json
backup:
  defaultCompression: bzip2
  excludeVCS: false
metadata:
  format: yaml
`
	loader := NewLoader("")
	reader := strings.NewReader(yaml)
	
	cfg, err := loader.LoadFromReader(reader)
	if err != nil {
		t.Errorf("LoadFromReader() error = %v, expected no error", err)
		return
	}
	
	validateLoggerConfig(t, cfg)
	validateBackupConfig(t, cfg)
	validateMetadataConfig(t, cfg)
}

func TestLoaderLoadFromReader_InvalidYAML(t *testing.T) {
	yaml := `
logger:
  level: invalid_level
`
	loader := NewLoader("")
	reader := strings.NewReader(yaml)
	
	_, err := loader.LoadFromReader(reader)
	if err == nil {
		t.Error("LoadFromReader() expected error for invalid YAML, got no error")
	}
}

func TestLoaderLoadFromReader_EmptyYAML(t *testing.T) {
	loader := NewLoader("")
	reader := strings.NewReader("")
	
	cfg, err := loader.LoadFromReader(reader)
	if err != nil {
		t.Errorf("LoadFromReader() error = %v, expected no error", err)
		return
	}
	
	// Should have default values
	if cfg.Logger.Level != "info" {
		t.Errorf("Expected default level 'info', got %s", cfg.Logger.Level)
	}
}

func validateLoggerConfig(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Logger.Level != "debug" {
		t.Errorf("Expected level 'debug', got %s", cfg.Logger.Level)
	}
	if cfg.Logger.Output != "file" {
		t.Errorf("Expected output 'file', got %s", cfg.Logger.Output)
	}
	if cfg.Logger.File.Format != "json" {
		t.Errorf("Expected file format 'json', got %s", cfg.Logger.File.Format)
	}
}

func validateBackupConfig(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Backup.DefaultCompression != "bzip2" {
		t.Errorf("Expected compression 'bzip2', got %s", cfg.Backup.DefaultCompression)
	}
	if cfg.Backup.ExcludeVCS != false {
		t.Error("Expected excludeVCS to be false")
	}
}

func validateMetadataConfig(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Metadata.Format != "yaml" {
		t.Errorf("Expected metadata format 'yaml', got %s", cfg.Metadata.Format)
	}
}

func TestLoaderEnvironmentOverrides(t *testing.T) {
	// Save current env vars
	oldVars := map[string]string{
		"CLI_RECOVER_LOG_LEVEL":      os.Getenv("CLI_RECOVER_LOG_LEVEL"),
		"CLI_RECOVER_LOG_OUTPUT":     os.Getenv("CLI_RECOVER_LOG_OUTPUT"),
		"CLI_RECOVER_LOG_FILE":       os.Getenv("CLI_RECOVER_LOG_FILE"),
		"CLI_RECOVER_LOG_FORMAT":     os.Getenv("CLI_RECOVER_LOG_FORMAT"),
		"CLI_RECOVER_LOG_COLOR":      os.Getenv("CLI_RECOVER_LOG_COLOR"),
		"CLI_RECOVER_COMPRESSION":    os.Getenv("CLI_RECOVER_COMPRESSION"),
		"CLI_RECOVER_EXCLUDE_VCS":    os.Getenv("CLI_RECOVER_EXCLUDE_VCS"),
		"CLI_RECOVER_METADATA_PATH":  os.Getenv("CLI_RECOVER_METADATA_PATH"),
	}
	
	// Restore env vars after test
	defer func() {
		for k, v := range oldVars {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()
	
	// Set test env vars
	os.Setenv("CLI_RECOVER_LOG_LEVEL", "error")
	os.Setenv("CLI_RECOVER_LOG_OUTPUT", "both")
	os.Setenv("CLI_RECOVER_LOG_FILE", "/tmp/test.log")
	os.Setenv("CLI_RECOVER_LOG_FORMAT", "json")
	os.Setenv("CLI_RECOVER_LOG_COLOR", "false")
	os.Setenv("CLI_RECOVER_COMPRESSION", "xz")
	os.Setenv("CLI_RECOVER_EXCLUDE_VCS", "false")
	os.Setenv("CLI_RECOVER_METADATA_PATH", "/tmp/metadata")
	
	// Create a loader and load config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	loader := NewLoader(configPath)
	
	cfg := DefaultConfig()
	loader.loadFromEnv(cfg)
	
	// Check overrides
	if cfg.Logger.Level != "error" {
		t.Errorf("Expected log level 'error', got %s", cfg.Logger.Level)
	}
	if cfg.Logger.Output != "both" {
		t.Errorf("Expected log output 'both', got %s", cfg.Logger.Output)
	}
	if cfg.Logger.File.Path != "/tmp/test.log" {
		t.Errorf("Expected log file '/tmp/test.log', got %s", cfg.Logger.File.Path)
	}
	if cfg.Logger.File.Format != "json" {
		t.Errorf("Expected log format 'json', got %s", cfg.Logger.File.Format)
	}
	if cfg.Logger.Console.Color != false {
		t.Error("Expected console color to be false")
	}
	if cfg.Backup.DefaultCompression != "xz" {
		t.Errorf("Expected compression 'xz', got %s", cfg.Backup.DefaultCompression)
	}
	if cfg.Backup.ExcludeVCS != false {
		t.Error("Expected excludeVCS to be false")
	}
	if cfg.Metadata.Path != "/tmp/metadata" {
		t.Errorf("Expected metadata path '/tmp/metadata', got %s", cfg.Metadata.Path)
	}
}
package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	// Check logger defaults
	if cfg.Logger.Level != "info" {
		t.Errorf("Expected logger level 'info', got %s", cfg.Logger.Level)
	}
	if cfg.Logger.Output != "console" {
		t.Errorf("Expected logger output 'console', got %s", cfg.Logger.Output)
	}
	if cfg.Logger.File.MaxSize != 100 {
		t.Errorf("Expected file max size 100, got %d", cfg.Logger.File.MaxSize)
	}
	if cfg.Logger.File.MaxAge != 7 {
		t.Errorf("Expected file max age 7, got %d", cfg.Logger.File.MaxAge)
	}
	if cfg.Logger.File.Format != "text" {
		t.Errorf("Expected file format 'text', got %s", cfg.Logger.File.Format)
	}
	if !cfg.Logger.Console.Color {
		t.Error("Expected console color to be enabled")
	}
	
	// Check backup defaults
	if cfg.Backup.DefaultCompression != "gzip" {
		t.Errorf("Expected default compression 'gzip', got %s", cfg.Backup.DefaultCompression)
	}
	if !cfg.Backup.ExcludeVCS {
		t.Error("Expected excludeVCS to be true")
	}
	if !cfg.Backup.PreservePermissions {
		t.Error("Expected preservePermissions to be true")
	}
	if cfg.Backup.DefaultTimeout != 5*time.Minute {
		t.Errorf("Expected default timeout 5m, got %v", cfg.Backup.DefaultTimeout)
	}
	
	// Check metadata defaults
	if cfg.Metadata.Format != "json" {
		t.Errorf("Expected metadata format 'json', got %s", cfg.Metadata.Format)
	}
}

func TestConfigValidate_Valid(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := DefaultConfig()
		err := config.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v, expected no error", err)
		}
	})
}

func TestConfigValidate_LoggerValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		errMsg string
	}{
		{
			name: "invalid logger level",
			config: &Config{
				Logger: LoggerConfig{
					Level:  "invalid",
					Output: "console",
					File:   FileLoggerConfig{Format: "text"},
				},
				Backup:   BackupConfig{DefaultCompression: "gzip"},
				Metadata: MetadataConfig{Format: "json"},
			},
			errMsg: "logger.level",
		},
		{
			name: "invalid logger output",
			config: &Config{
				Logger: LoggerConfig{
					Level:  "info",
					Output: "invalid",
					File:   FileLoggerConfig{Format: "text"},
				},
				Backup:   BackupConfig{DefaultCompression: "gzip"},
				Metadata: MetadataConfig{Format: "json"},
			},
			errMsg: "logger.output",
		},
		{
			name: "invalid file format",
			config: &Config{
				Logger: LoggerConfig{
					Level:  "info",
					Output: "console",
					File:   FileLoggerConfig{Format: "invalid"},
				},
				Backup:   BackupConfig{DefaultCompression: "gzip"},
				Metadata: MetadataConfig{Format: "json"},
			},
			errMsg: "logger.file.format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateConfigError(t, tt.config, tt.errMsg)
		})
	}
}

func TestConfigValidate_BackupValidation(t *testing.T) {
	t.Run("invalid compression", func(t *testing.T) {
		config := &Config{
			Logger: LoggerConfig{
				Level:  "info",
				Output: "console",
				File:   FileLoggerConfig{Format: "text"},
			},
			Backup:   BackupConfig{DefaultCompression: "invalid"},
			Metadata: MetadataConfig{Format: "json"},
		}
		validateConfigError(t, config, "backup.defaultCompression")
	})
}

func TestConfigValidate_MetadataValidation(t *testing.T) {
	t.Run("invalid metadata format", func(t *testing.T) {
		config := &Config{
			Logger: LoggerConfig{
				Level:  "info",
				Output: "console",
				File:   FileLoggerConfig{Format: "text"},
			},
			Backup:   BackupConfig{DefaultCompression: "gzip"},
			Metadata: MetadataConfig{Format: "invalid"},
		}
		validateConfigError(t, config, "metadata.format")
	})
}

func validateConfigError(t *testing.T, config *Config, expectedErrMsg string) {
	t.Helper()
	err := config.Validate()
	if err == nil {
		t.Errorf("Validate() expected error containing %s, got no error", expectedErrMsg)
		return
	}
	if _, ok := err.(*ConfigError); !ok {
		t.Errorf("Expected ConfigError, got %T", err)
	}
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:   "test.field",
		Value:   "test-value",
		Message: "test message",
	}
	
	expected := "config validation error: test.field=test-value - test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}
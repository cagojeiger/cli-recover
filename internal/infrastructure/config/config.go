package config

import (
	"path/filepath"
	"time"
)

// Config represents the complete configuration for cli-recover
type Config struct {
	Logger   LoggerConfig   `yaml:"logger"`
	Backup   BackupConfig   `yaml:"backup"`
	Metadata MetadataConfig `yaml:"metadata"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Level   string           `yaml:"level"`  // debug, info, warn, error, fatal
	Output  string           `yaml:"output"` // console, file, both
	File    FileLoggerConfig `yaml:"file"`
	Console ConsoleConfig    `yaml:"console"`
}

// FileLoggerConfig represents file logger specific configuration
type FileLoggerConfig struct {
	Path    string `yaml:"path"`    // Log file path
	MaxSize int    `yaml:"maxSize"` // Maximum size in MB
	MaxAge  int    `yaml:"maxAge"`  // Maximum age in days
	Format  string `yaml:"format"`  // text, json
}

// ConsoleConfig represents console logger configuration
type ConsoleConfig struct {
	Color bool `yaml:"color"` // Enable/disable color output
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	DefaultCompression  string        `yaml:"defaultCompression"`  // gzip, bzip2, xz, none
	ExcludeVCS          bool          `yaml:"excludeVCS"`          // Exclude version control systems
	PreservePermissions bool          `yaml:"preservePermissions"` // Preserve file permissions
	DefaultTimeout      time.Duration `yaml:"defaultTimeout"`      // Default operation timeout
}

// MetadataConfig represents metadata storage configuration
type MetadataConfig struct {
	Path   string `yaml:"path"`   // Metadata storage path
	Format string `yaml:"format"` // json, yaml
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir := "~/.cli-recover"

	return &Config{
		Logger: LoggerConfig{
			Level:  "info",
			Output: "console",
			File: FileLoggerConfig{
				Path:    filepath.Join(homeDir, "logs", "cli-recover.log"),
				MaxSize: 100,
				MaxAge:  7,
				Format:  "text",
			},
			Console: ConsoleConfig{
				Color: true,
			},
		},
		Backup: BackupConfig{
			DefaultCompression:  "gzip",
			ExcludeVCS:          true,
			PreservePermissions: true,
			DefaultTimeout:      5 * time.Minute,
		},
		Metadata: MetadataConfig{
			Path:   filepath.Join(homeDir, "metadata"),
			Format: "json",
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate logger level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[c.Logger.Level] {
		return &ConfigError{Field: "logger.level", Value: c.Logger.Level, Message: "invalid log level"}
	}

	// Validate logger output
	validOutputs := map[string]bool{
		"console": true,
		"file":    true,
		"both":    true,
	}
	if !validOutputs[c.Logger.Output] {
		return &ConfigError{Field: "logger.output", Value: c.Logger.Output, Message: "invalid output type"}
	}

	// Validate file format
	validFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validFormats[c.Logger.File.Format] {
		return &ConfigError{Field: "logger.file.format", Value: c.Logger.File.Format, Message: "invalid file format"}
	}

	// Validate compression
	validCompressions := map[string]bool{
		"gzip":  true,
		"bzip2": true,
		"xz":    true,
		"none":  true,
	}
	if !validCompressions[c.Backup.DefaultCompression] {
		return &ConfigError{Field: "backup.defaultCompression", Value: c.Backup.DefaultCompression, Message: "invalid compression type"}
	}

	// Validate metadata format
	validMetadataFormats := map[string]bool{
		"json": true,
		"yaml": true,
	}
	if !validMetadataFormats[c.Metadata.Format] {
		return &ConfigError{Field: "metadata.format", Value: c.Metadata.Format, Message: "invalid metadata format"}
	}

	return nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Value   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config validation error: " + e.Field + "=" + e.Value + " - " + e.Message
}

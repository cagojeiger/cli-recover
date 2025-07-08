package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader handles configuration loading from various sources
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: expandPath(configPath),
	}
}

// Load loads configuration from file
func (l *Loader) Load() (*Config, error) {
	// Start with default config
	cfg := DefaultConfig()
	
	// If config file doesn't exist, return default
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return cfg, nil
	}
	
	// Load from file
	if err := l.loadFromFile(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", l.configPath, err)
	}
	
	// Override with environment variables
	l.loadFromEnv(cfg)
	
	// Validate final configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	
	return cfg, nil
}

// LoadFromReader loads configuration from an io.Reader
func (l *Loader) LoadFromReader(r io.Reader) (*Config, error) {
	cfg := DefaultConfig()
	
	decoder := yaml.NewDecoder(r)
	err := decoder.Decode(cfg)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}
	
	// Expand paths
	cfg.Logger.File.Path = expandPath(cfg.Logger.File.Path)
	cfg.Metadata.Path = expandPath(cfg.Metadata.Path)
	
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	
	return cfg, nil
}

// loadFromFile loads configuration from a file
func (l *Loader) loadFromFile(cfg *Config) error {
	file, err := os.Open(l.configPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return err
	}
	
	// Expand paths
	cfg.Logger.File.Path = expandPath(cfg.Logger.File.Path)
	cfg.Metadata.Path = expandPath(cfg.Metadata.Path)
	
	return nil
}

// loadFromEnv loads configuration from environment variables
func (l *Loader) loadFromEnv(cfg *Config) {
	// Logger settings
	if level := os.Getenv("CLI_RECOVER_LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}
	if output := os.Getenv("CLI_RECOVER_LOG_OUTPUT"); output != "" {
		cfg.Logger.Output = output
	}
	if filePath := os.Getenv("CLI_RECOVER_LOG_FILE"); filePath != "" {
		cfg.Logger.File.Path = expandPath(filePath)
	}
	if format := os.Getenv("CLI_RECOVER_LOG_FORMAT"); format != "" {
		cfg.Logger.File.Format = format
	}
	if color := os.Getenv("CLI_RECOVER_LOG_COLOR"); color == "false" {
		cfg.Logger.Console.Color = false
	}
	
	// Backup settings
	if compression := os.Getenv("CLI_RECOVER_COMPRESSION"); compression != "" {
		cfg.Backup.DefaultCompression = compression
	}
	if excludeVCS := os.Getenv("CLI_RECOVER_EXCLUDE_VCS"); excludeVCS == "false" {
		cfg.Backup.ExcludeVCS = false
	}
	
	// Metadata settings
	if metadataPath := os.Getenv("CLI_RECOVER_METADATA_PATH"); metadataPath != "" {
		cfg.Metadata.Path = expandPath(metadataPath)
	}
}

// expandPath expands ~ to home directory and resolves relative paths
func expandPath(path string) string {
	if path == "" {
		return path
	}
	
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	
	// Make absolute if relative
	if !filepath.IsAbs(path) {
		if absPath, err := filepath.Abs(path); err == nil {
			path = absPath
		}
	}
	
	return path
}

// ConfigPath returns the default configuration file path
func ConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".cli-recover/config.yaml"
	}
	return filepath.Join(homeDir, ".cli-recover", "config.yaml")
}
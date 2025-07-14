package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// Config represents the cli-pipe configuration
type Config struct {
	Version int        `yaml:"version"`
	Logs    LogConfig  `yaml:"logs"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Directory      string `yaml:"directory"`
	RetentionDays  int    `yaml:"retention_days"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Version: 1,
		Logs: LogConfig{
			Directory:     filepath.Join(homeDir, ".cli-pipe", "logs"),
			RetentionDays: 30,
		},
	}
}

// ConfigDir returns the cli-pipe config directory
func ConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cli-pipe")
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configPath := ConfigPath()
	
	// If config doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Expand home directory in paths
	config.expandPaths()
	
	return &config, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	// Ensure config directory exists
	configDir := ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write to file
	configPath := ConfigPath()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// expandPaths expands ~ in paths to the home directory
func (c *Config) expandPaths() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	if len(c.Logs.Directory) > 0 && c.Logs.Directory[0] == '~' {
		c.Logs.Directory = filepath.Join(homeDir, c.Logs.Directory[1:])
	}
}

// EnsureLogDir creates the log directory if it doesn't exist
func (c *Config) EnsureLogDir() error {
	return os.MkdirAll(c.Logs.Directory, 0755)
}
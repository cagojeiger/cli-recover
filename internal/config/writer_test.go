package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWriterWrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	writer := NewWriter(configPath)
	cfg := DefaultConfig()
	
	// Write config
	err := writer.Write(cfg)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	// Check file exists
	if !writer.Exists() {
		t.Error("Config file should exist after writing")
	}
	
	// Read back and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	
	var readCfg Config
	if err := yaml.Unmarshal(data, &readCfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}
	
	// Verify some fields
	if readCfg.Logger.Level != cfg.Logger.Level {
		t.Errorf("Expected logger level %s, got %s", cfg.Logger.Level, readCfg.Logger.Level)
	}
	if readCfg.Backup.DefaultCompression != cfg.Backup.DefaultCompression {
		t.Errorf("Expected compression %s, got %s", cfg.Backup.DefaultCompression, readCfg.Backup.DefaultCompression)
	}
}

func TestWriterBackup(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	backupPath := configPath + ".backup"
	
	writer := NewWriter(configPath)
	
	// Write initial config
	cfg1 := DefaultConfig()
	cfg1.Logger.Level = "debug"
	if err := writer.Write(cfg1); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}
	
	// Write new config (should create backup)
	cfg2 := DefaultConfig()
	cfg2.Logger.Level = "error"
	if err := writer.Write(cfg2); err != nil {
		t.Fatalf("Failed to write new config: %v", err)
	}
	
	// Check backup exists
	if _, err := os.Stat(backupPath); err != nil {
		t.Error("Backup file should exist")
	}
	
	// Verify backup contains old config
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	
	var backupCfg Config
	if err := yaml.Unmarshal(data, &backupCfg); err != nil {
		t.Fatalf("Failed to unmarshal backup config: %v", err)
	}
	
	if backupCfg.Logger.Level != "debug" {
		t.Errorf("Backup should contain old config with level 'debug', got %s", backupCfg.Logger.Level)
	}
}

func TestWriterCreateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	writer := NewWriter(configPath)
	
	// Create default config
	err := writer.CreateDefaultConfig()
	if err != nil {
		t.Fatalf("Failed to create default config: %v", err)
	}
	
	// Check file exists
	if !writer.Exists() {
		t.Error("Config file should exist after creating default")
	}
	
	// Read and check content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	
	content := string(data)
	
	// Check for expected content
	expectedStrings := []string{
		"# CLI-Recover Configuration File",
		"logger:",
		"level: info",
		"backup:",
		"defaultCompression: gzip",
		"metadata:",
		"format: json",
	}
	
	for _, expected := range expectedStrings {
		if !contains(content, expected) {
			t.Errorf("Expected config to contain '%s'", expected)
		}
	}
}

func TestWriterValidation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	writer := NewWriter(configPath)
	
	// Try to write invalid config
	cfg := &Config{
		Logger: LoggerConfig{
			Level:  "invalid",
			Output: "console",
			File:   FileLoggerConfig{Format: "text"},
		},
		Backup:   BackupConfig{DefaultCompression: "gzip"},
		Metadata: MetadataConfig{Format: "json"},
	}
	
	err := writer.Write(cfg)
	if err == nil {
		t.Error("Expected error writing invalid config")
	}
	
	// File should not exist
	if writer.Exists() {
		t.Error("Config file should not exist after failed write")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
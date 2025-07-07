package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected logger.Level
		wantErr  bool
	}{
		{"debug", logger.DebugLevel, false},
		{"DEBUG", logger.DebugLevel, false},
		{"info", logger.InfoLevel, false},
		{"INFO", logger.InfoLevel, false},
		{"warn", logger.WarnLevel, false},
		{"warning", logger.WarnLevel, false},
		{"error", logger.ErrorLevel, false},
		{"fatal", logger.FatalLevel, false},
		{"unknown", logger.InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := parseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if level != tt.expected {
				t.Errorf("parseLevel() = %v, want %v", level, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != "info" {
		t.Errorf("DefaultConfig().Level = %v, want info", cfg.Level)
	}
	if cfg.Output != "console" {
		t.Errorf("DefaultConfig().Output = %v, want console", cfg.Output)
	}
	if cfg.MaxSize != 100 {
		t.Errorf("DefaultConfig().MaxSize = %v, want 100", cfg.MaxSize)
	}
	if cfg.MaxAge != 7 {
		t.Errorf("DefaultConfig().MaxAge = %v, want 7", cfg.MaxAge)
	}
	if cfg.JSONFormat != false {
		t.Errorf("DefaultConfig().JSONFormat = %v, want false", cfg.JSONFormat)
	}
	if cfg.UseColor != true {
		t.Errorf("DefaultConfig().UseColor = %v, want true", cfg.UseColor)
	}
}

func TestNewLogger_Console(t *testing.T) {
	cfg := Config{
		Level:    "debug",
		Output:   "console",
		UseColor: false,
	}

	l, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	if l.GetLevel() != logger.DebugLevel {
		t.Errorf("Logger level = %v, want %v", l.GetLevel(), logger.DebugLevel)
	}
}

func TestNewLogger_File(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Output:     "file",
		FilePath:   logPath,
		MaxSize:    10,
		MaxAge:     7,
		JSONFormat: true,
	}

	l, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// Close file logger
	if fl, ok := l.(*FileLogger); ok {
		defer fl.Close()
	}

	// Check if log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestNewLogger_Both(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "warn",
		Output:     "both",
		FilePath:   logPath,
		MaxSize:    10,
		MaxAge:     7,
		JSONFormat: false,
		UseColor:   false,
	}

	l, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	if l.GetLevel() != logger.WarnLevel {
		t.Errorf("Logger level = %v, want %v", l.GetLevel(), logger.WarnLevel)
	}

	// Test logging
	l.Warn("test warning")

	// Check if log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestNewLogger_InvalidOutput(t *testing.T) {
	cfg := Config{
		Level:  "info",
		Output: "invalid",
	}

	_, err := NewLogger(cfg)
	if err == nil {
		t.Error("Expected error for invalid output type")
	}
}

func TestMultiLogger(t *testing.T) {
	// Create a multi logger with two console loggers for testing
	l1 := NewConsoleLogger(logger.DebugLevel, false)
	l2 := NewConsoleLogger(logger.InfoLevel, false)
	
	multi := NewMultiLogger(l1, l2)
	
	// Test initial level
	multi.SetLevel(logger.WarnLevel)
	if multi.GetLevel() != logger.WarnLevel {
		t.Errorf("MultiLogger level = %v, want %v", multi.GetLevel(), logger.WarnLevel)
	}
	
	// Test WithField
	l3 := multi.WithField("key", "value")
	if l3 == multi {
		t.Error("WithField should return a new logger")
	}
	
	// Test WithFields
	l4 := multi.WithFields(logger.F("key1", "value1"), logger.F("key2", "value2"))
	if l4 == multi {
		t.Error("WithFields should return a new logger")
	}
}
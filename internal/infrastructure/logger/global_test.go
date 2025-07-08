package logger

import (
	"os"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

func TestGlobalLogger(t *testing.T) {
	// Test default global logger
	l := GetGlobalLogger()
	if l == nil {
		t.Fatal("Global logger should not be nil")
	}

	// Test setting a new global logger
	newLogger := NewConsoleLogger(logger.DebugLevel, false)
	SetGlobalLogger(newLogger)

	l2 := GetGlobalLogger()
	if l2.GetLevel() != logger.DebugLevel {
		t.Errorf("Global logger level = %v, want %v", l2.GetLevel(), logger.DebugLevel)
	}
}

func TestGlobalLoggerFunctions(t *testing.T) {
	// Set a test logger
	testLogger := NewConsoleLogger(logger.DebugLevel, false)
	SetGlobalLogger(testLogger)

	// Test convenience functions (just ensure they don't panic)
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	// Test WithField
	l := WithField("key", "value")
	if l == nil {
		t.Error("WithField returned nil")
	}

	// Test WithFields
	l2 := WithFields(logger.F("key1", "value1"))
	if l2 == nil {
		t.Error("WithFields returned nil")
	}

	// Test level functions
	SetLevel(logger.ErrorLevel)
	if GetLevel() != logger.ErrorLevel {
		t.Errorf("GetLevel() = %v, want %v", GetLevel(), logger.ErrorLevel)
	}
}

func TestInitializeFromConfig(t *testing.T) {
	cfg := Config{
		Level:    "warn",
		Output:   "console",
		UseColor: false,
	}

	err := InitializeFromConfig(cfg)
	if err != nil {
		t.Fatalf("InitializeFromConfig() error = %v", err)
	}

	if GetLevel() != logger.WarnLevel {
		t.Errorf("Global logger level = %v, want %v", GetLevel(), logger.WarnLevel)
	}
}

func TestInitializeFromEnv(t *testing.T) {
	// Save original env vars
	origLevel := os.Getenv("CLI_RECOVER_LOG_LEVEL")
	origOutput := os.Getenv("CLI_RECOVER_LOG_OUTPUT")
	origColor := os.Getenv("CLI_RECOVER_LOG_COLOR")

	// Set test env vars
	os.Setenv("CLI_RECOVER_LOG_LEVEL", "error")
	os.Setenv("CLI_RECOVER_LOG_OUTPUT", "console")
	os.Setenv("CLI_RECOVER_LOG_COLOR", "false")

	// Restore env vars after test
	defer func() {
		os.Setenv("CLI_RECOVER_LOG_LEVEL", origLevel)
		os.Setenv("CLI_RECOVER_LOG_OUTPUT", origOutput)
		os.Setenv("CLI_RECOVER_LOG_COLOR", origColor)
	}()

	err := InitializeFromEnv()
	if err != nil {
		t.Fatalf("InitializeFromEnv() error = %v", err)
	}

	if GetLevel() != logger.ErrorLevel {
		t.Errorf("Global logger level = %v, want %v", GetLevel(), logger.ErrorLevel)
	}
}

package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

func TestFileLogger_Creation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := FileLoggerOptions{
		FilePath:   logPath,
		MaxSize:    10,
		MaxAge:     7,
		JSONFormat: false,
	}

	l, err := NewFileLogger(logger.InfoLevel, opts)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer l.Close()

	// Check if file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestFileLogger_LogLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := FileLoggerOptions{
		FilePath:   logPath,
		JSONFormat: false,
	}

	l, err := NewFileLogger(logger.DebugLevel, opts)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer l.Close()

	// Log messages at different levels
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logs := string(content)
	if !strings.Contains(logs, "DEBUG") || !strings.Contains(logs, "debug message") {
		t.Error("Debug log not found")
	}
	if !strings.Contains(logs, "INFO") || !strings.Contains(logs, "info message") {
		t.Error("Info log not found")
	}
	if !strings.Contains(logs, "WARN") || !strings.Contains(logs, "warn message") {
		t.Error("Warn log not found")
	}
	if !strings.Contains(logs, "ERROR") || !strings.Contains(logs, "error message") {
		t.Error("Error log not found")
	}
}

func TestFileLogger_LogLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := FileLoggerOptions{
		FilePath:   logPath,
		JSONFormat: false,
	}

	l, err := NewFileLogger(logger.InfoLevel, opts)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer l.Close()

	// Debug should not be logged
	l.Debug("should not appear")
	// Info should be logged
	l.Info("should appear")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logs := string(content)
	if strings.Contains(logs, "should not appear") {
		t.Error("Debug message was logged when level is Info")
	}
	if !strings.Contains(logs, "should appear") {
		t.Error("Info message was not logged when level is Info")
	}
}

func TestFileLogger_WithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := FileLoggerOptions{
		FilePath:   logPath,
		JSONFormat: false,
	}

	l, err := NewFileLogger(logger.InfoLevel, opts)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer l.Close()

	l2 := l.WithField("key1", "value1")
	l3 := l2.WithFields(logger.F("key2", "value2"), logger.F("key3", 123))

	l3.Info("test message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logs := string(content)
	if !strings.Contains(logs, "key1=value1") {
		t.Error("Field key1 not found in log")
	}
	if !strings.Contains(logs, "key2=value2") {
		t.Error("Field key2 not found in log")
	}
	if !strings.Contains(logs, "key3=123") {
		t.Error("Field key3 not found in log")
	}
}

func TestFileLogger_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := FileLoggerOptions{
		FilePath:   logPath,
		JSONFormat: true,
	}

	l, err := NewFileLogger(logger.InfoLevel, opts)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer l.Close()

	l.WithField("user", "test").Info("test message")

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logs := string(content)
	// Check JSON format
	if !strings.Contains(logs, `"level":"INFO"`) {
		t.Error("JSON format missing level field")
	}
	if !strings.Contains(logs, `"message":"test message"`) {
		t.Error("JSON format missing message field")
	}
	if !strings.Contains(logs, `"user":"test"`) {
		t.Error("JSON format missing user field")
	}
	if !strings.Contains(logs, `"timestamp"`) {
		t.Error("JSON format missing timestamp field")
	}
}

func TestFileLogger_Rotation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := FileLoggerOptions{
		FilePath:   logPath,
		MaxSize:    1, // 1MB, small for testing
		JSONFormat: false,
	}

	l, err := NewFileLogger(logger.InfoLevel, opts)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer l.Close()

	// Write logs until rotation happens
	largeMessage := strings.Repeat("x", 1024) // 1KB message
	for i := 0; i < 1100; i++ {               // Write more than 1MB
		l.Info(largeMessage)
	}

	// Give some time for rotation
	time.Sleep(100 * time.Millisecond)

	// Check if rotation happened
	files, err := filepath.Glob(filepath.Join(tmpDir, "test.log.*"))
	if err != nil {
		t.Fatalf("Failed to find rotated files: %v", err)
	}

	if len(files) == 0 {
		t.Error("No rotated files found")
	}

	// Check if current log file exists and is smaller than max size
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	if info.Size() > 1024*1024 {
		t.Error("Current log file is larger than max size after rotation")
	}
}

func TestFileLogger_SetLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := FileLoggerOptions{
		FilePath:   logPath,
		JSONFormat: false,
	}

	l, err := NewFileLogger(logger.DebugLevel, opts)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}
	defer l.Close()

	if l.GetLevel() != logger.DebugLevel {
		t.Errorf("Initial level = %v, want %v", l.GetLevel(), logger.DebugLevel)
	}

	l.SetLevel(logger.ErrorLevel)

	if l.GetLevel() != logger.ErrorLevel {
		t.Errorf("After SetLevel, level = %v, want %v", l.GetLevel(), logger.ErrorLevel)
	}
}

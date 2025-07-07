package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

func TestConsoleLogger_LogLevels(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	defer func() {
		os.Stderr = oldStderr
	}()

	l := NewConsoleLogger(logger.DebugLevel, false)

	// Test each log level
	tests := []struct {
		name     string
		logFunc  func(string, ...logger.Field)
		level    string
		message  string
	}{
		{
			name:    "debug",
			logFunc: l.Debug,
			level:   "DEBUG",
			message: "debug message",
		},
		{
			name:    "info",
			logFunc: l.Info,
			level:   "INFO",
			message: "info message",
		},
		{
			name:    "warn",
			logFunc: l.Warn,
			level:   "WARN",
			message: "warn message",
		},
		{
			name:    "error",
			logFunc: l.Error,
			level:   "ERROR",
			message: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			buf := new(bytes.Buffer)

			// Log message
			tt.logFunc(tt.message)

			// Close write side and read output
			w.Close()
			buf.ReadFrom(r)
			output := buf.String()

			// Reopen pipe for next test
			r, w, _ = os.Pipe()
			os.Stderr = w

			// Check output contains expected parts
			if !strings.Contains(output, tt.level) {
				t.Errorf("output missing level %s: %s", tt.level, output)
			}
			if !strings.Contains(output, tt.message) {
				t.Errorf("output missing message %s: %s", tt.message, output)
			}
		})
	}
}

func TestConsoleLogger_LogLevel(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	defer func() {
		os.Stderr = oldStderr
	}()

	l := NewConsoleLogger(logger.InfoLevel, false)

	// Debug should not be logged
	l.Debug("should not appear")

	// Info should be logged
	l.Info("should appear")

	// Close write side and read output
	w.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	if strings.Contains(output, "should not appear") {
		t.Error("Debug message was logged when level is Info")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("Info message was not logged when level is Info")
	}
}

func TestConsoleLogger_WithFields(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	defer func() {
		os.Stderr = oldStderr
	}()

	l := NewConsoleLogger(logger.InfoLevel, false)
	l2 := l.WithField("key1", "value1")
	l3 := l2.WithFields(logger.F("key2", "value2"), logger.F("key3", 123))

	l3.Info("test message")

	// Close write side and read output
	w.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Check all fields are present
	if !strings.Contains(output, "key1=value1") {
		t.Error("Field key1 not found in output")
	}
	if !strings.Contains(output, "key2=value2") {
		t.Error("Field key2 not found in output")
	}
	if !strings.Contains(output, "key3=123") {
		t.Error("Field key3 not found in output")
	}
}

func TestConsoleLogger_SetLevel(t *testing.T) {
	l := NewConsoleLogger(logger.DebugLevel, false)

	if l.GetLevel() != logger.DebugLevel {
		t.Errorf("Initial level = %v, want %v", l.GetLevel(), logger.DebugLevel)
	}

	l.SetLevel(logger.ErrorLevel)

	if l.GetLevel() != logger.ErrorLevel {
		t.Errorf("After SetLevel, level = %v, want %v", l.GetLevel(), logger.ErrorLevel)
	}
}

func TestConsoleLogger_Colors(t *testing.T) {
	tests := []struct {
		level     logger.Level
		wantColor string
	}{
		{logger.DebugLevel, "\033[36m"},
		{logger.InfoLevel, "\033[32m"},
		{logger.WarnLevel, "\033[33m"},
		{logger.ErrorLevel, "\033[31m"},
		{logger.FatalLevel, "\033[35m"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			color, reset := getColorForLevel(tt.level)
			if color != tt.wantColor {
				t.Errorf("getColorForLevel(%v) color = %v, want %v", tt.level, color, tt.wantColor)
			}
			if reset != "\033[0m" {
				t.Errorf("getColorForLevel(%v) reset = %v, want \\033[0m", tt.level, reset)
			}
		})
	}
}
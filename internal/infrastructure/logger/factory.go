package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

// Config represents logger configuration
type Config struct {
	Level      string // Log level: debug, info, warn, error, fatal
	Output     string // Output type: console, file, both
	FilePath   string // File path for file output
	MaxSize    int64  // Maximum size of log file in MB
	MaxAge     int    // Maximum age of log files in days
	JSONFormat bool   // Use JSON format for file output
	UseColor   bool   // Use color for console output
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		Level:      "info",
		Output:     "console",
		FilePath:   filepath.Join(homeDir, ".cli-recover", "logs", "cli-recover.log"),
		MaxSize:    100,
		MaxAge:     7,
		JSONFormat: false,
		UseColor:   true,
	}
}

// MultiLogger wraps multiple loggers and forwards calls to all of them
type MultiLogger struct {
	loggers []logger.Logger
	level   logger.Level
}

// NewMultiLogger creates a new multi logger
func NewMultiLogger(loggers ...logger.Logger) *MultiLogger {
	return &MultiLogger{
		loggers: loggers,
		level:   logger.InfoLevel,
	}
}

// Debug logs a message at debug level
func (m *MultiLogger) Debug(msg string, fields ...logger.Field) {
	for _, l := range m.loggers {
		l.Debug(msg, fields...)
	}
}

// Info logs a message at info level
func (m *MultiLogger) Info(msg string, fields ...logger.Field) {
	for _, l := range m.loggers {
		l.Info(msg, fields...)
	}
}

// Warn logs a message at warn level
func (m *MultiLogger) Warn(msg string, fields ...logger.Field) {
	for _, l := range m.loggers {
		l.Warn(msg, fields...)
	}
}

// Error logs a message at error level
func (m *MultiLogger) Error(msg string, fields ...logger.Field) {
	for _, l := range m.loggers {
		l.Error(msg, fields...)
	}
}

// Fatal logs a message at fatal level and exits the program
func (m *MultiLogger) Fatal(msg string, fields ...logger.Field) {
	for _, l := range m.loggers {
		l.Fatal(msg, fields...)
	}
}

// WithContext returns a logger with the given context
func (m *MultiLogger) WithContext(ctx context.Context) logger.Logger {
	newLoggers := make([]logger.Logger, len(m.loggers))
	for i, l := range m.loggers {
		newLoggers[i] = l.WithContext(ctx)
	}
	return NewMultiLogger(newLoggers...)
}

// WithField returns a logger with the given field
func (m *MultiLogger) WithField(key string, value interface{}) logger.Logger {
	newLoggers := make([]logger.Logger, len(m.loggers))
	for i, l := range m.loggers {
		newLoggers[i] = l.WithField(key, value)
	}
	return NewMultiLogger(newLoggers...)
}

// WithFields returns a logger with the given fields
func (m *MultiLogger) WithFields(fields ...logger.Field) logger.Logger {
	newLoggers := make([]logger.Logger, len(m.loggers))
	for i, l := range m.loggers {
		newLoggers[i] = l.WithFields(fields...)
	}
	return NewMultiLogger(newLoggers...)
}

// SetLevel sets the minimum logging level
func (m *MultiLogger) SetLevel(level logger.Level) {
	m.level = level
	for _, l := range m.loggers {
		l.SetLevel(level)
	}
}

// GetLevel returns the current logging level
func (m *MultiLogger) GetLevel() logger.Level {
	return m.level
}

// NewLogger creates a new logger based on configuration
func NewLogger(cfg Config) (logger.Logger, error) {
	// Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// Create loggers based on output type
	switch strings.ToLower(cfg.Output) {
	case "console":
		return NewConsoleLogger(level, cfg.UseColor), nil

	case "file":
		return NewFileLogger(level, FileLoggerOptions{
			FilePath:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			JSONFormat: cfg.JSONFormat,
		})

	case "both":
		consoleLogger := NewConsoleLogger(level, cfg.UseColor)
		fileLogger, err := NewFileLogger(level, FileLoggerOptions{
			FilePath:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			JSONFormat: cfg.JSONFormat,
		})
		if err != nil {
			return nil, err
		}
		multi := NewMultiLogger(consoleLogger, fileLogger)
		multi.SetLevel(level)
		return multi, nil

	default:
		return nil, fmt.Errorf("unknown output type: %s", cfg.Output)
	}
}

// parseLevel parses a string log level to logger.Level
func parseLevel(levelStr string) (logger.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return logger.DebugLevel, nil
	case "info":
		return logger.InfoLevel, nil
	case "warn", "warning":
		return logger.WarnLevel, nil
	case "error":
		return logger.ErrorLevel, nil
	case "fatal":
		return logger.FatalLevel, nil
	default:
		return logger.InfoLevel, fmt.Errorf("unknown log level: %s", levelStr)
	}
}
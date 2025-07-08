// Package logger defines the logging interface for the application
package logger

import (
	"context"
)

// Level represents the severity of a log message
type Level int

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production
	DebugLevel Level = iota
	// InfoLevel is the default logging priority
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual human review
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs
	ErrorLevel
	// FatalLevel logs a message, then calls os.Exit(1)
	FatalLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Logger is the interface for logging
type Logger interface {
	// Debug logs a message at debug level
	Debug(msg string, fields ...Field)
	// Info logs a message at info level
	Info(msg string, fields ...Field)
	// Warn logs a message at warn level
	Warn(msg string, fields ...Field)
	// Error logs a message at error level
	Error(msg string, fields ...Field)
	// Fatal logs a message at fatal level and exits the program
	Fatal(msg string, fields ...Field)

	// WithContext returns a logger with the given context
	WithContext(ctx context.Context) Logger
	// WithField returns a logger with the given field
	WithField(key string, value interface{}) Logger
	// WithFields returns a logger with the given fields
	WithFields(fields ...Field) Logger

	// SetLevel sets the minimum logging level
	SetLevel(level Level)
	// GetLevel returns the current logging level
	GetLevel() Level
}

// F is a convenience function for creating a Field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

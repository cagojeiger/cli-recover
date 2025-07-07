package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

// ConsoleLogger implements logger.Logger interface for console output
type ConsoleLogger struct {
	mu       sync.RWMutex
	level    logger.Level
	fields   []logger.Field
	useColor bool
}

// NewConsoleLogger creates a new console logger
func NewConsoleLogger(level logger.Level, useColor bool) *ConsoleLogger {
	return &ConsoleLogger{
		level:    level,
		fields:   []logger.Field{},
		useColor: useColor,
	}
}

// Debug logs a message at debug level
func (l *ConsoleLogger) Debug(msg string, fields ...logger.Field) {
	l.log(logger.DebugLevel, msg, fields...)
}

// Info logs a message at info level
func (l *ConsoleLogger) Info(msg string, fields ...logger.Field) {
	l.log(logger.InfoLevel, msg, fields...)
}

// Warn logs a message at warn level
func (l *ConsoleLogger) Warn(msg string, fields ...logger.Field) {
	l.log(logger.WarnLevel, msg, fields...)
}

// Error logs a message at error level
func (l *ConsoleLogger) Error(msg string, fields ...logger.Field) {
	l.log(logger.ErrorLevel, msg, fields...)
}

// Fatal logs a message at fatal level and exits the program
func (l *ConsoleLogger) Fatal(msg string, fields ...logger.Field) {
	l.log(logger.FatalLevel, msg, fields...)
	os.Exit(1)
}

// WithContext returns a logger with the given context
func (l *ConsoleLogger) WithContext(ctx context.Context) logger.Logger {
	// For now, we don't extract anything from context
	// This can be extended later to extract request IDs, trace IDs, etc.
	return l
}

// WithField returns a logger with the given field
func (l *ConsoleLogger) WithField(key string, value interface{}) logger.Logger {
	return l.WithFields(logger.F(key, value))
}

// WithFields returns a logger with the given fields
func (l *ConsoleLogger) WithFields(fields ...logger.Field) logger.Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &ConsoleLogger{
		level:    l.level,
		fields:   make([]logger.Field, len(l.fields)+len(fields)),
		useColor: l.useColor,
	}
	copy(newLogger.fields, l.fields)
	copy(newLogger.fields[len(l.fields):], fields)
	return newLogger
}

// SetLevel sets the minimum logging level
func (l *ConsoleLogger) SetLevel(level logger.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *ConsoleLogger) GetLevel() logger.Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// log is the internal logging method
func (l *ConsoleLogger) log(level logger.Level, msg string, fields ...logger.Field) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if level < l.level {
		return
	}

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Get color codes if enabled
	var levelColor, resetColor string
	if l.useColor {
		levelColor, resetColor = getColorForLevel(level)
	}

	// Format level
	levelStr := fmt.Sprintf("%-5s", level.String())

	// Build fields string
	allFields := append(l.fields, fields...)
	fieldsStr := ""
	if len(allFields) > 0 {
		fieldsStr = " "
		for i, f := range allFields {
			if i > 0 {
				fieldsStr += " "
			}
			fieldsStr += fmt.Sprintf("%s=%v", f.Key, f.Value)
		}
	}

	// Output format: timestamp [LEVEL] message fields
	fmt.Fprintf(os.Stderr, "%s %s[%s]%s %s%s\n",
		timestamp,
		levelColor,
		levelStr,
		resetColor,
		msg,
		fieldsStr,
	)
}

// getColorForLevel returns ANSI color codes for the given log level
func getColorForLevel(level logger.Level) (levelColor, resetColor string) {
	resetColor = "\033[0m"
	switch level {
	case logger.DebugLevel:
		levelColor = "\033[36m" // Cyan
	case logger.InfoLevel:
		levelColor = "\033[32m" // Green
	case logger.WarnLevel:
		levelColor = "\033[33m" // Yellow
	case logger.ErrorLevel:
		levelColor = "\033[31m" // Red
	case logger.FatalLevel:
		levelColor = "\033[35m" // Magenta
	default:
		levelColor = ""
	}
	return
}
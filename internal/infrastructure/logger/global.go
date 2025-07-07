package logger

import (
	"context"
	"os"
	"sync"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

var (
	globalLogger logger.Logger
	globalMutex  sync.RWMutex
)

func init() {
	// Initialize with default console logger
	globalLogger = NewConsoleLogger(logger.InfoLevel, true)
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(l logger.Logger) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalLogger = l
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() logger.Logger {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalLogger
}

// Debug logs a message at debug level using the global logger
func Debug(msg string, fields ...logger.Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

// Info logs a message at info level using the global logger
func Info(msg string, fields ...logger.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

// Warn logs a message at warn level using the global logger
func Warn(msg string, fields ...logger.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

// Error logs a message at error level using the global logger
func Error(msg string, fields ...logger.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

// Fatal logs a message at fatal level using the global logger and exits
func Fatal(msg string, fields ...logger.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

// WithContext returns a logger with the given context using the global logger
func WithContext(ctx context.Context) logger.Logger {
	return GetGlobalLogger().WithContext(ctx)
}

// WithField returns a logger with the given field using the global logger
func WithField(key string, value interface{}) logger.Logger {
	return GetGlobalLogger().WithField(key, value)
}

// WithFields returns a logger with the given fields using the global logger
func WithFields(fields ...logger.Field) logger.Logger {
	return GetGlobalLogger().WithFields(fields...)
}

// SetLevel sets the minimum logging level for the global logger
func SetLevel(level logger.Level) {
	GetGlobalLogger().SetLevel(level)
}

// GetLevel returns the current logging level of the global logger
func GetLevel() logger.Level {
	return GetGlobalLogger().GetLevel()
}

// InitializeFromConfig initializes the global logger from configuration
func InitializeFromConfig(cfg Config) error {
	l, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	SetGlobalLogger(l)
	return nil
}

// InitializeFromEnv initializes the global logger from environment variables
func InitializeFromEnv() error {
	cfg := DefaultConfig()
	
	// Override from environment variables
	if level := os.Getenv("CLI_RECOVER_LOG_LEVEL"); level != "" {
		cfg.Level = level
	}
	if output := os.Getenv("CLI_RECOVER_LOG_OUTPUT"); output != "" {
		cfg.Output = output
	}
	if filePath := os.Getenv("CLI_RECOVER_LOG_FILE"); filePath != "" {
		cfg.FilePath = filePath
	}
	if format := os.Getenv("CLI_RECOVER_LOG_FORMAT"); format == "json" {
		cfg.JSONFormat = true
	}
	if color := os.Getenv("CLI_RECOVER_LOG_COLOR"); color == "false" {
		cfg.UseColor = false
	}
	
	return InitializeFromConfig(cfg)
}
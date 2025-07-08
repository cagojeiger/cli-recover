package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
)

// FileLogger implements logger.Logger interface for file output
type FileLogger struct {
	mu           sync.RWMutex
	level        logger.Level
	fields       []logger.Field
	file         *os.File
	filePath     string
	maxSize      int64 // Maximum size of log file in bytes
	maxAge       int   // Maximum age of log files in days
	jsonFormat   bool
	rotationLock sync.Mutex
}

// FileLoggerOptions contains options for creating a file logger
type FileLoggerOptions struct {
	FilePath   string
	MaxSize    int64 // Maximum size of log file in MB (default: 100MB)
	MaxAge     int   // Maximum age of log files in days (default: 7)
	JSONFormat bool  // Use JSON format instead of text
}

// NewFileLogger creates a new file logger
func NewFileLogger(level logger.Level, opts FileLoggerOptions) (*FileLogger, error) {
	// Set defaults
	if opts.MaxSize == 0 {
		opts.MaxSize = 100 // 100MB
	}
	if opts.MaxAge == 0 {
		opts.MaxAge = 7 // 7 days
	}

	// Ensure directory exists
	dir := filepath.Dir(opts.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(opts.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	l := &FileLogger{
		level:      level,
		fields:     []logger.Field{},
		file:       file,
		filePath:   opts.FilePath,
		maxSize:    opts.MaxSize * 1024 * 1024, // Convert MB to bytes
		maxAge:     opts.MaxAge,
		jsonFormat: opts.JSONFormat,
	}

	// Start rotation cleanup goroutine
	go l.cleanupOldLogs()

	return l, nil
}

// Close closes the log file
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Debug logs a message at debug level
func (l *FileLogger) Debug(msg string, fields ...logger.Field) {
	l.log(logger.DebugLevel, msg, fields...)
}

// Info logs a message at info level
func (l *FileLogger) Info(msg string, fields ...logger.Field) {
	l.log(logger.InfoLevel, msg, fields...)
}

// Warn logs a message at warn level
func (l *FileLogger) Warn(msg string, fields ...logger.Field) {
	l.log(logger.WarnLevel, msg, fields...)
}

// Error logs a message at error level
func (l *FileLogger) Error(msg string, fields ...logger.Field) {
	l.log(logger.ErrorLevel, msg, fields...)
}

// Fatal logs a message at fatal level and exits the program
func (l *FileLogger) Fatal(msg string, fields ...logger.Field) {
	l.log(logger.FatalLevel, msg, fields...)
	os.Exit(1)
}

// WithContext returns a logger with the given context
func (l *FileLogger) WithContext(ctx context.Context) logger.Logger {
	return l
}

// WithField returns a logger with the given field
func (l *FileLogger) WithField(key string, value interface{}) logger.Logger {
	return l.WithFields(logger.F(key, value))
}

// WithFields returns a logger with the given fields
func (l *FileLogger) WithFields(fields ...logger.Field) logger.Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newLogger := &FileLogger{
		level:      l.level,
		fields:     make([]logger.Field, len(l.fields)+len(fields)),
		file:       l.file,
		filePath:   l.filePath,
		maxSize:    l.maxSize,
		maxAge:     l.maxAge,
		jsonFormat: l.jsonFormat,
	}
	copy(newLogger.fields, l.fields)
	copy(newLogger.fields[len(l.fields):], fields)
	return newLogger
}

// SetLevel sets the minimum logging level
func (l *FileLogger) SetLevel(level logger.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *FileLogger) GetLevel() logger.Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// log is the internal logging method
func (l *FileLogger) log(level logger.Level, msg string, fields ...logger.Field) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	l.mu.RUnlock()

	// Check if rotation is needed
	if err := l.rotateIfNeeded(); err != nil {
		// If rotation fails, try to continue logging
		fmt.Fprintf(os.Stderr, "Failed to rotate log: %v\n", err)
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.file == nil {
		return
	}

	// Create log entry
	entry := l.createLogEntry(level, msg, fields...)

	// Write to file
	if _, err := l.file.Write(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
	}
}

// createLogEntry creates a formatted log entry
func (l *FileLogger) createLogEntry(level logger.Level, msg string, fields ...logger.Field) []byte {
	timestamp := time.Now()
	allFields := append(l.fields, fields...)

	if l.jsonFormat {
		// JSON format
		entry := map[string]interface{}{
			"timestamp": timestamp.Format(time.RFC3339),
			"level":     level.String(),
			"message":   msg,
		}

		// Add fields
		for _, f := range allFields {
			entry[f.Key] = f.Value
		}

		data, _ := json.Marshal(entry)
		return append(data, '\n')
	}

	// Text format
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

	return []byte(fmt.Sprintf("%s [%-5s] %s%s\n",
		timestamp.Format("2006-01-02 15:04:05.000"),
		level.String(),
		msg,
		fieldsStr,
	))
}

// rotateIfNeeded checks if log rotation is needed and performs it
func (l *FileLogger) rotateIfNeeded() error {
	l.rotationLock.Lock()
	defer l.rotationLock.Unlock()

	l.mu.RLock()
	filePath := l.filePath
	l.mu.RUnlock()

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Check if rotation is needed
	if info.Size() < l.maxSize {
		return nil
	}

	// Close current file
	l.mu.Lock()
	if err := l.file.Close(); err != nil {
		l.mu.Unlock()
		return err
	}

	// Rotate file
	rotatedPath := fmt.Sprintf("%s.%s", filePath, time.Now().Format("20060102-150405"))
	if err := os.Rename(filePath, rotatedPath); err != nil {
		l.mu.Unlock()
		return err
	}

	// Open new file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		l.mu.Unlock()
		return err
	}

	l.file = file
	l.mu.Unlock()

	return nil
}

// cleanupOldLogs periodically removes old log files
func (l *FileLogger) cleanupOldLogs() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.RLock()
		dir := filepath.Dir(l.filePath)
		baseName := filepath.Base(l.filePath)
		maxAge := l.maxAge
		l.mu.RUnlock()

		// Find and remove old log files
		cutoff := time.Now().AddDate(0, 0, -maxAge)

		files, err := filepath.Glob(filepath.Join(dir, baseName+".*"))
		if err != nil {
			continue
		}

		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				os.Remove(file)
			}
		}
	}
}

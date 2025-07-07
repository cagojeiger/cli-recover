package log

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Type represents the type of log
type Type string

const (
	TypeBackup  Type = "backup"
	TypeRestore Type = "restore"
)

// Status represents the status of the operation
type Status string

const (
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Log represents a log entry for an operation
type Log struct {
	ID        string    // Unique identifier (timestamp-based)
	Type      Type      // backup or restore
	Provider  string    // filesystem, minio, mongodb
	Status    Status    // running, completed, failed
	StartTime time.Time
	EndTime   *time.Time
	FilePath  string // Path to the log file
	Metadata  map[string]string
}

// NewLog creates a new log entry
func NewLog(logType Type, provider string) (*Log, error) {
	if logType != TypeBackup && logType != TypeRestore {
		return nil, fmt.Errorf("invalid log type: %s", logType)
	}

	if provider == "" {
		return nil, fmt.Errorf("provider cannot be empty")
	}

	now := time.Now()
	id := now.Format("20060102-150405.000000")

	return &Log{
		ID:        id,
		Type:      logType,
		Provider:  provider,
		Status:    StatusRunning,
		StartTime: now,
		Metadata:  make(map[string]string),
	}, nil
}

// Complete marks the log as completed
func (l *Log) Complete() {
	now := time.Now()
	l.Status = StatusCompleted
	l.EndTime = &now
}

// Fail marks the log as failed
func (l *Log) Fail(reason string) {
	now := time.Now()
	l.Status = StatusFailed
	l.EndTime = &now
	l.Metadata["error"] = reason
}

// Duration returns the duration of the operation
func (l *Log) Duration() time.Duration {
	if l.EndTime != nil {
		return l.EndTime.Sub(l.StartTime)
	}
	return time.Since(l.StartTime)
}

// Filename returns the suggested filename for this log
func (l *Log) Filename() string {
	// Format: backup-filesystem-20240107-150405.log
	return fmt.Sprintf("%s-%s-%s.log", l.Type, l.Provider, l.ID)
}

// GenerateLogPath generates the full path for the log file
func (l *Log) GenerateLogPath(baseDir string) string {
	// Organize by type/provider/date
	date := l.StartTime.Format("2006-01-02")
	return filepath.Join(baseDir, string(l.Type), l.Provider, date, l.Filename())
}

// IsRunning checks if the log is still running
func (l *Log) IsRunning() bool {
	return l.Status == StatusRunning
}

// SetMetadata sets a metadata key-value pair
func (l *Log) SetMetadata(key, value string) {
	if l.Metadata == nil {
		l.Metadata = make(map[string]string)
	}
	l.Metadata[key] = value
}

// GetMetadata gets a metadata value
func (l *Log) GetMetadata(key string) (string, bool) {
	if l.Metadata == nil {
		return "", false
	}
	val, ok := l.Metadata[key]
	return val, ok
}

// Validate validates the log entry
func (l *Log) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("log ID cannot be empty")
	}
	if l.Type == "" {
		return fmt.Errorf("log type cannot be empty")
	}
	if l.Provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if strings.Contains(l.Provider, "/") || strings.Contains(l.Provider, "..") {
		return fmt.Errorf("invalid provider name: %s", l.Provider)
	}
	return nil
}
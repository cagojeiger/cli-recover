package backup

import (
	"errors"
	"fmt"
	"time"
)

// Progress represents the current state of a backup operation
type Progress struct {
	Current int64  // Current bytes processed
	Total   int64  // Total bytes to process
	Message string // Status message
}

// CalculateSpeed calculates the transfer speed in bytes per second
func (p Progress) CalculateSpeed(duration time.Duration) float64 {
	if duration == 0 {
		return 0
	}
	return float64(p.Current) / duration.Seconds()
}

// CalculateETA calculates the estimated time remaining
func (p Progress) CalculateETA(speed float64) time.Duration {
	if speed == 0 || p.Current >= p.Total {
		return 0
	}

	remaining := p.Total - p.Current
	seconds := float64(remaining) / speed
	return time.Duration(seconds * float64(time.Second))
}

// Percentage returns the completion percentage
func (p Progress) Percentage() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Current) / float64(p.Total) * 100
}

// FormatETA formats a duration as a human-readable ETA string
func FormatETA(eta time.Duration) string {
	if eta == 0 {
		return "0s"
	}

	hours := int(eta.Hours())
	minutes := int(eta.Minutes()) % 60
	seconds := int(eta.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// Options represents backup operation options
type Options struct {
	Namespace  string
	PodName    string
	SourcePath string
	OutputFile string
	Compress   bool
	Exclude    []string
	Container  string                 // Optional: specific container in pod
	Extra      map[string]interface{} // Provider-specific options
}

// Validate checks if all required options are provided
func (o Options) Validate() error {
	if o.Namespace == "" {
		return errors.New("namespace is required")
	}
	if o.PodName == "" {
		return errors.New("pod name is required")
	}
	if o.SourcePath == "" {
		return errors.New("source path is required")
	}
	if o.OutputFile == "" {
		return errors.New("output file is required")
	}
	return nil
}

// ErrorCode represents the type of error
type ErrorCode string

const (
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
	ErrCodeInternal     ErrorCode = "INTERNAL"
	ErrCodeTimeout      ErrorCode = "TIMEOUT"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
)

// Result represents the result of a backup operation
type Result struct {
	BackupFile  string    // Path to the created backup file
	Size        int64     // Size of the backup in bytes
	Checksum    string    // SHA256 checksum of the backup file
	StartTime   time.Time // When the backup started
	EndTime     time.Time // When the backup completed
	FileCount   int       // Number of files backed up (if available)
}

// BackupError represents a domain-specific error
type BackupError struct {
	Code      ErrorCode
	Message   string
	Cause     error
	Timestamp time.Time
}

// NewBackupError creates a new BackupError wrapping the given error
func NewBackupError(message string, cause error) *BackupError {
	return &BackupError{
		Code:      ErrCodeInternal,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// Error implements the error interface
func (e BackupError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Is checks if the error has the same code
func (e BackupError) Is(target error) bool {
	if t, ok := target.(BackupError); ok {
		return e.Code == t.Code
	}
	return false
}

// Unwrap returns the underlying error
func (e BackupError) Unwrap() error {
	return e.Cause
}

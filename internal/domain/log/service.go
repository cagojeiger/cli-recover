package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Service provides high-level log management
type Service struct {
	repository Repository
	logDir     string
}

// NewService creates a new log service
func NewService(repository Repository, logDir string) *Service {
	return &Service{
		repository: repository,
		logDir:     logDir,
	}
}

// StartLog creates a new log entry and opens a log file for writing
func (s *Service) StartLog(logType Type, provider string, metadata map[string]string) (*Log, *Writer, error) {
	// Create log entry
	log, err := NewLog(logType, provider)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log entry: %w", err)
	}

	// Add metadata
	for k, v := range metadata {
		log.SetMetadata(k, v)
	}

	// Generate log file path
	log.FilePath = log.GenerateLogPath(s.logDir)

	// Create log writer
	writer, err := NewWriter(log.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log writer: %w", err)
	}

	// Save log entry
	if err := s.repository.Save(log); err != nil {
		writer.Close()
		return nil, nil, fmt.Errorf("failed to save log entry: %w", err)
	}

	// Write initial info
	writer.WriteLine("Operation: %s %s", logType, provider)
	for k, v := range metadata {
		writer.WriteLine("%s: %s", k, v)
	}
	writer.WriteLine("---")

	return log, writer, nil
}

// CompleteLog marks a log as completed
func (s *Service) CompleteLog(logID string) error {
	log, err := s.repository.Get(logID)
	if err != nil {
		return fmt.Errorf("failed to get log: %w", err)
	}

	log.Complete()
	
	if err := s.repository.Update(log); err != nil {
		return fmt.Errorf("failed to update log: %w", err)
	}

	return nil
}

// FailLog marks a log as failed
func (s *Service) FailLog(logID string, reason string) error {
	log, err := s.repository.Get(logID)
	if err != nil {
		return fmt.Errorf("failed to get log: %w", err)
	}

	log.Fail(reason)
	
	if err := s.repository.Update(log); err != nil {
		return fmt.Errorf("failed to update log: %w", err)
	}

	return nil
}

// GetLog retrieves a log entry
func (s *Service) GetLog(logID string) (*Log, error) {
	return s.repository.Get(logID)
}

// ListLogs lists logs with filters
func (s *Service) ListLogs(filter ListFilter) ([]*Log, error) {
	return s.repository.List(filter)
}

// ReadLogFile reads the content of a log file
func (s *Service) ReadLogFile(logID string) ([]byte, error) {
	log, err := s.repository.Get(logID)
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	if log.FilePath == "" {
		return nil, fmt.Errorf("log file path not set")
	}

	content, err := os.ReadFile(log.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	return content, nil
}

// TailLogFile returns a reader for tailing a log file
func (s *Service) TailLogFile(logID string) (io.ReadCloser, error) {
	log, err := s.repository.Get(logID)
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	if log.FilePath == "" {
		return nil, fmt.Errorf("log file path not set")
	}

	file, err := os.Open(log.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}

// CleanupOldLogs removes logs older than the specified duration
func (s *Service) CleanupOldLogs(maxAge time.Duration) error {
	// Get old logs
	cutoff := time.Now().Add(-maxAge)
	filter := ListFilter{
		EndDate: &cutoff,
	}
	
	oldLogs, err := s.repository.List(filter)
	if err != nil {
		return fmt.Errorf("failed to list old logs: %w", err)
	}

	// Delete each old log
	for _, log := range oldLogs {
		// Remove log file
		if log.FilePath != "" {
			os.Remove(log.FilePath)
		}
		
		// Remove metadata
		if err := s.repository.Delete(log.ID); err != nil {
			// Continue on error
			continue
		}
	}

	return nil
}

// GetLogFilePath returns the full path for a log file
func (s *Service) GetLogFilePath(log *Log) string {
	if log.FilePath != "" {
		return log.FilePath
	}
	return filepath.Join(s.logDir, log.GenerateLogPath(s.logDir))
}
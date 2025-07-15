package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cagojeiger/cli-pipe/internal/utils"
)

// LogCleaner handles cleanup of old log files
type LogCleaner struct {
	logger Logger
}

// NewLogCleaner creates a new log cleaner
func NewLogCleaner(logger Logger) *LogCleaner {
	if logger == nil {
		logger = Default()
	}
	return &LogCleaner{
		logger: logger,
	}
}

// CleanOldLogs removes log directories older than retentionDays
func (c *LogCleaner) CleanOldLogs(logDir string, retentionDays int) error {
	if retentionDays <= 0 {
		c.logger.Debug("log retention disabled", "retention_days", retentionDays)
		return nil
	}

	c.logger.Info("cleaning old logs", "directory", logDir, "retention_days", retentionDays)

	// Calculate cutoff time
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	
	// List all entries in log directory
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			c.logger.Debug("log directory does not exist", "path", logDir)
			return nil
		}
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	var removed int
	var totalSize int64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip files, only process directories
		}

		dirPath := filepath.Join(logDir, entry.Name())
		
		// Get directory info
		info, err := entry.Info()
		if err != nil {
			c.logger.Warn("failed to get directory info", "path", dirPath, "error", err)
			continue
		}

		// Check if directory is old enough to remove
		if info.ModTime().Before(cutoff) {
			// Calculate directory size before removal
			size, _ := c.calculateDirSize(dirPath)
			
			c.logger.Debug("removing old log directory",
				"path", dirPath,
				"age_days", int(time.Since(info.ModTime()).Hours()/24),
				"size", size)

			if err := os.RemoveAll(dirPath); err != nil {
				c.logger.Error("failed to remove log directory", "path", dirPath, "error", err)
				continue
			}

			removed++
			totalSize += size
		}
	}

	if removed > 0 {
		c.logger.Info("cleaned old logs",
			"removed_directories", removed,
			"reclaimed_bytes", totalSize,
			"reclaimed_human", utils.FormatBytes(totalSize))
	} else {
		c.logger.Debug("no old logs to clean")
	}

	return nil
}

// CleanOldLogFiles removes individual log files older than retentionDays
func (c *LogCleaner) CleanOldLogFiles(logDir string, pattern string, retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	
	// Find files matching pattern
	matches, err := filepath.Glob(filepath.Join(logDir, pattern))
	if err != nil {
		return fmt.Errorf("failed to list log files: %w", err)
	}

	var removed int
	var totalSize int64

	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			size := info.Size()
			
			c.logger.Debug("removing old log file",
				"path", path,
				"age_days", int(time.Since(info.ModTime()).Hours()/24),
				"size", size)

			if err := os.Remove(path); err != nil {
				c.logger.Error("failed to remove log file", "path", path, "error", err)
				continue
			}

			removed++
			totalSize += size
		}
	}

	if removed > 0 {
		c.logger.Info("cleaned old log files",
			"pattern", pattern,
			"removed_files", removed,
			"reclaimed_bytes", totalSize)
	}

	return nil
}

// calculateDirSize calculates the total size of a directory
func (c *LogCleaner) calculateDirSize(path string) (int64, error) {
	var size int64
	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}


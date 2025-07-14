package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// RotatingFileWriter implements io.WriteCloser with rotation support
type RotatingFileWriter struct {
	mu sync.Mutex

	filename   string
	maxSize    int64 // Maximum size in bytes
	maxBackups int   // Maximum number of old log files to retain
	maxAge     int   // Maximum age in days to retain old log files

	file *os.File
	size int64
}

// NewRotatingFileWriter creates a new RotatingFileWriter
func NewRotatingFileWriter(filename string, maxSizeMB, maxBackups, maxAge int) *RotatingFileWriter {
	return &RotatingFileWriter{
		filename:   filename,
		maxSize:    int64(maxSizeMB) * 1024 * 1024,
		maxBackups: maxBackups,
		maxAge:     maxAge,
	}
}

// Write implements io.Writer
func (w *RotatingFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err = w.openExistingOrNew(); err != nil {
			return 0, err
		}
	}

	if w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	w.size += int64(n)

	return n, err
}

// Close implements io.Closer
func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	err := w.file.Close()
	w.file = nil
	return err
}

func (w *RotatingFileWriter) openExistingOrNew() error {
	// Ensure directory exists
	dir := filepath.Dir(w.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	info, err := os.Stat(w.filename)
	if os.IsNotExist(err) {
		return w.openNew()
	}
	if err != nil {
		return err
	}

	file, err := os.OpenFile(w.filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = file
	w.size = info.Size()
	return nil
}

func (w *RotatingFileWriter) openNew() error {
	// Ensure directory exists
	dir := filepath.Dir(w.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	w.file = file
	w.size = 0
	return nil
}

func (w *RotatingFileWriter) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}

	// Rename current log file
	newName := w.backupName()
	if err := os.Rename(w.filename, newName); err != nil {
		return err
	}

	// Compress the rotated file
	if err := w.compressFile(newName); err != nil {
		// Log error but don't fail rotation
		fmt.Fprintf(os.Stderr, "Failed to compress rotated log: %v\n", err)
	}

	// Clean up old files
	if err := w.deleteOldFiles(); err != nil {
		// Log error but don't fail rotation
		fmt.Fprintf(os.Stderr, "Failed to delete old logs: %v\n", err)
	}

	return w.openNew()
}

func (w *RotatingFileWriter) backupName() string {
	dir := filepath.Dir(w.filename)
	filename := filepath.Base(w.filename)
	ext := filepath.Ext(filename)
	prefix := strings.TrimSuffix(filename, ext)

	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

func (w *RotatingFileWriter) compressFile(filename string) error {
	src, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filename + ".gz")
	if err != nil {
		return err
	}
	defer dst.Close()

	gz := gzip.NewWriter(dst)
	defer gz.Close()

	_, err = io.Copy(gz, src)
	if err != nil {
		return err
	}

	// Remove original file after successful compression
	return os.Remove(filename)
}

func (w *RotatingFileWriter) deleteOldFiles() error {
	dir := filepath.Dir(w.filename)
	filename := filepath.Base(w.filename)
	ext := filepath.Ext(filename)
	prefix := strings.TrimSuffix(filename, ext)

	// Find all backup files
	pattern := filepath.Join(dir, prefix+"-*"+ext+"*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// Sort by modification time
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var files []fileInfo

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			path:    match,
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	// Delete files beyond maxBackups
	if w.maxBackups > 0 && len(files) > w.maxBackups {
		for i := w.maxBackups; i < len(files); i++ {
			os.Remove(files[i].path)
		}
	}

	// Delete files older than maxAge
	if w.maxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -w.maxAge)
		for _, file := range files {
			if file.modTime.Before(cutoff) {
				os.Remove(file.path)
			}
		}
	}

	return nil
}
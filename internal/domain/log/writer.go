package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Writer provides thread-safe writing to a log file
type Writer struct {
	mu       sync.Mutex
	file     *os.File
	filePath string
	isClosed bool
}

// NewWriter creates a new log writer
func NewWriter(filePath string) (*Writer, error) {
	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open file for writing (create if not exists, append if exists)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Write header
	header := fmt.Sprintf("\n=== Log started at %s ===\n", time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	return &Writer{
		file:     file,
		filePath: filePath,
	}, nil
}

// Write writes data to the log file
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isClosed {
		return 0, fmt.Errorf("log writer is closed")
	}

	return w.file.Write(p)
}

// WriteString writes a string to the log file
func (w *Writer) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

// WriteLine writes a line to the log file with timestamp
func (w *Writer) WriteLine(format string, args ...interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isClosed {
		return fmt.Errorf("log writer is closed")
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
	
	_, err := w.file.WriteString(line)
	return err
}

// Flush flushes any buffered data to disk
func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isClosed {
		return fmt.Errorf("log writer is closed")
	}

	return w.file.Sync()
}

// Close closes the log file
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isClosed {
		return nil
	}

	// Write footer
	footer := fmt.Sprintf("=== Log ended at %s ===\n", time.Now().Format(time.RFC3339))
	w.file.WriteString(footer)

	w.isClosed = true
	return w.file.Close()
}

// GetPath returns the log file path
func (w *Writer) GetPath() string {
	return w.filePath
}

// MultiWriter creates a writer that writes to multiple destinations
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter creates a new multi-writer
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Write writes to all writers
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// TeeWriter creates a writer that writes to both a log file and stdout/stderr
func TeeWriter(logWriter *Writer, output io.Writer) io.Writer {
	return NewMultiWriter(logWriter, output)
}
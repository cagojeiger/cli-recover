package logger

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotatingFileWriter_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	w := NewRotatingFileWriter(logFile, 1, 3, 7) // 1MB, 3 backups, 7 days
	defer w.Close()
	
	// Write some data
	data := []byte("test log line\n")
	n, err := w.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d, want %d", n, len(data))
	}
	
	// Verify file exists
	if _, err := os.Stat(logFile); err != nil {
		t.Errorf("Log file not created: %v", err)
	}
}

func TestRotatingFileWriter_Rotation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	// Small max size to trigger rotation
	w := NewRotatingFileWriter(logFile, 0, 3, 7) // 0MB (1KB), 3 backups, 7 days
	w.maxSize = 1024 // Override to 1KB for testing
	defer w.Close()
	
	// Write data to trigger rotation
	largeData := make([]byte, 1100) // More than 1KB
	for i := range largeData {
		largeData[i] = 'A'
	}
	
	_, err := w.Write(largeData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	
	// Check that rotation happened
	files, err := filepath.Glob(filepath.Join(tmpDir, "test-*.log*"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	
	if len(files) == 0 {
		t.Error("Expected rotated files, found none")
	}
}

func TestRotatingFileWriter_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	// Create existing file
	existingData := []byte("existing content\n")
	if err := os.WriteFile(logFile, existingData, 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}
	
	w := NewRotatingFileWriter(logFile, 1, 3, 7)
	defer w.Close()
	
	// Write should append
	newData := []byte("new content\n")
	_, err := w.Write(newData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	
	// Verify content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	
	expected := string(existingData) + string(newData)
	if string(content) != expected {
		t.Errorf("File content = %q, want %q", string(content), expected)
	}
}

func TestRotatingFileWriter_Compression(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test compression by creating a file and compressing it
	testData := []byte("test data for compression\n")
	testFile := filepath.Join(tmpDir, "compress-test.log")
	
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	w := &RotatingFileWriter{}
	if err := w.compressFile(testFile); err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	
	// Verify compressed file exists
	gzFile := testFile + ".gz"
	if _, err := os.Stat(gzFile); err != nil {
		t.Errorf("Compressed file not created: %v", err)
	}
	
	// Verify original file was removed
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Original file should have been removed")
	}
	
	// Verify compressed content
	f, err := os.Open(gzFile)
	if err != nil {
		t.Fatalf("Failed to open compressed file: %v", err)
	}
	defer f.Close()
	
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gz.Close()
	
	content, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("Failed to read compressed content: %v", err)
	}
	
	if string(content) != string(testData) {
		t.Errorf("Decompressed content = %q, want %q", string(content), string(testData))
	}
}

func TestRotatingFileWriter_DeleteOldFiles(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	w := NewRotatingFileWriter(logFile, 1, 2, 1) // Keep 2 files, 1 day
	
	// Create some old backup files
	now := time.Now()
	oldFiles := []struct {
		name string
		age  time.Duration
	}{
		{"test-20240101-120000.log", 48 * time.Hour}, // 2 days old
		{"test-20240102-120000.log", 36 * time.Hour}, // 1.5 days old
		{"test-20240103-120000.log", 12 * time.Hour}, // 0.5 days old
		{"test-20240104-120000.log", 6 * time.Hour},  // 0.25 days old
	}
	
	for _, f := range oldFiles {
		path := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		// Set modification time
		modTime := now.Add(-f.age)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("Failed to set file time: %v", err)
		}
	}
	
	// Delete old files
	if err := w.deleteOldFiles(); err != nil {
		t.Fatalf("deleteOldFiles failed: %v", err)
	}
	
	// Check remaining files
	files, err := filepath.Glob(filepath.Join(tmpDir, "test-*.log"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	
	// Should keep only 2 newest files that are less than 1 day old
	if len(files) > 2 {
		t.Errorf("Expected at most 2 files, found %d", len(files))
	}
}

func TestRotatingFileWriter_FormatHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	w := NewRotatingFileWriter(logFile, 1, 3, 7)
	
	// Test backup name generation
	backup := w.backupName()
	if !filepath.IsAbs(backup) {
		t.Error("Backup name should be absolute path")
	}
	
	dir := filepath.Dir(backup)
	if dir != tmpDir {
		t.Errorf("Backup directory = %q, want %q", dir, tmpDir)
	}
	
	// Should contain timestamp pattern
	base := filepath.Base(backup)
	if len(base) < len("test-20060102-150405.log") {
		t.Error("Backup name too short, missing timestamp")
	}
}
package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogCleaner_CleanOldLogs(t *testing.T) {
	tmpDir := t.TempDir()
	cleaner := NewLogCleaner(nil)
	
	// Create test directories with different ages
	testDirs := []struct {
		name string
		age  time.Duration
		keep bool
	}{
		{"pipeline1_20240101_120000", 10 * 24 * time.Hour, false}, // 10 days old
		{"pipeline2_20240105_120000", 5 * 24 * time.Hour, false},  // 5 days old  
		{"pipeline3_20240109_120000", 1 * 24 * time.Hour, true},   // 1 day old
		{"pipeline4_20240110_120000", 6 * time.Hour, true},        // 6 hours old
	}
	
	for _, td := range testDirs {
		dirPath := filepath.Join(tmpDir, td.name)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		
		// Create some files in the directory
		testFile := filepath.Join(dirPath, "pipeline.log")
		if err := os.WriteFile(testFile, []byte("test log content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Set modification time
		modTime := time.Now().Add(-td.age)
		if err := os.Chtimes(dirPath, modTime, modTime); err != nil {
			t.Fatalf("Failed to set directory time: %v", err)
		}
	}
	
	// Clean logs older than 3 days
	err := cleaner.CleanOldLogs(tmpDir, 3)
	if err != nil {
		t.Fatalf("CleanOldLogs failed: %v", err)
	}
	
	// Verify results
	for _, td := range testDirs {
		dirPath := filepath.Join(tmpDir, td.name)
		_, err := os.Stat(dirPath)
		
		if td.keep {
			if os.IsNotExist(err) {
				t.Errorf("Directory %s should have been kept but was removed", td.name)
			}
		} else {
			if !os.IsNotExist(err) {
				t.Errorf("Directory %s should have been removed but still exists", td.name)
			}
		}
	}
}

func TestLogCleaner_CleanOldLogs_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	cleaner := NewLogCleaner(nil)
	
	// Create a test directory
	testDir := filepath.Join(tmpDir, "test_dir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Clean with retention disabled (0 days)
	err := cleaner.CleanOldLogs(tmpDir, 0)
	if err != nil {
		t.Fatalf("CleanOldLogs failed: %v", err)
	}
	
	// Directory should still exist
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Directory should not have been removed when retention is disabled")
	}
}

func TestLogCleaner_CleanOldLogs_NonExistentDir(t *testing.T) {
	cleaner := NewLogCleaner(nil)
	
	// Should not error on non-existent directory
	err := cleaner.CleanOldLogs("/non/existent/path", 7)
	if err != nil {
		t.Errorf("CleanOldLogs should not error on non-existent directory: %v", err)
	}
}

func TestLogCleaner_CleanOldLogFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cleaner := NewLogCleaner(nil)
	
	// Create test files with different ages
	testFiles := []struct {
		name string
		age  time.Duration
		keep bool
	}{
		{"app-20240101.log", 10 * 24 * time.Hour, false}, // 10 days old
		{"app-20240105.log", 5 * 24 * time.Hour, false},  // 5 days old
		{"app-20240109.log", 1 * 24 * time.Hour, true},   // 1 day old
		{"app-20240110.log", 6 * time.Hour, true},        // 6 hours old
	}
	
	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		if err := os.WriteFile(filePath, []byte("test log"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Set modification time
		modTime := time.Now().Add(-tf.age)
		if err := os.Chtimes(filePath, modTime, modTime); err != nil {
			t.Fatalf("Failed to set file time: %v", err)
		}
	}
	
	// Clean files older than 3 days
	err := cleaner.CleanOldLogFiles(tmpDir, "app-*.log", 3)
	if err != nil {
		t.Fatalf("CleanOldLogFiles failed: %v", err)
	}
	
	// Verify results
	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		_, err := os.Stat(filePath)
		
		if tf.keep {
			if os.IsNotExist(err) {
				t.Errorf("File %s should have been kept but was removed", tf.name)
			}
		} else {
			if !os.IsNotExist(err) {
				t.Errorf("File %s should have been removed but still exists", tf.name)
			}
		}
	}
}

func TestLogCleaner_CalculateDirSize(t *testing.T) {
	tmpDir := t.TempDir()
	cleaner := NewLogCleaner(nil)
	
	// Create test directory with files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Create files of known sizes
	files := []struct {
		path string
		size int
	}{
		{filepath.Join(tmpDir, "file1.txt"), 100},
		{filepath.Join(tmpDir, "file2.txt"), 200},
		{filepath.Join(subDir, "file3.txt"), 300},
	}
	
	var expectedSize int64
	for _, f := range files {
		data := make([]byte, f.size)
		if err := os.WriteFile(f.path, data, 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		expectedSize += int64(f.size)
	}
	
	// Calculate directory size
	size, err := cleaner.calculateDirSize(tmpDir)
	if err != nil {
		t.Fatalf("calculateDirSize failed: %v", err)
	}
	
	if size != expectedSize {
		t.Errorf("calculateDirSize = %d, want %d", size, expectedSize)
	}
}

func TestFormatBytes(t *testing.T) {
	// This test is now removed as we use pipeline.FormatBytes
	t.Skip("formatBytes function removed, using pipeline.FormatBytes instead")
}
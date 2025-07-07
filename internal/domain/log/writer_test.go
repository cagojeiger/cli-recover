package log_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/log"
)

func TestNewWriter(t *testing.T) {
	// Test successful creation
	t.Run("Success", func(t *testing.T) {
		tempDir := t.TempDir()
		logPath := filepath.Join(tempDir, "test.log")

		writer, err := log.NewWriter(logPath)
		assert.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close()

		// Verify file exists
		assert.FileExists(t, logPath)

		// Verify header was written
		content, err := os.ReadFile(logPath)
		assert.NoError(t, err)
		assert.Contains(t, string(content), "=== Log started at")
	})

	// Test nested directory creation
	t.Run("Creates nested directories", func(t *testing.T) {
		tempDir := t.TempDir()
		logPath := filepath.Join(tempDir, "nested", "dir", "test.log")

		writer, err := log.NewWriter(logPath)
		assert.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close()

		assert.FileExists(t, logPath)
	})
}

func TestWriter_Write(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	writer, err := log.NewWriter(logPath)
	require.NoError(t, err)
	defer writer.Close()

	// Write data
	data := []byte("Test log entry\n")
	n, err := writer.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Verify content
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test log entry")
}

func TestWriter_WriteString(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	writer, err := log.NewWriter(logPath)
	require.NoError(t, err)
	defer writer.Close()

	// Write string
	n, err := writer.WriteString("Test string entry\n")
	assert.NoError(t, err)
	assert.Greater(t, n, 0)

	// Verify content
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test string entry")
}

func TestWriter_WriteLine(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	writer, err := log.NewWriter(logPath)
	require.NoError(t, err)
	defer writer.Close()

	// Write formatted line
	err = writer.WriteLine("Operation: %s, Status: %s", "backup", "running")
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	lines := strings.Split(string(content), "\n")
	
	// Find the line with our content
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Operation: backup, Status: running") {
			found = true
			// Should have timestamp
			assert.Contains(t, line, "[")
			assert.Contains(t, line, "]")
			// Timestamp format check
			assert.Regexp(t, `\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`, line)
			break
		}
	}
	assert.True(t, found, "Expected line not found")
}

func TestWriter_Close(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	writer, err := log.NewWriter(logPath)
	require.NoError(t, err)

	// Write some data
	writer.WriteString("Test data\n")

	// Close writer
	err = writer.Close()
	assert.NoError(t, err)

	// Verify footer was written
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "=== Log ended at")

	// Writing after close should fail
	_, err = writer.Write([]byte("Should fail"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")

	// Close again should not error
	err = writer.Close()
	assert.NoError(t, err)
}

func TestWriter_Concurrent(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	writer, err := log.NewWriter(logPath)
	require.NoError(t, err)
	defer writer.Close()

	// Write concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < 10; j++ {
				err := writer.WriteLine("Goroutine %d, Line %d", id, j)
				assert.NoError(t, err)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all lines were written
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	lines := strings.Split(string(content), "\n")
	
	// Count actual log lines (excluding header/footer)
	logLines := 0
	for _, line := range lines {
		if strings.Contains(line, "Goroutine") {
			logLines++
		}
	}
	assert.Equal(t, 100, logLines)
}

func TestMultiWriter(t *testing.T) {
	// Create buffers for testing
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	// Create multi-writer
	mw := log.NewMultiWriter(buf1, buf2)

	// Write data
	data := []byte("Test multi-write\n")
	n, err := mw.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Verify both buffers received the data
	assert.Equal(t, "Test multi-write\n", buf1.String())
	assert.Equal(t, "Test multi-write\n", buf2.String())
}

func TestTeeWriter(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Create log writer
	logWriter, err := log.NewWriter(logPath)
	require.NoError(t, err)
	defer logWriter.Close()

	// Create buffer for stdout
	stdout := &bytes.Buffer{}

	// Create tee writer
	tee := log.TeeWriter(logWriter, stdout)

	// Write data
	data := []byte("Tee output test\n")
	n, err := tee.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Verify both destinations received the data
	assert.Equal(t, "Tee output test\n", stdout.String())

	// Check log file
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Tee output test")
}
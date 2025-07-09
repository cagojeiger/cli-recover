package progress

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// Test CI Reporter logs at intervals
func TestCIReporter_LogsPeriodically(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	// Create reporter with short interval for testing
	reporter := NewCIReporter(&buf, 100*time.Millisecond)

	// Act
	reporter.Start("CI backup operation", 1000)
	reporter.Update(100, 1000) // 10%
	time.Sleep(150 * time.Millisecond)
	reporter.Update(500, 1000) // 50%
	time.Sleep(150 * time.Millisecond)
	reporter.Update(1000, 1000) // 100%
	reporter.Complete()

	// Assert
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have multiple log lines
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 log lines, got %d", len(lines))
	}

	// Progress update lines should contain structured info
	progressLineCount := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Only check lines that are progress updates (not start/complete)
		if strings.Contains(line, "current=") {
			progressLineCount++
			if !strings.Contains(line, "operation=") || !strings.Contains(line, "percent=") {
				t.Errorf("Expected structured log format, got: %s", line)
			}
		}
	}

	// Should have at least 2 progress updates
	if progressLineCount < 2 {
		t.Errorf("Expected at least 2 progress update lines, got %d", progressLineCount)
	}
}

// Test CI Reporter includes all required fields
func TestCIReporter_IncludesAllFields(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewCIReporter(&buf, time.Hour) // Long interval to control output

	// Act
	reporter.Start("Test operation", 1024*1024) // 1MB
	reporter.Update(512*1024, 1024*1024)        // 512KB
	reporter.forceLog()                         // Force immediate log for testing

	// Assert
	output := buf.String()

	// Should include all required fields
	requiredFields := []string{
		"operation=",
		"current=",
		"total=",
		"percent=",
		"rate=",
		"elapsed=",
	}

	for _, field := range requiredFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected output to contain field %s", field)
		}
	}
}

// Test CI Reporter handles completion
func TestCIReporter_LogsCompletion(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewCIReporter(&buf, time.Hour)

	// Act
	reporter.Start("Quick task", 100)
	reporter.Update(100, 100)
	reporter.Complete()

	// Assert
	output := buf.String()
	if !strings.Contains(output, "status=completed") {
		t.Error("Expected completion status in output")
	}
}

// Test CI Reporter handles errors
func TestCIReporter_LogsErrors(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewCIReporter(&buf, time.Hour)
	testErr := bytes.ErrTooLarge

	// Act
	reporter.Start("Failing task", 100)
	reporter.Error(testErr)

	// Assert
	output := buf.String()
	if !strings.Contains(output, "status=error") {
		t.Error("Expected error status in output")
	}
	if !strings.Contains(output, testErr.Error()) {
		t.Error("Expected error message in output")
	}
}

package progress

import (
	"bytes"
	"strings"
	"testing"
)

// Test Pipe Reporter outputs clean lines
func TestPipeReporter_OutputsCleanLines(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewPipeReporter(&buf)

	// Act
	reporter.Start("Pipe operation", 100)
	reporter.Update(25, 100)
	reporter.Update(50, 100)
	reporter.Update(75, 100)
	reporter.Update(100, 100)
	reporter.Complete()

	// Assert
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should NOT contain carriage returns (unlike terminal)
	if strings.Contains(output, "\r") {
		t.Error("Pipe output should not contain carriage returns")
	}

	// Each update should be on a new line
	if len(lines) < 5 {
		t.Errorf("Expected at least 5 lines of output, got %d", len(lines))
	}

	// Should contain progress percentages
	foundPercentages := 0
	for _, line := range lines {
		if strings.Contains(line, "25.0%") || strings.Contains(line, "50.0%") ||
			strings.Contains(line, "75.0%") || strings.Contains(line, "100.0%") {
			foundPercentages++
		}
	}
	if foundPercentages < 4 {
		t.Errorf("Expected to find 4 percentage updates, found %d", foundPercentages)
	}
}

// Test Pipe Reporter formats for parsing
func TestPipeReporter_MachineReadableFormat(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewPipeReporter(&buf)

	// Act
	reporter.Start("Backup", 1024)
	reporter.Update(512, 1024)

	// Assert
	output := buf.String()

	// Should have consistent, parseable format
	if !strings.Contains(output, "PROGRESS:") {
		t.Error("Expected PROGRESS: prefix for machine parsing")
	}

	// Should include numeric values for parsing
	if !strings.Contains(output, "512/1024") {
		t.Error("Expected raw byte counts for parsing")
	}
}

// Test Pipe Reporter completion message
func TestPipeReporter_CompletionMessage(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewPipeReporter(&buf)

	// Act
	reporter.Start("Task", 100)
	reporter.Update(100, 100)
	reporter.Complete()

	// Assert
	output := buf.String()
	if !strings.Contains(output, "COMPLETE:") {
		t.Error("Expected COMPLETE: prefix for completion")
	}
}

// Test Pipe Reporter error message
func TestPipeReporter_ErrorMessage(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewPipeReporter(&buf)
	testErr := bytes.ErrTooLarge

	// Act
	reporter.Start("Task", 100)
	reporter.Error(testErr)

	// Assert
	output := buf.String()
	if !strings.Contains(output, "ERROR:") {
		t.Error("Expected ERROR: prefix for errors")
	}
	if !strings.Contains(output, testErr.Error()) {
		t.Error("Expected error message in output")
	}
}

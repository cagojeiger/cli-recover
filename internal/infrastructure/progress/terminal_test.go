package progress

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// Test Terminal Reporter updates in place with carriage return
func TestTerminalReporter_UpdatesInPlace(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewTerminalReporter(&buf)

	// Act
	reporter.Start("Backing up files", 100)
	time.Sleep(110 * time.Millisecond) // Wait for throttle period
	reporter.Update(25, 100)
	time.Sleep(110 * time.Millisecond) // Wait for throttle period
	reporter.Update(50, 100)
	time.Sleep(110 * time.Millisecond) // Wait for throttle period
	reporter.Update(75, 100)
	reporter.Complete()

	// Assert
	output := buf.String()

	// Should contain carriage returns for in-place updates
	if !strings.Contains(output, "\r") {
		t.Error("Expected output to contain carriage returns for in-place updates")
	}

	// Should show progress percentages (with some tolerance for formatting)
	if !strings.Contains(output, "25.0%") && !strings.Contains(output, "25%") {
		t.Errorf("Expected output to contain 25%% progress, got: %s", output)
	}
	if !strings.Contains(output, "50.0%") && !strings.Contains(output, "50%") {
		t.Errorf("Expected output to contain 50%% progress, got: %s", output)
	}
	if !strings.Contains(output, "75.0%") && !strings.Contains(output, "75%") {
		t.Errorf("Expected output to contain 75%% progress, got: %s", output)
	}

	// Should end with a newline after completion
	if !strings.HasSuffix(output, "\n") {
		t.Error("Expected output to end with newline after completion")
	}
}

// Test Terminal Reporter shows progress bar
func TestTerminalReporter_ShowsProgressBar(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewTerminalReporter(&buf)

	// Act
	reporter.Start("Testing", 100)
	reporter.Update(40, 100)

	// Assert
	output := buf.String()

	// Should contain progress bar characters
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Expected output to contain progress bar brackets")
	}

	// Should contain filled and unfilled sections
	if !strings.Contains(output, "=") {
		t.Error("Expected output to contain progress bar fill characters")
	}
}

// Test Terminal Reporter formats sizes correctly
func TestTerminalReporter_FormatsSize(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewTerminalReporter(&buf)

	// Act
	reporter.Start("Large backup", 1024*1024*1024) // 1GB
	reporter.Update(512*1024*1024, 1024*1024*1024) // 512MB

	// Assert
	output := buf.String()

	// Should format sizes in human-readable format
	if !strings.Contains(output, "MB") || !strings.Contains(output, "GB") {
		t.Error("Expected output to contain human-readable sizes")
	}
}

// Test Terminal Reporter handles zero total gracefully
func TestTerminalReporter_HandlesZeroTotal(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewTerminalReporter(&buf)

	// Act & Assert - should not panic or divide by zero
	reporter.Start("Unknown size", 0)
	reporter.Update(100, 0)
	reporter.Complete()
}

// Test Terminal Reporter throttles updates
func TestTerminalReporter_ThrottlesUpdates(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := NewTerminalReporter(&buf)
	reporter.Start("Fast updates", 1000)

	// Act - rapid updates
	for i := 0; i < 100; i++ {
		reporter.Update(int64(i*10), 1000)
		time.Sleep(1 * time.Millisecond) // Very fast updates
	}

	// Assert
	output := buf.String()
	updateCount := strings.Count(output, "\r")

	// Should have fewer updates than calls (due to throttling)
	// Allow some margin for timing variations
	if updateCount >= 100 {
		t.Errorf("Expected throttled updates (less than 100), got %d", updateCount)
	}
}

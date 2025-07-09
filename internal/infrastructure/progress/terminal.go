package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/progress"
)

// TerminalReporter displays progress in an interactive terminal with in-place updates
type TerminalReporter struct {
	writer       io.Writer
	mu           sync.Mutex
	lastUpdate   time.Time
	updatePeriod time.Duration
	startTime    time.Time
	operation    string
}

// NewTerminalReporter creates a new terminal progress reporter
func NewTerminalReporter(w io.Writer) *TerminalReporter {
	return &TerminalReporter{
		writer:       w,
		updatePeriod: 100 * time.Millisecond, // Throttle updates to 10 per second
	}
}

// Start begins tracking a new operation
func (t *TerminalReporter) Start(operation string, total int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.operation = operation
	t.startTime = time.Now()
	t.lastUpdate = time.Time{} // Force first update

	// Clear line and show initial message
	fmt.Fprintf(t.writer, "\r%s: Starting...", operation)
}

// Update reports progress
func (t *TerminalReporter) Update(current, total int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Throttle updates
	now := time.Now()
	if now.Sub(t.lastUpdate) < t.updatePeriod {
		return
	}
	t.lastUpdate = now

	// Calculate metrics
	percent := 0.0
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}

	elapsed := now.Sub(t.startTime)
	rate := float64(current) / elapsed.Seconds()

	// Build progress bar
	barWidth := 30
	filled := int(percent / 100 * float64(barWidth))
	bar := fmt.Sprintf("[%s%s]",
		strings.Repeat("=", filled),
		strings.Repeat("-", barWidth-filled))

	// Format sizes
	currentStr := formatBytes(current)
	totalStr := formatBytes(total)
	rateStr := formatBytes(int64(rate)) + "/s"

	// Clear line and update
	fmt.Fprintf(t.writer, "\r%s: %s %.1f%% (%s/%s) %s",
		t.operation, bar, percent, currentStr, totalStr, rateStr)
}

// Complete marks the operation as completed
func (t *TerminalReporter) Complete() {
	t.mu.Lock()
	defer t.mu.Unlock()

	elapsed := time.Since(t.startTime)
	fmt.Fprintf(t.writer, "\r%s: Completed in %s\n", t.operation, elapsed.Round(time.Second))
}

// Error reports an error
func (t *TerminalReporter) Error(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Fprintf(t.writer, "\r%s: Error - %v\n", t.operation, err)
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Ensure TerminalReporter implements ProgressReporter
var _ progress.ProgressReporter = (*TerminalReporter)(nil)

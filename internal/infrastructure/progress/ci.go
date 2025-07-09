package progress

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/progress"
)

// CIReporter outputs structured logs suitable for CI/CD environments
type CIReporter struct {
	writer      io.Writer
	mu          sync.Mutex
	lastLog     time.Time
	logInterval time.Duration
	startTime   time.Time
	operation   string
	lastCurrent int64
	lastTotal   int64
}

// NewCIReporter creates a new CI progress reporter
func NewCIReporter(w io.Writer, interval time.Duration) *CIReporter {
	if interval <= 0 {
		interval = 10 * time.Second // Default to 10 second intervals
	}
	return &CIReporter{
		writer:      w,
		logInterval: interval,
	}
}

// Start begins tracking a new operation
func (c *CIReporter) Start(operation string, total int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.operation = operation
	c.startTime = time.Now()
	c.lastLog = time.Time{} // Force first log
	c.lastTotal = total

	fmt.Fprintf(c.writer, "[PROGRESS] operation=%s status=started total=%d timestamp=%s\n",
		operation, total, c.startTime.Format(time.RFC3339))
}

// Update reports progress
func (c *CIReporter) Update(current, total int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastCurrent = current
	c.lastTotal = total

	// Check if we should log
	now := time.Now()
	if now.Sub(c.lastLog) >= c.logInterval || c.lastLog.IsZero() {
		c.logProgress(now)
	}
}

// Complete marks the operation as completed
func (c *CIReporter) Complete() {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.startTime)
	fmt.Fprintf(c.writer, "[PROGRESS] operation=%s status=completed elapsed=%s total=%d timestamp=%s\n",
		c.operation, elapsed, c.lastTotal, time.Now().Format(time.RFC3339))
}

// Error reports an error
func (c *CIReporter) Error(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Fprintf(c.writer, "[PROGRESS] operation=%s status=error error=%q timestamp=%s\n",
		c.operation, err.Error(), time.Now().Format(time.RFC3339))
}

// forceLog forces an immediate log (for testing)
func (c *CIReporter) forceLog() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logProgress(time.Now())
}

// logProgress logs the current progress
func (c *CIReporter) logProgress(now time.Time) {
	c.lastLog = now
	elapsed := now.Sub(c.startTime)

	percent := 0.0
	if c.lastTotal > 0 {
		percent = float64(c.lastCurrent) / float64(c.lastTotal) * 100
	}

	rate := float64(c.lastCurrent) / elapsed.Seconds()

	fmt.Fprintf(c.writer, "[PROGRESS] operation=%s current=%d total=%d percent=%.1f rate=%.0f elapsed=%s timestamp=%s\n",
		c.operation, c.lastCurrent, c.lastTotal, percent, rate, elapsed, now.Format(time.RFC3339))
}

// Ensure CIReporter implements ProgressReporter
var _ progress.ProgressReporter = (*CIReporter)(nil)

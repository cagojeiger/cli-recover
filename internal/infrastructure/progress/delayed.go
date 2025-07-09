package progress

import (
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/progress"
)

// DelayedReporter wraps another reporter and delays showing progress until 3 seconds have passed
type DelayedReporter struct {
	wrapped     progress.ProgressReporter
	mu          sync.Mutex
	started     bool
	startTime   time.Time
	operation   string
	totalBytes  int64
	lastCurrent int64
	lastTotal   int64
	delayPeriod time.Duration
	timer       *time.Timer
}

// NewDelayedReporter creates a reporter that waits 3 seconds before showing progress
func NewDelayedReporter(wrapped progress.ProgressReporter) *DelayedReporter {
	return &DelayedReporter{
		wrapped:     wrapped,
		delayPeriod: 3 * time.Second,
	}
}

// Start begins tracking a new operation
func (d *DelayedReporter) Start(operation string, total int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.started = false
	d.startTime = time.Now()
	d.operation = operation
	d.totalBytes = total
	d.lastCurrent = 0
	d.lastTotal = total

	// Cancel any existing timer
	if d.timer != nil {
		d.timer.Stop()
	}

	// Start a timer to show progress after 3 seconds
	d.timer = time.AfterFunc(d.delayPeriod, func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if !d.started {
			d.started = true
			d.wrapped.Start(d.operation, d.totalBytes)
			// Show the latest progress immediately
			if d.lastCurrent > 0 {
				d.wrapped.Update(d.lastCurrent, d.lastTotal)
			}
		}
	})
}

// Update reports progress (buffered until 3 seconds have passed)
func (d *DelayedReporter) Update(current, total int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Always store the latest values
	d.lastCurrent = current
	d.lastTotal = total

	// Only forward to wrapped reporter if we've started showing progress
	if d.started {
		d.wrapped.Update(current, total)
	}
}

// Complete marks the operation as completed
func (d *DelayedReporter) Complete() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel the timer if operation completes before 3 seconds
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}

	// If operation took less than 3 seconds, don't show progress at all
	elapsed := time.Since(d.startTime)
	if elapsed < d.delayPeriod && !d.started {
		// Operation completed quickly, no need to show progress
		return
	}

	// If we had started showing progress, complete it
	if d.started {
		d.wrapped.Complete()
	} else {
		// If we hadn't started but the operation took longer than 3 seconds,
		// show a simple completion message
		d.wrapped.Start(d.operation, d.totalBytes)
		d.wrapped.Complete()
	}
}

// Error reports an error
func (d *DelayedReporter) Error(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel the timer
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}

	// Always show errors, even if less than 3 seconds
	if !d.started {
		d.wrapped.Start(d.operation, d.totalBytes)
	}
	d.wrapped.Error(err)
}

// Ensure DelayedReporter implements ProgressReporter
var _ progress.ProgressReporter = (*DelayedReporter)(nil)

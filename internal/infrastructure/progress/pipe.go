package progress

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/progress"
)

// PipeReporter outputs clean, parseable progress suitable for piped output
type PipeReporter struct {
	writer    io.Writer
	mu        sync.Mutex
	startTime time.Time
	operation string
}

// NewPipeReporter creates a new pipe progress reporter
func NewPipeReporter(w io.Writer) *PipeReporter {
	return &PipeReporter{
		writer: w,
	}
}

// Start begins tracking a new operation
func (p *PipeReporter) Start(operation string, total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.operation = operation
	p.startTime = time.Now()

	fmt.Fprintf(p.writer, "PROGRESS: %s started (total: %d bytes)\n", operation, total)
}

// Update reports progress
func (p *PipeReporter) Update(current, total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	percent := 0.0
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}

	fmt.Fprintf(p.writer, "PROGRESS: %s - %d/%d bytes (%.1f%%)\n",
		p.operation, current, total, percent)
}

// Complete marks the operation as completed
func (p *PipeReporter) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	elapsed := time.Since(p.startTime)
	fmt.Fprintf(p.writer, "COMPLETE: %s finished in %s\n", p.operation, elapsed)
}

// Error reports an error
func (p *PipeReporter) Error(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprintf(p.writer, "ERROR: %s failed - %v\n", p.operation, err)
}

// Ensure PipeReporter implements ProgressReporter
var _ progress.ProgressReporter = (*PipeReporter)(nil)

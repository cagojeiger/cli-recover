package progress

import "io"

// ProgressReporter handles progress updates for different environments
type ProgressReporter interface {
	Start(operation string, total int64)
	Update(current, total int64)
	Complete()
	Error(err error)
}

// ProgressWriter wraps an io.Writer to track and report progress
type ProgressWriter struct {
	writer   io.Writer
	reporter ProgressReporter
	current  int64
	total    int64
}

// NewProgressWriter creates a new progress tracking writer
func NewProgressWriter(w io.Writer, reporter ProgressReporter, total int64) *ProgressWriter {
	return &ProgressWriter{
		writer:   w,
		reporter: reporter,
		current:  0,
		total:    total,
	}
}

// Write implements io.Writer, tracking bytes written and reporting progress
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.current += int64(n)

	// Report progress if we have a reporter
	if pw.reporter != nil {
		pw.reporter.Update(pw.current, pw.total)
	}

	return n, err
}

// Current returns the number of bytes written so far
func (pw *ProgressWriter) Current() int64 {
	return pw.current
}

// Total returns the total expected bytes
func (pw *ProgressWriter) Total() int64 {
	return pw.total
}

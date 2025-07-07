package tui

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// JobStatus represents the current state of a backup job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

const (
	maxOutputLines = 500 // Maximum lines to keep in output buffer
)

// BackupJob represents a single backup operation
type BackupJob struct {
	ID        string
	Command   string
	Status    JobStatus
	Progress  int
	Output    []string
	Error     error
	StartTime time.Time
	EndTime   time.Time
	
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewBackupJob creates a new backup job
func NewBackupJob(id, command string) (*BackupJob, error) {
	if id == "" {
		return nil, errors.New("job ID cannot be empty")
	}
	if command == "" {
		return nil, errors.New("command cannot be empty")
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &BackupJob{
		ID:       id,
		Command:  command,
		Status:   JobStatusPending,
		Progress: 0,
		Output:   make([]string, 0),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// NewBackupJobWithTimeout creates a new backup job with timeout
func NewBackupJobWithTimeout(id, command string, timeout time.Duration) (*BackupJob, error) {
	job, err := NewBackupJob(id, command)
	if err != nil {
		return nil, err
	}
	
	// Replace context with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	job.ctx = ctx
	job.cancel = cancel
	
	return job, nil
}

// Start begins execution of the job
func (j *BackupJob) Start() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if j.Status != JobStatusPending {
		return fmt.Errorf("cannot start job in status: %s", j.Status)
	}
	
	j.Status = JobStatusRunning
	j.StartTime = time.Now()
	return nil
}

// Complete marks the job as completed
func (j *BackupJob) Complete(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if j.Status != JobStatusRunning {
		return
	}
	
	j.EndTime = time.Now()
	j.Error = err
	
	if err != nil {
		j.Status = JobStatusFailed
	} else {
		j.Status = JobStatusCompleted
	}
}

// Cancel cancels the job
func (j *BackupJob) Cancel() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	if j.Status != JobStatusRunning {
		return fmt.Errorf("cannot cancel job in status: %s", j.Status)
	}
	
	j.Status = JobStatusCancelled
	j.EndTime = time.Now()
	j.cancel()
	
	return nil
}

// UpdateProgress updates the job progress
func (j *BackupJob) UpdateProgress(progress int, message string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	// Clamp progress to 0-100
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}
	
	j.Progress = progress
	
	if message != "" {
		j.addOutputLocked(message)
	}
}

// AddOutput adds a line to the output buffer
func (j *BackupJob) AddOutput(line string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	
	j.addOutputLocked(line)
}

// addOutputLocked adds output without locking (must be called with lock held)
func (j *BackupJob) addOutputLocked(line string) {
	j.Output = append(j.Output, line)
	
	// Keep only last N lines
	if len(j.Output) > maxOutputLines {
		j.Output = j.Output[len(j.Output)-maxOutputLines:]
	}
}

// GetStatus returns the current job status
func (j *BackupJob) GetStatus() JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status
}

// GetProgress returns the current progress
func (j *BackupJob) GetProgress() int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Progress
}

// GetOutput returns a copy of the output
func (j *BackupJob) GetOutput() []string {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	output := make([]string, len(j.Output))
	copy(output, j.Output)
	return output
}

// Duration returns how long the job has been running or ran
func (j *BackupJob) Duration() time.Duration {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	if j.StartTime.IsZero() {
		return 0
	}
	
	if j.EndTime.IsZero() {
		// Still running
		return time.Since(j.StartTime)
	}
	
	// Completed
	return j.EndTime.Sub(j.StartTime)
}

// Context returns the job's context
func (j *BackupJob) Context() context.Context {
	return j.ctx
}

// IsActive returns true if the job is currently running
func (j *BackupJob) IsActive() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status == JobStatusRunning
}

// CanCancel returns true if the job can be cancelled
func (j *BackupJob) CanCancel() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status == JobStatusRunning
}
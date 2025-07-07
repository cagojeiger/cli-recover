package tui

import (
	"fmt"
	"sync"
	"time"
)

// JobManager manages backup jobs without goroutines
// It's a simple state manager that works with Bubble Tea's message system
type JobManager struct {
	jobs      map[string]*BackupJob
	jobOrder  []string // Maintain insertion order
	maxJobs   int
	mu        sync.RWMutex
}

// NewJobManager creates a new job manager
func NewJobManager(maxJobs int) *JobManager {
	if maxJobs <= 0 {
		maxJobs = 3
	}
	
	return &JobManager{
		jobs:     make(map[string]*BackupJob),
		jobOrder: make([]string, 0),
		maxJobs:  maxJobs,
	}
}

// Add adds a new job to the manager
func (m *JobManager) Add(job *BackupJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.jobs[job.ID]; exists {
		return fmt.Errorf("job %s already exists", job.ID)
	}
	
	m.jobs[job.ID] = job
	m.jobOrder = append(m.jobOrder, job.ID)
	
	// Set initial status
	if m.getActiveCount() < m.maxJobs {
		job.Status = JobStatusPending
	} else {
		job.Status = JobStatusQueued
	}
	
	return nil
}

// Remove removes a job from the manager
func (m *JobManager) Remove(jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.jobs, jobID)
	
	// Remove from order list
	for i, id := range m.jobOrder {
		if id == jobID {
			m.jobOrder = append(m.jobOrder[:i], m.jobOrder[i+1:]...)
			break
		}
	}
}

// Get returns a specific job
func (m *JobManager) Get(jobID string) *BackupJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.jobs[jobID]
}

// GetAll returns all jobs in order
func (m *JobManager) GetAll() []*BackupJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]*BackupJob, 0, len(m.jobs))
	for _, id := range m.jobOrder {
		if job, exists := m.jobs[id]; exists {
			result = append(result, job)
		}
	}
	
	return result
}

// GetActive returns all active (running) jobs
func (m *JobManager) GetActive() []*BackupJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]*BackupJob, 0)
	for _, id := range m.jobOrder {
		if job, exists := m.jobs[id]; exists && job.Status == JobStatusRunning {
			result = append(result, job)
		}
	}
	
	return result
}

// GetQueued returns all queued jobs
func (m *JobManager) GetQueued() []*BackupJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]*BackupJob, 0)
	for _, id := range m.jobOrder {
		if job, exists := m.jobs[id]; exists && job.Status == JobStatusQueued {
			result = append(result, job)
		}
	}
	
	return result
}

// GetNextQueued returns the next queued job to run
func (m *JobManager) GetNextQueued() *BackupJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, id := range m.jobOrder {
		if job, exists := m.jobs[id]; exists && job.Status == JobStatusQueued {
			return job
		}
	}
	
	return nil
}

// CanRunMore checks if more jobs can be run
func (m *JobManager) CanRunMore() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.getActiveCount() < m.maxJobs
}

// UpdateStatus updates a job's status
func (m *JobManager) UpdateStatus(jobID string, status JobStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	
	job.Status = status
	
	// Update timestamps
	switch status {
	case JobStatusRunning:
		if job.StartTime.IsZero() {
			job.StartTime = time.Now()
		}
	case JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		if job.EndTime.IsZero() {
			job.EndTime = time.Now()
		}
	}
	
	return nil
}

// MarkCompleted marks a job as completed
func (m *JobManager) MarkCompleted(jobID string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	
	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err
	} else {
		job.Status = JobStatusCompleted
	}
	
	if job.EndTime.IsZero() {
		job.EndTime = time.Now()
	}
	
	return nil
}

// CancelJob cancels a specific job
func (m *JobManager) CancelJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	
	// Can only cancel pending, queued, or running jobs
	switch job.Status {
	case JobStatusPending, JobStatusQueued, JobStatusRunning:
		// Cancel the job's context if it's running
		if job.cancel != nil {
			job.cancel()
		}
		job.Status = JobStatusCancelled
		if job.EndTime.IsZero() {
			job.EndTime = time.Now()
		}
	default:
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}
	
	return nil
}

// CancelAll cancels all active jobs
func (m *JobManager) CancelAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, job := range m.jobs {
		switch job.Status {
		case JobStatusPending, JobStatusQueued, JobStatusRunning:
			if job.cancel != nil {
				job.cancel()
			}
			job.Status = JobStatusCancelled
			if job.EndTime.IsZero() {
				job.EndTime = time.Now()
			}
		}
	}
}

// Cleanup removes completed/failed/cancelled jobs older than duration
func (m *JobManager) Cleanup(olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	removed := 0
	
	// Create new order list
	newOrder := make([]string, 0)
	
	for _, id := range m.jobOrder {
		job, exists := m.jobs[id]
		if !exists {
			continue
		}
		
		// Keep active jobs and recent completed jobs
		shouldKeep := false
		switch job.Status {
		case JobStatusPending, JobStatusQueued, JobStatusRunning:
			shouldKeep = true
		case JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
			if job.EndTime.After(cutoff) {
				shouldKeep = true
			}
		}
		
		if shouldKeep {
			newOrder = append(newOrder, id)
		} else {
			delete(m.jobs, id)
			removed++
		}
	}
	
	m.jobOrder = newOrder
	return removed
}

// GetStats returns job statistics
func (m *JobManager) GetStats() (active, queued, completed, failed, cancelled int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, job := range m.jobs {
		switch job.Status {
		case JobStatusRunning:
			active++
		case JobStatusQueued:
			queued++
		case JobStatusCompleted:
			completed++
		case JobStatusFailed:
			failed++
		case JobStatusCancelled:
			cancelled++
		}
	}
	
	return
}

// GetMaxJobs returns the maximum number of concurrent jobs
func (m *JobManager) GetMaxJobs() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.maxJobs
}

// getActiveCount returns the number of active jobs (internal, no lock)
func (m *JobManager) getActiveCount() int {
	count := 0
	for _, job := range m.jobs {
		if job.Status == JobStatusRunning {
			count++
		}
	}
	return count
}
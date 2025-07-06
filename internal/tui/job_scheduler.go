package tui

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// JobExecutor interface for executing jobs
type JobExecutor interface {
	Execute(job *BackupJob) error
}

// JobScheduler manages concurrent backup jobs
type JobScheduler struct {
	maxJobs     int
	activeJobs  map[string]*BackupJob
	jobQueue    []*BackupJob
	executor    JobExecutor
	mu          sync.RWMutex
	jobComplete chan string
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(maxJobs int) *JobScheduler {
	if maxJobs <= 0 {
		maxJobs = 1
	}

	return &JobScheduler{
		maxJobs:     maxJobs,
		activeJobs:  make(map[string]*BackupJob),
		jobQueue:    make([]*BackupJob, 0),
		jobComplete: make(chan string, maxJobs),
	}
}

// SetExecutor sets the job executor
func (s *JobScheduler) SetExecutor(executor JobExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executor = executor
}

// Submit adds a job to the scheduler
func (s *JobScheduler) Submit(job *BackupJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job already exists
	if _, exists := s.activeJobs[job.ID]; exists {
		return fmt.Errorf("job %s already exists in active jobs", job.ID)
	}

	// Check queue for duplicates
	for _, queuedJob := range s.jobQueue {
		if queuedJob.ID == job.ID {
			return fmt.Errorf("job %s already exists in queue", job.ID)
		}
	}

	// Add to active jobs if there's capacity
	if len(s.activeJobs) < s.maxJobs {
		s.activeJobs[job.ID] = job
		return nil
	}

	// Otherwise queue the job
	s.jobQueue = append(s.jobQueue, job)
	return nil
}

// Cancel cancels a job
func (s *JobScheduler) Cancel(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check active jobs
	if job, exists := s.activeJobs[jobID]; exists {
		if job.CanCancel() {
			err := job.Cancel()
			if err != nil {
				return err
			}
		}
		delete(s.activeJobs, jobID)
		return nil
	}

	// Check queued jobs
	for i, job := range s.jobQueue {
		if job.ID == jobID {
			// Remove from queue
			s.jobQueue = append(s.jobQueue[:i], s.jobQueue[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("job %s not found", jobID)
}

// GetJob returns a specific job
func (s *JobScheduler) GetJob(jobID string) *BackupJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check active jobs
	if job, exists := s.activeJobs[jobID]; exists {
		return job
	}

	// Check queued jobs
	for _, job := range s.jobQueue {
		if job.ID == jobID {
			return job
		}
	}

	return nil
}

// GetActiveJobs returns a copy of active jobs
func (s *JobScheduler) GetActiveJobs() []*BackupJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*BackupJob, 0, len(s.activeJobs))
	for _, job := range s.activeJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetQueuedJobs returns a copy of queued jobs
func (s *JobScheduler) GetQueuedJobs() []*BackupJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*BackupJob, len(s.jobQueue))
	copy(jobs, s.jobQueue)
	return jobs
}

// GetAllJobs returns all jobs (active and queued)
func (s *JobScheduler) GetAllJobs() []*BackupJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allJobs := make([]*BackupJob, 0, len(s.activeJobs)+len(s.jobQueue))
	
	// Add active jobs
	for _, job := range s.activeJobs {
		allJobs = append(allJobs, job)
	}
	
	// Add queued jobs
	allJobs = append(allJobs, s.jobQueue...)
	
	return allJobs
}

// GetMaxJobs returns the maximum number of concurrent jobs
func (s *JobScheduler) GetMaxJobs() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.maxJobs
}

// Start begins processing jobs
func (s *JobScheduler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Shutdown requested
			s.shutdown()
			return
			
		case jobID := <-s.jobComplete:
			// Job completed, remove from active and process queue
			s.mu.Lock()
			delete(s.activeJobs, jobID)
			
			// Process next queued job if any
			if len(s.jobQueue) > 0 && len(s.activeJobs) < s.maxJobs {
				nextJob := s.jobQueue[0]
				s.jobQueue = s.jobQueue[1:]
				s.activeJobs[nextJob.ID] = nextJob
				
				// Start the job asynchronously
				go s.executeJob(nextJob)
			}
			s.mu.Unlock()
			
		default:
			// Process any pending jobs
			s.mu.Lock()
			needsProcessing := make([]*BackupJob, 0)
			
			for _, job := range s.activeJobs {
				if job.GetStatus() == JobStatusPending {
					needsProcessing = append(needsProcessing, job)
				}
			}
			s.mu.Unlock()
			
			// Start pending jobs
			for _, job := range needsProcessing {
				go s.executeJob(job)
			}
			
			// Small delay to prevent busy loop
			select {
			case <-ctx.Done():
				s.shutdown()
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
}

// executeJob runs a job and notifies completion
func (s *JobScheduler) executeJob(job *BackupJob) {
	// Start the job
	err := job.Start()
	if err != nil {
		job.Complete(err)
		s.jobComplete <- job.ID
		return
	}

	// Execute the job if executor is set
	if s.executor != nil {
		err = s.executor.Execute(job)
		job.Complete(err)
	} else {
		// No executor, just mark as completed
		job.Complete(nil)
	}

	// Notify completion
	s.jobComplete <- job.ID
}

// shutdown cancels all active jobs
func (s *JobScheduler) shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel all active jobs
	for _, job := range s.activeJobs {
		if job.CanCancel() {
			job.Cancel()
		}
	}

	// Clear queues
	s.activeJobs = make(map[string]*BackupJob)
	s.jobQueue = make([]*BackupJob, 0)
}
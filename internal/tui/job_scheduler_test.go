package tui

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewJobScheduler(t *testing.T) {
	tests := []struct {
		name         string
		maxJobs      int
		wantMaxJobs  int
	}{
		{
			name:        "valid max jobs",
			maxJobs:     3,
			wantMaxJobs: 3,
		},
		{
			name:        "zero max jobs defaults to 1",
			maxJobs:     0,
			wantMaxJobs: 1,
		},
		{
			name:        "negative max jobs defaults to 1",
			maxJobs:     -5,
			wantMaxJobs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewJobScheduler(tt.maxJobs)
			assert.NotNil(t, scheduler)
			assert.Equal(t, tt.wantMaxJobs, scheduler.GetMaxJobs())
			assert.Empty(t, scheduler.GetActiveJobs())
			assert.Empty(t, scheduler.GetQueuedJobs())
		})
	}
}

func TestJobSchedulerSubmit(t *testing.T) {
	scheduler := NewJobScheduler(2) // Max 2 concurrent jobs

	// Create test jobs
	job1, _ := NewBackupJob("job-1", "echo test1")
	job2, _ := NewBackupJob("job-2", "echo test2")
	job3, _ := NewBackupJob("job-3", "echo test3")

	// Submit first job
	err := scheduler.Submit(job1)
	assert.NoError(t, err)
	assert.Len(t, scheduler.GetActiveJobs(), 1)
	assert.Len(t, scheduler.GetQueuedJobs(), 0)

	// Submit second job
	err = scheduler.Submit(job2)
	assert.NoError(t, err)
	assert.Len(t, scheduler.GetActiveJobs(), 2)
	assert.Len(t, scheduler.GetQueuedJobs(), 0)

	// Submit third job (should be queued)
	err = scheduler.Submit(job3)
	assert.NoError(t, err)
	assert.Len(t, scheduler.GetActiveJobs(), 2)
	assert.Len(t, scheduler.GetQueuedJobs(), 1)

	// Cannot submit duplicate job
	err = scheduler.Submit(job1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestJobSchedulerCancel(t *testing.T) {
	scheduler := NewJobScheduler(2)

	// Create and submit jobs
	job1, _ := NewBackupJob("job-1", "sleep 10")
	job2, _ := NewBackupJob("job-2", "sleep 10")
	job3, _ := NewBackupJob("job-3", "sleep 10")

	scheduler.Submit(job1)
	scheduler.Submit(job2)
	scheduler.Submit(job3) // This one is queued

	// Start the active jobs
	job1.Start()
	job2.Start()

	// Cancel active job
	err := scheduler.Cancel("job-1")
	assert.NoError(t, err)
	assert.Equal(t, JobStatusCancelled, job1.GetStatus())

	// Cancel queued job
	err = scheduler.Cancel("job-3")
	assert.NoError(t, err)
	assert.Len(t, scheduler.GetQueuedJobs(), 0)

	// Cancel non-existent job
	err = scheduler.Cancel("job-999")
	assert.Error(t, err)
}

func TestJobSchedulerConcurrency(t *testing.T) {
	scheduler := NewJobScheduler(3)
	
	// Start scheduler processing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go scheduler.Start(ctx)

	// Track job completions
	completedJobs := make(map[string]bool)
	var mu sync.Mutex

	// Create mock process manager
	mockPM := &MockJobExecutor{
		executeFunc: func(job *BackupJob) error {
			// Simulate work
			time.Sleep(100 * time.Millisecond)
			
			mu.Lock()
			completedJobs[job.ID] = true
			mu.Unlock()
			
			return nil
		},
	}

	scheduler.SetExecutor(mockPM)

	// Submit 6 jobs (max 3 concurrent)
	for i := 0; i < 6; i++ {
		job, _ := NewBackupJob(fmt.Sprintf("job-%d", i), "echo test")
		err := scheduler.Submit(job)
		assert.NoError(t, err)
	}

	// Verify initial state
	assert.Len(t, scheduler.GetActiveJobs(), 3)
	assert.Len(t, scheduler.GetQueuedJobs(), 3)

	// Wait for all jobs to complete
	// 6 jobs * 100ms each, but only 3 can run concurrently
	// So minimum time is 200ms (2 batches)
	time.Sleep(500 * time.Millisecond)

	// All jobs should be completed
	mu.Lock()
	assert.Len(t, completedJobs, 6)
	mu.Unlock()

	assert.Empty(t, scheduler.GetActiveJobs())
	assert.Empty(t, scheduler.GetQueuedJobs())
}

func TestJobSchedulerGetJob(t *testing.T) {
	scheduler := NewJobScheduler(2)

	job1, _ := NewBackupJob("job-1", "echo test")
	scheduler.Submit(job1)

	// Get existing job
	retrieved := scheduler.GetJob("job-1")
	assert.NotNil(t, retrieved)
	assert.Equal(t, job1.ID, retrieved.ID)

	// Get non-existent job
	retrieved = scheduler.GetJob("job-999")
	assert.Nil(t, retrieved)
}

func TestJobSchedulerGetAllJobs(t *testing.T) {
	scheduler := NewJobScheduler(2)

	// Submit mix of active and queued jobs
	job1, _ := NewBackupJob("job-1", "echo test")
	job2, _ := NewBackupJob("job-2", "echo test")
	job3, _ := NewBackupJob("job-3", "echo test")

	scheduler.Submit(job1)
	scheduler.Submit(job2)
	scheduler.Submit(job3)

	// Get all jobs
	allJobs := scheduler.GetAllJobs()
	assert.Len(t, allJobs, 3)

	// Verify all jobs are included
	jobIDs := make(map[string]bool)
	for _, job := range allJobs {
		jobIDs[job.ID] = true
	}
	assert.True(t, jobIDs["job-1"])
	assert.True(t, jobIDs["job-2"])
	assert.True(t, jobIDs["job-3"])
}

func TestJobSchedulerQueueProcessing(t *testing.T) {
	scheduler := NewJobScheduler(1) // Only 1 concurrent job
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Track execution order
	executionOrder := make([]string, 0)
	var orderMu sync.Mutex

	mockPM := &MockJobExecutor{
		executeFunc: func(job *BackupJob) error {
			orderMu.Lock()
			executionOrder = append(executionOrder, job.ID)
			orderMu.Unlock()
			
			// Simulate quick work
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}

	scheduler.SetExecutor(mockPM)
	go scheduler.Start(ctx)

	// Submit jobs in order
	for i := 0; i < 3; i++ {
		job, _ := NewBackupJob(fmt.Sprintf("job-%d", i), "echo test")
		scheduler.Submit(job)
	}

	// Wait for all jobs to complete
	// 3 jobs * 50ms each = 150ms minimum
	time.Sleep(300 * time.Millisecond)

	// Verify execution order (FIFO)
	orderMu.Lock()
	assert.Equal(t, []string{"job-0", "job-1", "job-2"}, executionOrder)
	orderMu.Unlock()
}

func TestJobSchedulerShutdown(t *testing.T) {
	scheduler := NewJobScheduler(2)
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start scheduler
	go scheduler.Start(ctx)

	// Submit some jobs
	job1, _ := NewBackupJob("job-1", "sleep 10")
	job2, _ := NewBackupJob("job-2", "sleep 10")
	
	scheduler.Submit(job1)
	scheduler.Submit(job2)

	// Start jobs
	job1.Start()
	job2.Start()

	// Shutdown scheduler
	cancel()
	
	// Give time for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	// All jobs should be cancelled
	assert.Equal(t, JobStatusCancelled, job1.GetStatus())
	assert.Equal(t, JobStatusCancelled, job2.GetStatus())
}

func TestJobSchedulerRaceConditions(t *testing.T) {
	scheduler := NewJobScheduler(5)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go scheduler.Start(ctx)

	// Concurrent submissions
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			job, _ := NewBackupJob(fmt.Sprintf("job-%d", n), "echo test")
			scheduler.Submit(job)
		}(i)
	}

	// Concurrent cancellations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			scheduler.Cancel(fmt.Sprintf("job-%d", n))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				scheduler.GetActiveJobs()
				scheduler.GetQueuedJobs()
				scheduler.GetAllJobs()
			}
		}()
	}

	wg.Wait()
	
	// Should complete without race conditions or panics
	assert.True(t, true)
}

// MockJobExecutor for testing
type MockJobExecutor struct {
	executeFunc func(*BackupJob) error
}

func (m *MockJobExecutor) Execute(job *BackupJob) error {
	if m.executeFunc != nil {
		return m.executeFunc(job)
	}
	return nil
}
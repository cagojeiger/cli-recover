package tui

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBackupJob(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		command string
		wantErr bool
	}{
		{
			name:    "valid job creation",
			id:      "test-job-1",
			command: "kubectl exec nginx -- tar czf backup.tar.gz /data",
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			command: "kubectl exec nginx -- tar czf backup.tar.gz /data",
			wantErr: true,
		},
		{
			name:    "empty command",
			id:      "test-job-2",
			command: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := NewBackupJob(tt.id, tt.command)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, job)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, job)
				assert.Equal(t, tt.id, job.ID)
				assert.Equal(t, tt.command, job.Command)
				assert.Equal(t, JobStatusPending, job.Status)
				assert.NotNil(t, job.ctx)
				assert.NotNil(t, job.cancel)
				assert.Empty(t, job.Output)
				assert.Zero(t, job.Progress)
			}
		})
	}
}

func TestBackupJobStatusTransitions(t *testing.T) {
	job, err := NewBackupJob("test-job", "echo test")
	assert.NoError(t, err)

	// Initial status
	assert.Equal(t, JobStatusPending, job.Status)

	// Start job
	err = job.Start()
	assert.NoError(t, err)
	assert.Equal(t, JobStatusRunning, job.Status)
	assert.False(t, job.StartTime.IsZero())

	// Cannot start already running job
	err = job.Start()
	assert.Error(t, err)

	// Complete job
	job.Complete(nil)
	assert.Equal(t, JobStatusCompleted, job.Status)
	assert.False(t, job.EndTime.IsZero())
	assert.Nil(t, job.Error)

	// Cannot complete already completed job
	job.Complete(nil)
	assert.Equal(t, JobStatusCompleted, job.Status)
}

func TestBackupJobCancel(t *testing.T) {
	job, err := NewBackupJob("test-job", "sleep 10")
	assert.NoError(t, err)

	// Cannot cancel pending job
	err = job.Cancel()
	assert.Error(t, err)

	// Start job
	err = job.Start()
	assert.NoError(t, err)

	// Cancel running job
	err = job.Cancel()
	assert.NoError(t, err)
	assert.Equal(t, JobStatusCancelled, job.Status)
	assert.False(t, job.EndTime.IsZero())

	// Context should be cancelled
	select {
	case <-job.ctx.Done():
		// Expected
	default:
		t.Error("context should be cancelled")
	}
}

func TestBackupJobProgress(t *testing.T) {
	job, err := NewBackupJob("test-job", "echo test")
	assert.NoError(t, err)

	// Start job
	err = job.Start()
	assert.NoError(t, err)

	// Update progress
	job.UpdateProgress(25, "Processing...")
	assert.Equal(t, 25, job.Progress)
	assert.Contains(t, job.Output, "Processing...")

	// Progress should be clamped to 0-100
	job.UpdateProgress(-10, "Negative")
	assert.Equal(t, 0, job.Progress)

	job.UpdateProgress(150, "Over 100")
	assert.Equal(t, 100, job.Progress)
}

func TestBackupJobOutput(t *testing.T) {
	job, err := NewBackupJob("test-job", "echo test")
	assert.NoError(t, err)

	err = job.Start()
	assert.NoError(t, err)

	// Add output lines
	job.AddOutput("Line 1")
	job.AddOutput("Line 2")
	job.AddOutput("Line 3")

	assert.Len(t, job.Output, 3)
	assert.Equal(t, "Line 1", job.Output[0])
	assert.Equal(t, "Line 2", job.Output[1])
	assert.Equal(t, "Line 3", job.Output[2])

	// Test output limit
	for i := 0; i < 1000; i++ {
		job.AddOutput("Extra line")
	}
	
	// Should keep only last N lines
	assert.LessOrEqual(t, len(job.Output), maxOutputLines)
}

func TestBackupJobTimeout(t *testing.T) {
	// Create job with short timeout
	job, err := NewBackupJobWithTimeout("test-job", "sleep 10", 100*time.Millisecond)
	assert.NoError(t, err)

	err = job.Start()
	assert.NoError(t, err)

	// Wait for timeout
	time.Sleep(200 * time.Millisecond)

	// Context should be cancelled due to timeout
	select {
	case <-job.ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, job.ctx.Err())
	default:
		t.Error("context should be cancelled due to timeout")
	}
}

func TestBackupJobConcurrentAccess(t *testing.T) {
	job, err := NewBackupJob("test-job", "echo test")
	assert.NoError(t, err)

	err = job.Start()
	assert.NoError(t, err)

	// Concurrent reads and writes
	done := make(chan bool)
	
	// Writer goroutines
	for i := 0; i < 5; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				job.AddOutput(fmt.Sprintf("Writer %d: Line %d", n, j))
				job.UpdateProgress(j, "Progress")
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = job.GetStatus()
				_ = job.GetProgress()
				_ = job.GetOutput()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Job should still be in valid state
	assert.Equal(t, JobStatusRunning, job.Status)
	assert.GreaterOrEqual(t, job.Progress, 0)
	assert.LessOrEqual(t, job.Progress, 100)
}

func TestBackupJobError(t *testing.T) {
	job, err := NewBackupJob("test-job", "false")
	assert.NoError(t, err)

	err = job.Start()
	assert.NoError(t, err)

	// Complete with error
	testErr := fmt.Errorf("backup failed: disk full")
	job.Complete(testErr)

	assert.Equal(t, JobStatusFailed, job.Status)
	assert.Equal(t, testErr, job.Error)
	assert.False(t, job.EndTime.IsZero())
}

func TestBackupJobDuration(t *testing.T) {
	job, err := NewBackupJob("test-job", "echo test")
	assert.NoError(t, err)

	// Pending job has no duration
	assert.Equal(t, time.Duration(0), job.Duration())

	// Start job
	err = job.Start()
	assert.NoError(t, err)
	
	time.Sleep(100 * time.Millisecond)
	
	// Running job has duration
	duration := job.Duration()
	assert.Greater(t, duration, time.Duration(0))
	assert.Less(t, duration, 200*time.Millisecond)

	// Complete job
	job.Complete(nil)
	finalDuration := job.Duration()
	
	// Completed job duration should be fixed
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, finalDuration, job.Duration())
}
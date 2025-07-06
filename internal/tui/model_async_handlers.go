package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// handleBackupSubmit processes backup submission
func (m Model) handleBackupSubmit(msg BackupSubmitMsg) (Model, tea.Cmd) {
	// Generate unique job ID
	jobID := fmt.Sprintf("backup-%d", time.Now().Unix())
	
	// Create new backup job
	job, err := NewBackupJob(jobID, msg.Command)
	if err != nil {
		m.err = err
		return m, nil
	}
	
	// Submit to scheduler
	err = m.jobScheduler.Submit(job)
	if err != nil {
		m.err = err
		return m, nil
	}
	
	// Set up async executor
	executor := NewAsyncBackupExecutor(nil) // We'll get the program reference later
	
	// Start job asynchronously
	return m, ExecuteBackupAsync(job, executor)
}

// handleBackupProgress updates job progress
func (m Model) handleBackupProgress(msg BackupProgressMsg) (Model, tea.Cmd) {
	// Find the job
	job := m.jobScheduler.GetJob(msg.JobID)
	if job == nil {
		return m, nil
	}
	
	// Progress is already updated by the executor
	// Just trigger a refresh if we're on the job manager screen
	if m.screen == ScreenJobManager {
		return m, nil
	}
	
	return m, nil
}

// handleBackupComplete processes job completion
func (m Model) handleBackupComplete(msg BackupCompleteMsg) (Model, tea.Cmd) {
	// The job is already marked as complete by the executor
	// Just log for now
	debugLog("Backup %s completed: success=%v, error=%v", msg.JobID, msg.Success, msg.Error)
	
	// If we're on the job manager screen, update display
	if m.screen == ScreenJobManager {
		return m, nil
	}
	
	return m, nil
}

// handleBackupCancel cancels a backup job
func (m Model) handleBackupCancel(msg BackupCancelMsg) (Model, tea.Cmd) {
	err := m.jobScheduler.Cancel(msg.JobID)
	if err != nil {
		m.err = err
	}
	
	return m, nil
}

// Helper method to check if any jobs are active
func (m Model) hasActiveJobs() bool {
	return len(m.jobScheduler.GetActiveJobs()) > 0
}

// Helper method to get job count summary
func (m Model) getJobSummary() string {
	active := len(m.jobScheduler.GetActiveJobs())
	queued := len(m.jobScheduler.GetQueuedJobs())
	
	if active == 0 && queued == 0 {
		return "No active backups"
	}
	
	summary := fmt.Sprintf("%d active", active)
	if queued > 0 {
		summary += fmt.Sprintf(", %d queued", queued)
	}
	
	return summary
}
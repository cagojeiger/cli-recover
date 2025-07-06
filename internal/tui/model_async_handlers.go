package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// handleBackupSubmit processes backup submission
func (m Model) handleBackupSubmit(msg BackupSubmitMsg) (Model, tea.Cmd) {
	// Generate unique job ID
	jobID := fmt.Sprintf("backup-%s-%d", m.selectedPod, time.Now().Unix())
	
	// Create new backup job
	job, err := NewBackupJob(jobID, msg.Command)
	if err != nil {
		m.err = err
		return m, nil
	}
	
	// Add to job manager
	err = m.jobManager.Add(job)
	if err != nil {
		m.err = err
		return m, nil
	}
	
	// Check if we can run it immediately
	if m.jobManager.CanRunMore() {
		// Execute immediately
		return m, tea.Batch(
			func() tea.Msg {
				return JobExecuteMsg{Job: job}
			},
			func() tea.Msg {
				return ScreenJobManagerMsg{}
			},
		)
	}
	
	// Job is queued, just show job manager
	return m, func() tea.Msg {
		return ScreenJobManagerMsg{}
	}
}

// handleBackupProgress updates job progress
func (m Model) handleBackupProgress(msg BackupProgressMsg) (Model, tea.Cmd) {
	// Find the job
	job := m.jobManager.Get(msg.JobID)
	if job == nil {
		return m, nil
	}
	
	// Update job progress
	job.AddOutput(msg.Output)
	if msg.Progress >= 0 {
		job.UpdateProgress(msg.Progress, msg.Output)
	}
	
	// Just trigger a refresh if we're on the job manager screen
	if m.screen == ScreenJobManager {
		return m, nil
	}
	
	return m, nil
}

// handleBackupComplete processes job completion
func (m Model) handleBackupComplete(msg BackupCompleteMsg) (Model, tea.Cmd) {
	// Mark job as complete
	err := m.jobManager.MarkCompleted(msg.JobID, msg.Error)
	if err != nil {
		debugLog("Failed to mark job as complete: %v", err)
	}
	
	debugLog("Backup %s completed: success=%v, error=%v", msg.JobID, msg.Success, msg.Error)
	
	// If the completed job was the active job, clear the selection
	if m.activeJobID == msg.JobID {
		m.activeJobID = ""
		m.jobDetailView = false
		
		// Try to select another job
		allJobs := m.jobManager.GetAll()
		if m.selected < len(allJobs) {
			m.activeJobID = allJobs[m.selected].ID
		} else if len(allJobs) > 0 {
			// Adjust selection if out of bounds
			m.selected = len(allJobs) - 1
			m.activeJobID = allJobs[m.selected].ID
		}
	}
	
	// Check if there are more jobs to run
	return m, waitForNextJobCmd(m.jobManager, m.program)
}

// handleBackupCancel cancels a backup job
func (m Model) handleBackupCancel(msg BackupCancelMsg) (Model, tea.Cmd) {
	err := m.jobManager.CancelJob(msg.JobID)
	if err != nil {
		m.err = err
	}
	
	// If the cancelled job was the active job, clear the selection
	if m.activeJobID == msg.JobID {
		m.activeJobID = ""
		m.jobDetailView = false
		
		// Try to select another job
		allJobs := m.jobManager.GetAll()
		if m.selected < len(allJobs) {
			m.activeJobID = allJobs[m.selected].ID
		} else if len(allJobs) > 0 {
			// Adjust selection if out of bounds
			m.selected = len(allJobs) - 1
			m.activeJobID = allJobs[m.selected].ID
		}
	}
	
	return m, nil
}

// handleBackupStart marks job as started
func (m Model) handleBackupStart(msg BackupStartMsg) (Model, tea.Cmd) {
	err := m.jobManager.UpdateStatus(msg.JobID, JobStatusRunning)
	if err != nil {
		debugLog("Failed to update job status: %v", err)
	}
	
	return m, nil
}

// handleBackupError handles backup errors
func (m Model) handleBackupError(msg BackupErrorMsg) (Model, tea.Cmd) {
	err := m.jobManager.MarkCompleted(msg.JobID, msg.Error)
	if err != nil {
		debugLog("Failed to mark job as failed: %v", err)
	}
	
	// If the failed job was the active job, clear the selection
	if m.activeJobID == msg.JobID {
		m.activeJobID = ""
		m.jobDetailView = false
		
		// Try to select another job
		allJobs := m.jobManager.GetAll()
		if m.selected < len(allJobs) {
			m.activeJobID = allJobs[m.selected].ID
		} else if len(allJobs) > 0 {
			// Adjust selection if out of bounds
			m.selected = len(allJobs) - 1
			m.activeJobID = allJobs[m.selected].ID
		}
	}
	
	// Check if there are more jobs to run
	return m, waitForNextJobCmd(m.jobManager, m.program)
}

// handleJobExecute starts executing a job
func (m Model) handleJobExecute(msg JobExecuteMsg) (Model, tea.Cmd) {
	job := msg.Job
	
	// Update job status to running
	err := m.jobManager.UpdateStatus(job.ID, JobStatusRunning)
	if err != nil {
		return m, func() tea.Msg {
			return BackupErrorMsg{JobID: job.ID, Error: err}
		}
	}
	
	// Start job execution
	return m, tea.Batch(
		executeBackupCmd(job, m.program),
		monitorJobCmd(job),
	)
}
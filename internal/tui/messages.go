package tui

import (
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
)

// Message types for backup job management

// BackupSubmitMsg requests submission of a new backup job
type BackupSubmitMsg struct {
	Command string
}

// BackupProgressMsg updates progress for a specific job
type BackupProgressMsg struct {
	JobID     string
	Output    string
	Percent   int
	Progress  int         // Same as Percent for compatibility
	Timestamp time.Time
}

// BackupCompleteMsg indicates a job has completed
type BackupCompleteMsg struct {
	JobID    string
	Success  bool
	Error    error
	Duration time.Duration
}

// BackupCancelMsg requests cancellation of a job
type BackupCancelMsg struct {
	JobID string
}

// JobListUpdateMsg requests update of job list display
type JobListUpdateMsg struct{}

// ScreenJobManagerMsg switches to job manager screen
type ScreenJobManagerMsg struct{}

// Additional helper messages

// BackupStartMsg indicates a backup has started
type BackupStartMsg struct {
	JobID   string
	Command string
}

// BackupQueuedMsg indicates a job was queued
type BackupQueuedMsg struct {
	JobID    string
	Position int
}

// EmergencyShutdownMsg requests emergency shutdown of all jobs
type EmergencyShutdownMsg struct{}

// NavigateBackMsg requests to go back to previous screen
type NavigateBackMsg struct{}

// JobDetailMsg requests to show job details
type JobDetailMsg struct {
	JobID string
}

// RefreshMsg requests a UI refresh
type RefreshMsg struct{}

// BackupErrorMsg is sent when a backup job encounters an error during setup
type BackupErrorMsg struct {
	JobID string
	Error error
}

// JobExecuteMsg requests execution of a backup job
type JobExecuteMsg struct {
	Job *BackupJob
}

// Helper functions for creating tea.Cmd

// SubmitBackupCmd creates a command to submit a backup job
func SubmitBackupCmd(command string) tea.Cmd {
	return func() tea.Msg {
		return BackupSubmitMsg{Command: command}
	}
}

// CancelBackupCmd creates a command to cancel a backup job
func CancelBackupCmd(jobID string) tea.Cmd {
	return func() tea.Msg {
		return BackupCancelMsg{JobID: jobID}
	}
}

// UpdateJobListCmd creates a command to update job list
func UpdateJobListCmd() tea.Cmd {
	return func() tea.Msg {
		return JobListUpdateMsg{}
	}
}
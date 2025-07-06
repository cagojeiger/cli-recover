package tui

import tea "github.com/charmbracelet/bubbletea"

// Message types for backup job management

// BackupSubmitMsg requests submission of a new backup job
type BackupSubmitMsg struct {
	Command string
}

// BackupProgressMsg updates progress for a specific job
type BackupProgressMsg struct {
	JobID   string
	Output  string
	Percent int
}

// BackupCompleteMsg indicates a job has completed
type BackupCompleteMsg struct {
	JobID   string
	Success bool
	Error   error
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
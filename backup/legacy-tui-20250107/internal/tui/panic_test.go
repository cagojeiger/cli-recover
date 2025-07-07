package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cagojeiger/cli-recover/internal/runner"
)

// TestJobManagerPanicRecovery tests that the TUI handles panics gracefully
func TestJobManagerPanicRecovery(t *testing.T) {
	// Create a model with job manager
	runner := runner.NewRunner()
	model := InitialModel(runner)
	model.jobManager = NewJobManager(3)
	
	// Add a test job
	job, err := NewBackupJob("test-job-1", "backup filesystem pod /path")
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}
	
	// Add job to manager
	err = model.jobManager.Add(job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}
	
	// Set the job as active
	model.activeJobID = job.ID
	model.selected = 0
	model.screen = ScreenJobManager
	
	// Mark job as completed
	model.jobManager.MarkCompleted(job.ID, nil)
	
	// Simulate pressing Enter to toggle detail view
	// This should not panic even though the job is completed
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")}
	
	// Test that Update doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Update panicked: %v", r)
			}
		}()
		
		updatedModel, _ := model.Update(keyMsg)
		_ = updatedModel
	}()
	
	// Test that View doesn't panic with detail view enabled
	model.jobDetailView = true
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("View panicked: %v", r)
			}
		}()
		
		view := model.View()
		if view == "" {
			t.Error("View returned empty string")
		}
	}()
}

// TestJobManagerBoundsChecking tests array bounds handling
func TestJobManagerBoundsChecking(t *testing.T) {
	// Create a model with job manager
	runner := runner.NewRunner()
	model := InitialModel(runner)
	model.jobManager = NewJobManager(3)
	
	// Add multiple jobs
	for i := 0; i < 3; i++ {
		job, _ := NewBackupJob(
			fmt.Sprintf("test-job-%d", i),
			fmt.Sprintf("backup filesystem pod%d /path", i),
		)
		model.jobManager.Add(job)
	}
	
	// Set selection to last job
	allJobs := model.jobManager.GetAll()
	model.selected = len(allJobs) - 1
	model.activeJobID = allJobs[model.selected].ID
	model.screen = ScreenJobManager
	
	// Complete all jobs
	for _, job := range allJobs {
		model.jobManager.MarkCompleted(job.ID, nil)
	}
	
	// Handle completion message for the active job
	completeMsg := BackupCompleteMsg{
		JobID:   model.activeJobID,
		Success: true,
	}
	
	// This should not panic and should handle bounds correctly
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("handleBackupComplete panicked: %v", r)
			}
		}()
		
		updatedModel, _ := model.handleBackupComplete(completeMsg)
		
		// Check that selection is adjusted
		if updatedModel.selected >= len(updatedModel.jobManager.GetAll()) {
			t.Error("Selection not adjusted after job completion")
		}
	}()
}

// TestJobDetailViewValidation tests job detail view validation
func TestJobDetailViewValidation(t *testing.T) {
	// Create a model with job manager
	runner := runner.NewRunner()
	model := InitialModel(runner)
	model.jobManager = NewJobManager(3)
	model.screen = ScreenJobManager
	
	// Set a non-existent job ID
	model.activeJobID = "non-existent-job"
	model.selected = 0
	
	// Try to toggle detail view with Enter key
	model = handleJobManagerKey(model, "enter")
	
	// Should not enable detail view for non-existent job
	if model.jobDetailView {
		t.Error("Detail view enabled for non-existent job")
	}
	
	// activeJobID should be cleared
	if model.activeJobID == "non-existent-job" {
		t.Error("activeJobID not cleared for non-existent job")
	}
}
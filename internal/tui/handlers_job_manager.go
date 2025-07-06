package tui

import ()

// handleJobManagerKey handles key presses in the job manager screen
func handleJobManagerKey(m Model, key string) Model {
	// Get all jobs to count total
	allJobs := m.jobManager.GetAll()
	totalJobs := len(allJobs)
	
	switch key {
	case "j", "down":
		if m.selected < totalJobs-1 {
			m.selected++
			// Update active job ID
			if m.selected < len(allJobs) {
				m.activeJobID = allJobs[m.selected].ID
			}
		}
		
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			// Update active job ID
			if m.selected < len(allJobs) {
				m.activeJobID = allJobs[m.selected].ID
			}
		}
		
	case "enter":
		// Toggle job detail view
		m.jobDetailView = !m.jobDetailView
		return m
		
	case "c":
		// Cancel selected job
		if m.activeJobID != "" {
			m.jobManager.CancelJob(m.activeJobID)
		}
		
	case "K":
		// Cancel all jobs
		m.jobManager.CancelAll()
		
	case "r":
		// Refresh (just return, view will update)
		return m
		
	case "b", "esc":
		// Back to previous screen
		m = m.popScreen()
		m.activeJobID = ""
		
	case "q":
		// Quit with confirmation if jobs are active
		activeCount := len(m.jobManager.GetActive())
		if activeCount > 0 {
			// TODO: Add confirmation dialog
			m.quit = true
		} else {
			m.quit = true
		}
	}
	
	return m
}

// Helper to show job manager from any screen
func showJobManager(m Model) Model {
	m = m.pushScreen(ScreenJobManager)
	
	// Select first job if any
	allJobs := m.jobManager.GetAll()
	if len(allJobs) > 0 {
		m.activeJobID = allJobs[0].ID
		m.selected = 0
	}
	
	return m
}
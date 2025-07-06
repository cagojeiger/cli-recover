package tui

import ()

// handleJobManagerKey handles key presses in the job manager screen
func handleJobManagerKey(m Model, key string) Model {
	activeJobs := m.jobManager.GetActive()
	queuedJobs := m.jobManager.GetQueued()
	totalJobs := len(activeJobs) + len(queuedJobs)
	
	switch key {
	case "j", "down":
		if m.selected < totalJobs-1 {
			m.selected++
			// Update active job ID
			if m.selected < len(activeJobs) {
				m.activeJobID = activeJobs[m.selected].ID
			} else {
				idx := m.selected - len(activeJobs)
				if idx < len(queuedJobs) {
					m.activeJobID = queuedJobs[idx].ID
				}
			}
		}
		
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			// Update active job ID
			if m.selected < len(activeJobs) {
				m.activeJobID = activeJobs[m.selected].ID
			} else {
				idx := m.selected - len(activeJobs)
				if idx < len(queuedJobs) {
					m.activeJobID = queuedJobs[idx].ID
				}
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
		if len(activeJobs) > 0 {
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
	
	// Select first active job if any
	activeJobs := m.jobManager.GetActive()
	if len(activeJobs) > 0 {
		m.activeJobID = activeJobs[0].ID
	}
	
	return m
}
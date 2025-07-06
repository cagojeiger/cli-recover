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
			// Update active job ID with bounds check
			if m.selected < len(allJobs) {
				m.activeJobID = allJobs[m.selected].ID
			} else {
				// Safety: clear activeJobID if out of bounds
				m.activeJobID = ""
				debugLog("WARNING: selected index %d out of bounds (total: %d)", m.selected, len(allJobs))
			}
		}
		
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			// Update active job ID with bounds check
			if m.selected < len(allJobs) && m.selected >= 0 {
				m.activeJobID = allJobs[m.selected].ID
			} else {
				// Safety: clear activeJobID if out of bounds
				m.activeJobID = ""
				debugLog("WARNING: selected index %d out of bounds (total: %d)", m.selected, len(allJobs))
			}
		}
		
	case "enter":
		// Toggle job detail view, but only if we have a valid job selected
		if m.activeJobID != "" {
			// Verify the job still exists before toggling detail view
			if job := m.jobManager.Get(m.activeJobID); job != nil {
				m.jobDetailView = !m.jobDetailView
			} else {
				// Job no longer exists, clear the activeJobID
				m.activeJobID = ""
				// Try to set a new active job if possible
				allJobs := m.jobManager.GetAll()
				if m.selected < len(allJobs) {
					m.activeJobID = allJobs[m.selected].ID
				}
			}
		}
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
		// If activeJobID exists in the list, find its index
		found := false
		for i, job := range allJobs {
			if job.ID == m.activeJobID {
				m.selected = i
				found = true
				break
			}
		}
		
		// If not found or no activeJobID, select the first job
		if !found {
			m.activeJobID = allJobs[0].ID
			m.selected = 0
		}
	} else {
		// No jobs, clear selection
		m.activeJobID = ""
		m.selected = 0
	}
	
	return m
}
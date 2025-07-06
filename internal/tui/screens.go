package tui

import (
	"fmt"
	"time"
	
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)


// viewMainMenu renders the main menu
func viewMainMenu(m Model, width int) string {
	items := []string{"Backup", "Restore", "Exit"}
	
	var view string
	view += "Main Menu:\n"
	
	// Menu items
	for i, item := range items {
		if i == m.selected {
			view += fmt.Sprintf("  > %s\n", item)
		} else {
			view += fmt.Sprintf("    %s\n", item)
		}
	}
	
	return view
}

// viewNamespaceList renders namespace selection
func viewNamespaceList(m Model, width int) string {
	var view string
	view += "Select Namespace:\n"
	
	// Namespace list
	for i, ns := range m.namespaces {
		if i == m.selected {
			view += fmt.Sprintf("  > %s\n", ns)
		} else {
			view += fmt.Sprintf("    %s\n", ns)
		}
	}
	
	return view
}

// viewPodList renders pod selection
func viewPodList(m Model, width int) string {
	var view string
	view += fmt.Sprintf("Pods in %s:\n", m.selectedNamespace)
	
	// Pod list
	for i, pod := range m.pods {
		// Simple pod display
		display := fmt.Sprintf("%-30s %s %s", pod.Name, pod.Status, pod.Ready)
		
		if i == m.selected {
			view += fmt.Sprintf("  > %s\n", display)
		} else {
			view += fmt.Sprintf("    %s\n", display)
		}
	}
	
	return view
}

// viewContainerList renders container selection for multi-container pods
func viewContainerList(m Model, width int) string {
	var view string
	
	// Find the selected pod to get its containers
	var selectedPod kubernetes.Pod
	for _, pod := range m.pods {
		if pod.Name == m.selectedPod {
			selectedPod = pod
			break
		}
	}
	
	view += fmt.Sprintf("Containers in %s:\n", m.selectedPod)
	
	// Container list
	for i, container := range selectedPod.Containers {
		if i == m.selected {
			view += fmt.Sprintf("  > %s\n", container)
		} else {
			view += fmt.Sprintf("    %s\n", container)
		}
	}
	
	return view
}

// viewDirectoryBrowser renders directory browsing screen
func viewDirectoryBrowser(m Model, width int) string {
	var view string
	view += fmt.Sprintf("Browse: %s\n", m.currentPath)
	
	// Directory entries
	for i, entry := range m.directories {
		icon := "📄"
		if entry.Type == "dir" {
			icon = "📁"
		}
		
		// Simple entry display
		display := fmt.Sprintf("%s %-30s %s", icon, entry.Name, entry.Size)
		
		if i == m.selected {
			view += fmt.Sprintf("  > %s\n", display)
		} else {
			view += fmt.Sprintf("    %s\n", display)
		}
	}
	
	return view
}

// viewBackupType renders the backup type selection screen
func viewBackupType(m Model, width int) string {
	backupTypes := []struct {
		name        string
		description string
	}{
		{"filesystem", "Backup files and directories from pod filesystem"},
	}
	
	var view string
	view += "Select Backup Type:\n"
	
	// Backup type options
	for i, bt := range backupTypes {
		if i == m.selected {
			view += fmt.Sprintf("  > %-12s - %s\n", bt.name, bt.description)
		} else {
			view += fmt.Sprintf("    %-12s - %s\n", bt.name, bt.description)
		}
	}
	
	return view
}

// viewExecuting renders the backup execution screen
func viewExecuting(m Model, width int) string {
	var view string
	view += "Backup Progress:\n\n"
	
	// Show last N lines of output
	for _, line := range m.executeOutput {
		view += line + "\n"
	}
	
	return view
}

// viewJobManager renders the job manager screen
func viewJobManager(m Model, width int) string {
	var view string
	view += "=== Backup Job Manager ===\n\n"
	
	// Check if showing job detail
	if m.jobDetailView && m.activeJobID != "" {
		return viewJobDetail(m, width)
	}
	
	// Get all jobs and categorize by status
	allJobs := m.jobManager.GetAll()
	var activeJobs, queuedJobs, recentJobs []*BackupJob
	
	for _, job := range allJobs {
		switch job.Status {
		case JobStatusRunning:
			activeJobs = append(activeJobs, job)
		case JobStatusQueued, JobStatusPending:
			queuedJobs = append(queuedJobs, job)
		case JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
			recentJobs = append(recentJobs, job)
		}
	}
	
	// Sort recent jobs by end time (newest first)
	// Limit to last 10 recent jobs
	if len(recentJobs) > 10 {
		recentJobs = recentJobs[len(recentJobs)-10:]
	}
	
	// Job index for navigation
	jobIndex := 0
	
	// Active jobs section
	view += fmt.Sprintf("Active Jobs (%d/%d):\n", len(activeJobs), m.jobManager.GetMaxJobs())
	view += "─────────────────────────────────────────\n"
	
	if len(activeJobs) == 0 {
		view += "  No active jobs\n"
	} else {
		for _, job := range activeJobs {
			selected := m.selected == jobIndex && m.activeJobID == job.ID
			marker := "  "
			if selected {
				marker = "> "
			}
			
			progress := job.GetProgress()
			duration := job.Duration()
			
			view += fmt.Sprintf("%s🔄 [%s] Running (%d%%) - %s\n", 
				marker, job.ID, progress, duration.Round(time.Second))
			
			// Show last output line if selected
			if selected && len(job.GetOutput()) > 0 {
				output := job.GetOutput()
				lastLine := output[len(output)-1]
				if len(lastLine) > width-6 {
					lastLine = lastLine[:width-9] + "..."
				}
				view += fmt.Sprintf("     └─ %s\n", lastLine)
			}
			jobIndex++
		}
	}
	
	view += "\n"
	
	// Queued jobs section
	if len(queuedJobs) > 0 {
		view += fmt.Sprintf("Queued Jobs (%d):\n", len(queuedJobs))
		view += "─────────────────────────────────────────\n"
		
		for _, job := range queuedJobs {
			selected := m.selected == jobIndex
			marker := "  "
			if selected {
				marker = "> "
				m.activeJobID = job.ID
			}
			cmdPreview := job.Command
			if len(cmdPreview) > width-30 {
				cmdPreview = cmdPreview[:width-33] + "..."
			}
			view += fmt.Sprintf("%s⏳ [%s] Waiting - %s\n", marker, job.ID, cmdPreview)
			jobIndex++
		}
		view += "\n"
	}
	
	// Recent jobs section (completed/failed/cancelled)
	if len(recentJobs) > 0 {
		view += fmt.Sprintf("Recent Jobs (%d):\n", len(recentJobs))
		view += "─────────────────────────────────────────\n"
		
		for _, job := range recentJobs {
			selected := m.selected == jobIndex
			marker := "  "
			if selected {
				marker = "> "
				m.activeJobID = job.ID
			}
			
			// Status icon
			statusIcon := "❓"
			statusText := string(job.Status)
			switch job.Status {
			case JobStatusCompleted:
				statusIcon = "✅"
				statusText = "Completed"
			case JobStatusFailed:
				statusIcon = "❌"
				statusText = "Failed"
			case JobStatusCancelled:
				statusIcon = "🚫"
				statusText = "Cancelled"
			}
			
			duration := job.Duration()
			view += fmt.Sprintf("%s%s [%s] %s - %s\n", 
				marker, statusIcon, job.ID, statusText, duration.Round(time.Second))
			
			// Show error if failed and selected
			if selected && job.Error != nil {
				errMsg := job.Error.Error()
				if len(errMsg) > width-10 {
					errMsg = errMsg[:width-13] + "..."
				}
				view += fmt.Sprintf("     └─ Error: %s\n", errMsg)
			}
			jobIndex++
		}
		view += "\n"
	}
	
	// Controls
	view += "\nControls:\n"
	view += "  [↑/↓] Navigate jobs\n"
	view += "  [Enter] View job details\n"
	view += "  [c] Cancel selected job\n"
	view += "  [K] Cancel ALL jobs\n"
	view += "  [r] Refresh\n"
	view += "  [b/Esc] Back to main menu\n"
	
	return view
}

// viewJobDetail renders detailed job information
func viewJobDetail(m Model, width int) string {
	job := m.jobManager.Get(m.activeJobID)
	if job == nil {
		return "Job not found\n\n[Enter] Back to list"
	}
	
	var view string
	view += fmt.Sprintf("=== Job Details: %s ===\n\n", job.ID)
	
	// Job info
	view += fmt.Sprintf("Command: %s\n", job.Command)
	view += fmt.Sprintf("Status: %s\n", job.GetStatus())
	view += fmt.Sprintf("Progress: %d%%\n", job.GetProgress())
	view += fmt.Sprintf("Duration: %s\n", job.Duration())
	
	if job.Error != nil {
		view += fmt.Sprintf("Error: %v\n", job.Error)
	}
	
	view += "\n"
	view += "Output:\n"
	view += "─────────────────────────────────────────\n"
	
	// Show last N lines of output
	output := job.GetOutput()
	maxLines := 20
	startIdx := 0
	if len(output) > maxLines {
		startIdx = len(output) - maxLines
	}
	
	for i := startIdx; i < len(output); i++ {
		line := output[i]
		if len(line) > width-2 {
			line = line[:width-5] + "..."
		}
		view += line + "\n"
	}
	
	view += "\n"
	view += "[Enter] Back to list  [c] Cancel job  [b/Esc] Exit job manager\n"
	
	return view
}
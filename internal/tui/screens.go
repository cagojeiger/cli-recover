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
	
	// Get all jobs
	activeJobs := m.jobScheduler.GetActiveJobs()
	queuedJobs := m.jobScheduler.GetQueuedJobs()
	
	// Active jobs section
	view += fmt.Sprintf("Active Jobs (%d/%d):\n", len(activeJobs), m.jobScheduler.GetMaxJobs())
	view += "─────────────────────────────────────────\n"
	
	if len(activeJobs) == 0 {
		view += "  No active jobs\n"
	} else {
		for i, job := range activeJobs {
			selected := m.selected == i && m.activeJobID == job.ID
			marker := "  "
			if selected {
				marker = "> "
			}
			
			status := job.GetStatus()
			progress := job.GetProgress()
			duration := job.Duration()
			
			view += fmt.Sprintf("%s[%s] %s (%d%%) - %s\n", 
				marker, job.ID, status, progress, duration.Round(time.Second))
			
			// Show last output line if selected
			if selected && len(job.GetOutput()) > 0 {
				output := job.GetOutput()
				lastLine := output[len(output)-1]
				if len(lastLine) > width-6 {
					lastLine = lastLine[:width-9] + "..."
				}
				view += fmt.Sprintf("     └─ %s\n", lastLine)
			}
		}
	}
	
	view += "\n"
	
	// Queued jobs section
	if len(queuedJobs) > 0 {
		view += fmt.Sprintf("Queued Jobs (%d):\n", len(queuedJobs))
		view += "─────────────────────────────────────────\n"
		
		for i, job := range queuedJobs {
			marker := "  "
			if m.selected == len(activeJobs)+i {
				marker = "> "
			}
			view += fmt.Sprintf("%s[%s] %s (waiting)\n", marker, job.ID, job.Command)
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
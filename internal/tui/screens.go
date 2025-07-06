package tui

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)


// viewMainMenu renders the main menu
func viewMainMenu(m Model, width int) string {
	items := []struct {
		name string
		desc string
	}{
		{"Backup", "Create backups from Kubernetes pods"},
		{"Restore", "Restore backups to Kubernetes pods (Coming soon)"},
		{"Exit", "Exit CLI Recover"},
	}
	
	var view string
	view += "Main Menu:\n\n"
	
	// Menu items with descriptions for wider screens
	for i, item := range items {
		marker := "  "
		if i == m.selected {
			marker = "> "
		}
		
		if width < 80 {
			// Simple display
			view += fmt.Sprintf("%s%s\n", marker, item.name)
		} else {
			// Extended display with descriptions
			view += fmt.Sprintf("%s%-10s - %s\n", marker, item.name, item.desc)
		}
	}
	
	// Add job summary if we have jobs
	if m.jobManager != nil {
		active := len(m.jobManager.GetActive())
		queued := len(m.jobManager.GetQueued())
		if active > 0 || queued > 0 {
			view += fmt.Sprintf("\n─────────────────────────────\n")
			view += fmt.Sprintf("Active Jobs: %d | Queued: %d\n", active, queued)
			if active > 0 {
				view += "Press [J] to view Job Manager\n"
			}
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
	
	// Pod list with dynamic info based on width
	for i, pod := range m.pods {
		var display string
		
		if width < 80 {
			// Compact display
			display = fmt.Sprintf("%-30s %s", pod.Name, pod.Status)
		} else {
			// Extended display with more info
			// Extract container count from pod.Containers
			containerCount := len(pod.Containers)
			containerInfo := "1 container"
			if containerCount > 1 {
				containerInfo = fmt.Sprintf("%d containers", containerCount)
			}
			
			// Format with additional information
			display = fmt.Sprintf("%-40s %-8s %-6s %s", 
				pod.Name, pod.Status, pod.Ready, containerInfo)
		}
		
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
	
	// Show more info for wider screens
	if width >= 100 && m.selectedPath != "" {
		view += fmt.Sprintf("Selected: %s\n", m.selectedPath)
	}
	view += "\n"
	
	// Directory entries with dynamic display
	for i, entry := range m.directories {
		icon := "📄"
		if entry.Type == "dir" {
			icon = "📁"
		}
		
		var display string
		if width < 80 {
			// Compact: just name and size
			display = fmt.Sprintf("%s %-30s %s", icon, entry.Name, entry.Size)
		} else {
			// Extended: add type and permissions if available
			typeStr := "file"
			if entry.Type == "dir" {
				typeStr = "dir "
			}
			
			// Add modification info if we have width
			if width >= 100 {
				display = fmt.Sprintf("%s %-40s %4s %10s", 
					icon, entry.Name, typeStr, entry.Size)
			} else {
				display = fmt.Sprintf("%s %-35s %s %s", 
					icon, entry.Name, typeStr, entry.Size)
			}
		}
		
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
	// Limit based on terminal height to show more jobs
	maxRecentJobs := 10
	if m.height > 30 {
		// Show more jobs on larger terminals
		maxRecentJobs = (m.height - 20) / 2  // Reserve space for headers and controls
	}
	if len(recentJobs) > maxRecentJobs {
		recentJobs = recentJobs[len(recentJobs)-maxRecentJobs:]
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
			
			// Build progress bar
			barWidth := 20
			if width >= 100 {
				barWidth = 30
			}
			progressBar := makeProgressBar(progress, barWidth)
			
			if width < 80 {
				// Compact display
				view += fmt.Sprintf("%s%s %d%% %s\n", 
					marker, job.ID[:8], progress, duration.Round(time.Second))
			} else {
				// Extended display
				view += fmt.Sprintf("%s🔄 [%s] %s %d%% - %s\n", 
					marker, job.ID[:16], progressBar, progress, duration.Round(time.Second))
				
				// Show last output line if selected
				if selected && len(job.GetOutput()) > 0 {
					output := job.GetOutput()
					lastLine := output[len(output)-1]
					maxLen := width - 10
					if len(lastLine) > maxLen {
						lastLine = lastLine[:maxLen-3] + "..."
					}
					view += fmt.Sprintf("     └─ %s\n", lastLine)
				}
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
			selected := m.selected == jobIndex && m.activeJobID == job.ID
			marker := "  "
			if selected {
				marker = "> "
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
			selected := m.selected == jobIndex && m.activeJobID == job.ID
			marker := "  "
			if selected {
				marker = "> "
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
			
			if width < 80 {
				// Compact display
				view += fmt.Sprintf("%s%s %s %s\n", 
					marker, statusIcon, job.ID[:8], duration.Round(time.Second))
			} else {
				// Extended display with more info
				// Extract pod name from command
				cmdParts := strings.Fields(job.Command)
				podName := "unknown"
				if len(cmdParts) >= 3 {
					podName = cmdParts[2]
					if len(podName) > 20 {
						podName = podName[:17] + "..."
					}
				}
				
				view += fmt.Sprintf("%s%s [%s] %s - %s - %s\n", 
					marker, statusIcon, job.ID[:16], statusText, 
					duration.Round(time.Second), podName)
				
				// Show error if failed and selected
				if selected && job.Error != nil {
					errMsg := job.Error.Error()
					maxLen := width - 15
					if len(errMsg) > maxLen {
						errMsg = errMsg[:maxLen-3] + "..."
					}
					view += fmt.Sprintf("     └─ Error: %s\n", errMsg)
				}
			}
			jobIndex++
		}
		view += "\n"
	}
	
	// Show current position if we have jobs
	totalJobsDisplayed := len(activeJobs) + len(queuedJobs) + len(recentJobs)
	if totalJobsDisplayed > 0 {
		// Calculate which job is selected across all sections
		currentPos := m.selected + 1
		view += fmt.Sprintf("\nPosition: %d/%d jobs", currentPos, totalJobsDisplayed)
		
		// Show total jobs in system if different from displayed
		totalInSystem := len(allJobs)
		if totalInSystem > totalJobsDisplayed {
			view += fmt.Sprintf(" (%d total in system)", totalInSystem)
		}
		view += "\n"
	}
	
	// Controls - make more compact for small screens
	view += "\nControls: "
	if width < 80 {
		view += "[↑/↓] Nav [Enter] Details [c] Cancel [b] Back"
	} else {
		view += "[↑/↓] Navigate [Enter] Details [c] Cancel Job [K] Cancel All [r] Refresh [b] Back"
	}
	view += "\n"
	
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
	
	// Show output - use available terminal height
	output := job.GetOutput()
	// Reserve lines for header (10) and footer (3)
	availableLines := m.height - 13
	if availableLines < 10 {
		availableLines = 10
	}
	
	// Show line count info if output is truncated
	if len(output) > availableLines {
		view += fmt.Sprintf("(Showing last %d of %d lines)\n", availableLines, len(output))
		view += "─────────────────────────────────────────\n"
	}
	
	startIdx := 0
	if len(output) > availableLines {
		startIdx = len(output) - availableLines
	}
	
	for i := startIdx; i < len(output); i++ {
		line := output[i]
		if len(line) > width-2 {
			line = line[:width-5] + "..."
		}
		view += line + "\n"
	}
	
	// Footer
	view += "\n"
	view += "[Enter] Back to list  [c] Cancel job  [h] Home  [b/Esc] Exit job manager\n"
	
	return view
}

// makeProgressBar creates a visual progress bar
func makeProgressBar(progress int, width int) string {
	if width < 10 {
		width = 10
	}
	
	filled := (progress * width) / 100
	empty := width - filled
	
	bar := "["
	bar += strings.Repeat("█", filled)
	bar += strings.Repeat("░", empty)
	bar += "]"
	
	return bar
}
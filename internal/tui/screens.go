package tui

import (
	"fmt"
	
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
package tui

import (
	"fmt"
	"strings"
)

// viewBackupOptions renders backup options configuration screen
func viewBackupOptions(m Model, width int) string {
	// Only filesystem backup is supported now
	return viewFilesystemOptions(m, width)
}

// viewPathInput renders path input screen
func viewPathInput(m Model, width int) string {
	var view string
	
	view += "=== Backup Configuration ===\n\n"
	
	// Group related information together
	if width < 80 {
		// Compact view
		view += fmt.Sprintf("Target: %s/%s\n", m.selectedNamespace, m.selectedPod)
		view += fmt.Sprintf("Path:   %s\n", m.selectedPath)
	} else {
		// Extended view with more details
		view += fmt.Sprintf("Namespace: %s\n", m.selectedNamespace)
		view += fmt.Sprintf("Pod:       %s\n", m.selectedPod)
		if m.selectedContainer != "" {
			view += fmt.Sprintf("Container: %s\n", m.selectedContainer)
		}
		view += fmt.Sprintf("Path:      %s\n", m.selectedPath)
	}
	
	// Show output file location
	outputFile := m.backupOptions.OutputFile
	if outputFile == "" {
		// Generate default filename if not specified
		pathSuffix := m.selectedPath
		if pathSuffix == "/" {
			pathSuffix = "root"
		} else {
			pathSuffix = strings.TrimPrefix(pathSuffix, "/")
			pathSuffix = strings.ReplaceAll(pathSuffix, "/", "-")
			pathSuffix = strings.ReplaceAll(pathSuffix, " ", "-")
			pathSuffix = strings.ReplaceAll(pathSuffix, ".", "-")
		}
		outputFile = fmt.Sprintf("backup-%s-%s-%s.tar.gz", m.selectedNamespace, m.selectedPod, pathSuffix)
	}
	view += fmt.Sprintf("Output:    %s\n", outputFile)
	
	// Add compression info
	if width >= 80 {
		view += fmt.Sprintf("\nCompression: %s", m.backupOptions.CompressionType)
		if len(m.backupOptions.ExcludePatterns) > 0 {
			view += fmt.Sprintf(" | Excludes: %s", strings.Join(m.backupOptions.ExcludePatterns, ", "))
		}
		view += "\n"
	}
	
	view += "\n"
	view += "Command to execute:\n"
	view += fmt.Sprintf("$ %s\n", m.commandBuilder.Preview())
	
	// Add recent backups if space allows
	if width >= 100 {
		// TODO: Would need to access recent backups from job manager
		// For now, just show a placeholder structure
		view += "\nRecent backups from this pod:\n"
		view += "• No recent backups found\n"
	}
	
	return view
}
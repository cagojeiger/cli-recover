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
	
	view += "Backup Configuration:\n\n"
	
	// Configuration details
	pathLabel := "Path"
	
	view += fmt.Sprintf("Namespace: %s\n", m.selectedNamespace)
	view += fmt.Sprintf("Pod: %s\n", m.selectedPod)
	view += fmt.Sprintf("%s: %s\n", pathLabel, m.selectedPath)
	
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
	view += fmt.Sprintf("Output: %s\n", outputFile)
	
	view += "\n---\n"
	view += "Command to execute:\n"
	view += fmt.Sprintf("$ %s\n", m.commandBuilder.Preview())
	
	return view
}
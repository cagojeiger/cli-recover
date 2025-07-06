package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)

func handleSpace(m Model) Model {
	if m.screen == ScreenDirectoryBrowser {
		m.selectedPath = m.currentPath
		
		// Update command builder with selected path
		m.commandBuilder.SetPath(m.selectedPath)
		
		m.screen = ScreenBackupOptions
		m.optionCategory = 0
		m.optionSelected = 0
		m.selected = 0
		
		// Set default output filename based on selected path
		m = setDefaultOutputFilename(m)
		
		return m
	}
	if m.screen == ScreenBackupOptions {
		return handleOptionToggle(m)
	}
	return m
}

// handleOptionToggle toggles backup options
func handleOptionToggle(m Model) Model {
	// Only filesystem backup is supported now
	return handleFilesystemOptionToggle(m)
}

// handleFilesystemOptionToggle handles filesystem backup option toggles
func handleFilesystemOptionToggle(m Model) Model {
	switch m.optionCategory {
	case 0: // Compression
		compressionTypes := []string{"gzip", "bzip2", "xz", "none"}
		if m.optionSelected < len(compressionTypes) {
			m.backupOptions.CompressionType = compressionTypes[m.optionSelected]
		}
		
	case 1: // Excludes
		excludeOptions := []string{"*.log", "tmp/*", ".git", "node_modules/*", "*.tmp"}
		if m.optionSelected < len(excludeOptions) {
			pattern := excludeOptions[m.optionSelected]
			// Toggle pattern in exclude list
			found := false
			for i, existing := range m.backupOptions.ExcludePatterns {
				if existing == pattern {
					// Remove pattern
					m.backupOptions.ExcludePatterns = append(
						m.backupOptions.ExcludePatterns[:i],
						m.backupOptions.ExcludePatterns[i+1:]...,
					)
					found = true
					break
				}
			}
			if !found {
				// Add pattern
				m.backupOptions.ExcludePatterns = append(m.backupOptions.ExcludePatterns, pattern)
			}
		} else if m.optionSelected == len(excludeOptions) {
			// Toggle VCS exclusion
			m.backupOptions.ExcludeVCS = !m.backupOptions.ExcludeVCS
		}
		
	case 2: // Advanced
		switch m.optionSelected {
		case 0:
			m.backupOptions.Verbose = !m.backupOptions.Verbose
		case 1:
			m.backupOptions.ShowTotals = !m.backupOptions.ShowTotals
		case 2:
			m.backupOptions.PreservePerms = !m.backupOptions.PreservePerms
		}
		
	case 3: // Output
		switch m.optionSelected {
		case 0:
			// Output file input
			return startTextInput(m, "output", m.backupOptions.OutputFile)
		// case 1: removed dry-run as requested
		}
	}
	
	// Update command builder with new options
	m.commandBuilder.SetOptions(m.backupOptions)
	
	return m
}

// handleEnter processes enter key based on current screen
func handleEnter(m Model) Model {
	switch m.screen {
	case ScreenMain:
		return handleMainMenuEnter(m)
	case ScreenBackupType:
		return handleBackupTypeEnter(m)
	case ScreenNamespaceList:
		return handleNamespaceEnter(m)
	case ScreenPodList:
		return handlePodEnter(m)
	case ScreenContainerList:
		return handleContainerEnter(m)
	case ScreenDirectoryBrowser:
		return handleDirectoryEnter(m)
	case ScreenBackupOptions:
		return handleBackupOptionsEnter(m)
	case ScreenPathInput:
		return handlePathInputEnter(m)
	}
	
	return m
}

func handleMainMenuEnter(m Model) Model {
	debugLog("handleMainMenuEnter: selected=%d", m.selected)
	
	switch m.selected {
	case 0: // Backup
		debugLog("Starting backup flow - selecting backup type")
		m.screen = ScreenBackupType
		m.selected = 0
		
	case 1: // Restore
		debugLog("Restore selected - not implemented")
		m.err = fmt.Errorf("restore not implemented yet")
		
	case 2: // Exit
		debugLog("Exit selected")
		m.quit = true
	}
	return m
}

func handleNamespaceEnter(m Model) Model {
	m.selectedNamespace = m.namespaces[m.selected]
	
	// Update command builder
	m.commandBuilder.SetNamespace(m.selectedNamespace)
	
	pods, err := kubernetes.GetPods(m.runner, m.selectedNamespace)
	if err != nil {
		m.err = err
		return m
	}
	m.pods = pods
	m.screen = ScreenPodList
	m.selected = 0
	return m
}

func handlePodEnter(m Model) Model {
	selectedPod := m.pods[m.selected]
	m.selectedPod = selectedPod.Name
	
	// Update command builder
	m.commandBuilder.SetPod(m.selectedPod)
	
	// Check if pod has multiple containers
	if len(selectedPod.Containers) > 1 {
		// Multi-container pod: show container selection
		m.screen = ScreenContainerList
		m.selected = 0
		return m
	} else if len(selectedPod.Containers) == 1 {
		// Single container pod: automatically select the container
		m.selectedContainer = selectedPod.Containers[0]
	}
	
	// Proceed to directory browsing
	directories, err := kubernetes.GetDirectoryContents(m.runner, m.selectedPod, m.selectedNamespace, "/", m.selectedContainer)
	if err != nil {
		m.err = err
		return m
	}
	m.currentPath = "/"
	m.directories = directories
	m.screen = ScreenDirectoryBrowser
	m.selected = 0
	return m
}

func handleContainerEnter(m Model) Model {
	// Find the selected pod to get its containers
	var selectedPod kubernetes.Pod
	for _, pod := range m.pods {
		if pod.Name == m.selectedPod {
			selectedPod = pod
			break
		}
	}
	
	if m.selected < len(selectedPod.Containers) {
		m.selectedContainer = selectedPod.Containers[m.selected]
		
		// Proceed to directory browsing
		directories, err := kubernetes.GetDirectoryContents(m.runner, m.selectedPod, m.selectedNamespace, "/", m.selectedContainer)
		if err != nil {
			m.err = err
			return m
		}
		m.currentPath = "/"
		m.directories = directories
		m.screen = ScreenDirectoryBrowser
		m.selected = 0
	}
	
	return m
}

func handleDirectoryEnter(m Model) Model {
	if m.selected >= len(m.directories) {
		return m
	}
	
	entry := m.directories[m.selected]
	if entry.Type == "dir" {
		// Navigate into directory
		newPath := entry.Name
		if m.currentPath != "/" {
			newPath = m.currentPath + "/" + entry.Name
		} else {
			newPath = "/" + entry.Name
		}
		
		directories, err := kubernetes.GetDirectoryContents(m.runner, m.selectedPod, m.selectedNamespace, newPath, m.selectedContainer)
		if err != nil {
			m.err = err
			return m
		}
		
		m.currentPath = newPath
		m.directories = directories
		m.selected = 0
	} else {
		// File selected, treat as path selection
		filePath := entry.Name
		if m.currentPath != "/" {
			filePath = m.currentPath + "/" + entry.Name
		} else {
			filePath = "/" + entry.Name
		}
		m.selectedPath = filePath
		
		// Update command builder
		m.commandBuilder.SetPath(m.selectedPath)
		
		m.screen = ScreenBackupOptions
		m.optionCategory = 0
		m.optionSelected = 0
		m.selected = 0
		
		// Set default output filename based on selected path
		m = setDefaultOutputFilename(m)
	}
	return m
}

func handleBackupTypeEnter(m Model) Model {
	debugLog("handleBackupTypeEnter: selected=%d", m.selected)
	
	backupTypes := []string{"filesystem"}
	if m.selected < len(backupTypes) {
		m.selectedBackupType = backupTypes[m.selected]
		debugLog("Selected backup type: %s", m.selectedBackupType)
		
		// Update command builder with backup type
		m.commandBuilder.SetBackupType(m.selectedBackupType)
		
		// Get namespaces and move to namespace selection
		namespaces, err := kubernetes.GetNamespaces(m.runner)
		if err != nil {
			m.err = err
			return m
		}
		m.namespaces = namespaces
		m.screen = ScreenNamespaceList
		m.selected = 0
	}
	
	return m
}

func handleBackupOptionsEnter(m Model) Model {
	// Proceed to path input for filesystem backup
	m.screen = ScreenPathInput
	m.selected = 0
	return m
}

func handlePathInputEnter(m Model) Model {
	debugLog("handlePathInputEnter: executing backup")
	debugLog("  namespace: %s, pod: %s, path: %s", m.selectedNamespace, m.selectedPod, m.selectedPath)
	debugLog("  options: %+v", m.backupOptions)
	
	// Get command from CommandBuilder
	args := m.commandBuilder.Build()
	debugLog("Generated command args: %v", args)
	
	// No dry-run mode - execute directly
	
	// Change to executing screen
	m.screen = ScreenExecuting
	m.executeOutput = []string{"Starting backup..."}
	
	// Create streaming executor for progress tracking
	streamingExec, err := NewStreamingExecutor(func(line string) {
		// Parse progress lines
		if strings.HasPrefix(line, "[") {
			m.executeOutput = append(m.executeOutput, line)
			// Keep only last 20 lines
			if len(m.executeOutput) > 20 {
				m.executeOutput = m.executeOutput[1:]
			}
		}
	})
	
	if err != nil {
		m.err = fmt.Errorf("failed to create executor: %w", err)
		return m
	}
	
	// Execute using streaming executor
	var output strings.Builder
	err = streamingExec.Execute(args, &output)
	
	if err != nil {
		debugLog("Backup failed: %v", err)
		// Provide more user-friendly error messages
		if strings.Contains(err.Error(), "executable file not found") {
			m.err = fmt.Errorf("internal error: cannot execute self - please report this issue")
		} else {
			m.err = fmt.Errorf("backup failed: %w", err)
		}
		m.screen = ScreenPathInput // Go back on error
	} else {
		debugLog("Backup completed successfully")
		// Parse output for completion message
		outputStr := output.String()
		if doneIdx := strings.Index(outputStr, "[DONE]"); doneIdx >= 0 {
			doneEnd := strings.Index(outputStr[doneIdx:], "\n")
			if doneEnd >= 0 {
				m.executeOutput = append(m.executeOutput, outputStr[doneIdx:doneIdx+doneEnd])
			}
		}
		m.executeOutput = append(m.executeOutput, "", "Press any key to continue...")
	}
	
	return m
}

// startTextInput initializes text input mode for a specific field
func startTextInput(m Model, field string, currentValue string) Model {
	m.editMode = true
	m.editField = field
	m.originalValue = currentValue
	
	// Setup text input
	m.textInput.SetValue(currentValue)
	m.textInput.Focus()
	
	// Reset echo mode (for password fields)
	m.textInput.EchoMode = textinput.EchoNormal
	
	// Set appropriate placeholder and prompt based on field
	switch field {
	case "container":
		m.textInput.Placeholder = "Container name (optional)"
		m.textInput.Prompt = "Container: "
	case "output":
		m.textInput.Placeholder = "Output filename (optional)"
		m.textInput.Prompt = "Output: "
	}
	
	return m
}

// saveTextInput saves the current text input value to the appropriate field
func saveTextInput(m Model) Model {
	value := m.textInput.Value()
	
	switch m.editField {
	case "output":
		m.backupOptions.OutputFile = value
		m.commandBuilder.SetOptions(m.backupOptions)
	}
	
	// Exit edit mode
	m.editMode = false
	m.editField = ""
	m.textInput.Blur()
	
	return m
}

// cancelTextInput cancels text input and restores original value
func cancelTextInput(m Model) Model {
	// Restore original value if needed
	switch m.editField {
	case "output":
		m.backupOptions.OutputFile = m.originalValue
	}
	
	// Exit edit mode
	m.editMode = false
	m.editField = ""
	m.textInput.Blur()
	
	return m
}
// setDefaultOutputFilename sets a default output filename based on the selected path and backup type
func setDefaultOutputFilename(m Model) Model {
	if m.selectedPath == "" {
		return m
	}
	
	timestamp := time.Now().Format("20060102_150405")
	baseName := filepath.Base(m.selectedPath)
	if baseName == "/" || baseName == "." {
		baseName = "backup"
	}
	
	// Clean the basename to remove invalid characters
	baseName = strings.ReplaceAll(baseName, " ", "_")
	baseName = strings.ReplaceAll(baseName, ":", "_")
	
	var extension string
	
	switch m.selectedBackupType {
	case "filesystem":
		switch m.backupOptions.CompressionType {
		case "gzip":
			extension = ".tar.gz"
		case "bzip2":
			extension = ".tar.bz2"
		case "xz":
			extension = ".tar.xz"
		default:
			extension = ".tar"
		}
		
		defaultName := fmt.Sprintf("%s_%s%s", baseName, timestamp, extension)
		if m.backupOptions.OutputFile == "" {
			m.backupOptions.OutputFile = defaultName
		}
	}
	
	return m
}

package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)

// HandleKey processes keyboard input
func HandleKey(m Model, key string) Model {
	switch key {
	case "q", "ctrl+c":
		return handleQuit(m)
	case "j", "down":
		return handleDownNavigation(m)
	case "k", "up":
		return handleUpNavigation(m)
	case "enter":
		return handleEnter(m)
	case " ", "space":
		return handleSpace(m)
	case "tab":
		return handleTab(m)
	case "b", "esc":
		return handleBack(m)
	}
	
	return m
}

func handleQuit(m Model) Model {
	if m.screen == ScreenMain && m.selected == 2 {
		m.quit = true
		return m
	}
	if m.screen != ScreenMain {
		m.quit = true
		return m
	}
	m.selected = 2
	return m
}

func handleDownNavigation(m Model) Model {
	if m.screen == ScreenBackupOptions {
		m.optionSelected++
		maxItems := getMaxItems(m)
		if m.optionSelected >= maxItems {
			m.optionSelected = maxItems - 1
		}
	} else {
		m.selected++
		maxItems := getMaxItems(m)
		if m.selected >= maxItems {
			m.selected = maxItems - 1
		}
	}
	return m
}

func handleUpNavigation(m Model) Model {
	if m.screen == ScreenBackupOptions {
		m.optionSelected--
		if m.optionSelected < 0 {
			m.optionSelected = 0
		}
	} else {
		m.selected--
		if m.selected < 0 {
			m.selected = 0
		}
	}
	return m
}

func handleSpace(m Model) Model {
	if m.screen == ScreenDirectoryBrowser {
		m.selectedPath = m.currentPath
		
		// Update command builder with selected path
		m.commandBuilder.SetPath(m.selectedPath)
		
		m.screen = ScreenBackupOptions
		m.optionCategory = 0
		m.optionSelected = 0
		m.selected = 0
		return m
	}
	if m.screen == ScreenBackupOptions {
		return handleOptionToggle(m)
	}
	return m
}

func handleTab(m Model) Model {
	if m.screen == ScreenBackupOptions {
		// Different number of tabs based on backup type
		maxTabs := 3 // default for filesystem
		switch m.selectedBackupType {
		case "minio":
			maxTabs = 2 // Connection, Backup Settings
		case "mongodb":
			maxTabs = 3 // Connection, Auth, Backup Settings
		}
		
		m.optionCategory = (m.optionCategory + 1) % maxTabs
		m.optionSelected = 0
	}
	return m
}

// handleOptionToggle toggles backup options
func handleOptionToggle(m Model) Model {
	switch m.selectedBackupType {
	case "minio":
		return handleMinioOptionToggle(m)
	case "mongodb":
		return handleMongoOptionToggle(m)
	default: // filesystem
		return handleFilesystemOptionToggle(m)
	}
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
	}
	
	// Update command builder with new options
	m.commandBuilder.SetOptions(m.backupOptions)
	
	return m
}

// handleMinioOptionToggle handles MinIO backup option toggles
func handleMinioOptionToggle(m Model) Model {
	switch m.optionCategory {
	case 0: // Connection
		// Connection settings are handled differently (text input)
		// For now, just return as-is
		
	case 1: // Backup Settings
		formats := []string{"tar", "zip"}
		if m.optionSelected < len(formats) {
			m.minioOptions.Format = formats[m.optionSelected]
		} else if m.optionSelected == len(formats) {
			// Toggle recursive
			m.minioOptions.Recursive = !m.minioOptions.Recursive
		}
	}
	
	// Update command builder with MinIO options
	m.commandBuilder.SetMinioOptions(m.minioOptions)
	
	return m
}

// handleMongoOptionToggle handles MongoDB backup option toggles
func handleMongoOptionToggle(m Model) Model {
	switch m.optionCategory {
	case 0: // Connection
		// Connection settings are handled differently (text input)
		
	case 1: // Auth
		// Auth settings are handled differently (text input)
		
	case 2: // Backup Settings
		switch m.optionSelected {
		case 0:
			m.mongoOptions.Gzip = !m.mongoOptions.Gzip
		case 1:
			m.mongoOptions.Oplog = !m.mongoOptions.Oplog
		case 2:
			// Collections are handled differently (text input)
		}
	}
	
	// Update command builder with MongoDB options
	m.commandBuilder.SetMongoOptions(m.mongoOptions)
	
	return m
}

// getMaxItems returns the number of items in current screen
func getMaxItems(m Model) int {
	switch m.screen {
	case ScreenMain:
		return 3 // Backup, Restore, Exit
	case ScreenBackupType:
		return 3 // filesystem, minio, mongodb
	case ScreenNamespaceList:
		return len(m.namespaces)
	case ScreenPodList:
		return len(m.pods)
	case ScreenDirectoryBrowser:
		return len(m.directories)
	case ScreenBackupOptions:
		switch m.selectedBackupType {
		case "minio":
			switch m.optionCategory {
			case 0: // Connection
				return 3 // endpoint, access key, secret key
			case 1: // Backup Settings
				return 3 // tar, zip, recursive
			}
		case "mongodb":
			switch m.optionCategory {
			case 0: // Connection
				return 1 // host
			case 1: // Auth
				return 3 // username, password, auth db
			case 2: // Backup Settings
				return 3 // gzip, oplog, collections
			}
		default: // filesystem
			switch m.optionCategory {
			case 0: // Compression
				return 4 // gzip, bzip2, xz, none
			case 1: // Excludes
				return 6 // 5 patterns + VCS toggle
			case 2: // Advanced
				return 3 // verbose, totals, preserve
			}
		}
		return 1
	default:
		return 1
	}
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
	m.selectedPod = m.pods[m.selected].Name
	
	// Update command builder
	m.commandBuilder.SetPod(m.selectedPod)
	
	// For MinIO and MongoDB, skip directory browsing
	switch m.selectedBackupType {
	case "minio":
		// For MinIO, path will be bucket/path specified in options
		m.selectedPath = "" // Will be set via options
		m.screen = ScreenBackupOptions
		m.optionCategory = 0
		m.optionSelected = 0
		m.selected = 0
		return m
		
	case "mongodb":
		// For MongoDB, path will be database name specified in options
		m.selectedPath = "" // Will be set via options
		m.screen = ScreenBackupOptions
		m.optionCategory = 0
		m.optionSelected = 0
		m.selected = 0
		return m
		
	default: // filesystem
		m.currentPath = "/"
		directories, err := kubernetes.GetDirectoryContents(m.runner, m.selectedPod, m.selectedNamespace, "/")
		if err != nil {
			m.err = err
			return m
		}
		m.directories = directories
		m.screen = ScreenDirectoryBrowser
		m.selected = 0
		return m
	}
}

func handleDirectoryEnter(m Model) Model {
	if m.selected < len(m.directories) {
		selectedEntry := m.directories[m.selected]
		if selectedEntry.Type == "dir" {
			newPath := filepath.Join(m.currentPath, selectedEntry.Name)
			directories, err := kubernetes.GetDirectoryContents(m.runner, m.selectedPod, m.selectedNamespace, newPath)
			if err != nil {
				m.err = err
				return m
			}
			m.currentPath = newPath
			m.directories = directories
			m.selected = 0
		}
	}
	return m
}

func handleBackupTypeEnter(m Model) Model {
	debugLog("handleBackupTypeEnter: selected=%d", m.selected)
	
	backupTypes := []string{"filesystem", "minio", "mongodb"}
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
	// For MinIO and MongoDB, set default paths if not already set
	switch m.selectedBackupType {
	case "minio":
		if m.selectedPath == "" {
			m.selectedPath = "mybucket" // Default bucket name
		}
		m.commandBuilder.SetPath(m.selectedPath)
		
	case "mongodb":
		if m.selectedPath == "" {
			m.selectedPath = "mydb" // Default database name
		}
		m.commandBuilder.SetPath(m.selectedPath)
	}
	
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
	
	// Execute using our own CLI
	var output strings.Builder
	err := m.executor.Execute(args, &output)
	
	if err != nil {
		debugLog("Backup failed: %v", err)
		// Provide more user-friendly error messages
		if strings.Contains(err.Error(), "executable file not found") {
			m.err = fmt.Errorf("internal error: cannot execute self - please report this issue")
		} else {
			m.err = fmt.Errorf("backup failed: %w", err)
		}
	} else {
		debugLog("Backup completed successfully")
		// Try to extract output filename from the output
		outputLines := strings.Split(output.String(), "\n")
		lastLine := ""
		for i := len(outputLines) - 1; i >= 0; i-- {
			if strings.TrimSpace(outputLines[i]) != "" {
				lastLine = outputLines[i]
				break
			}
		}
		m.err = fmt.Errorf("backup completed successfully: %s", lastLine)
	}
	
	return m
}

// handleBack processes back navigation
func handleBack(m Model) Model {
	switch m.screen {
	case ScreenBackupType:
		m.screen = ScreenMain
		m.selected = 0
		
	case ScreenNamespaceList:
		m.screen = ScreenBackupType
		m.selected = 0
		
	case ScreenPodList:
		m.screen = ScreenNamespaceList
		m.selected = 0
		
	case ScreenDirectoryBrowser:
		if m.currentPath == "/" {
			// Go back to pod list
			m.screen = ScreenPodList
			m.selected = 0
		} else {
			// Go to parent directory
			parentPath := filepath.Dir(m.currentPath)
			directories, err := kubernetes.GetDirectoryContents(m.runner, m.selectedPod, m.selectedNamespace, parentPath)
			if err != nil {
				m.err = err
				return m
			}
			m.currentPath = parentPath
			m.directories = directories
			m.selected = 0
		}
		
	case ScreenBackupOptions:
		m.screen = ScreenDirectoryBrowser
		m.selected = 0
		
	case ScreenPathInput:
		m.screen = ScreenBackupOptions
		m.selected = 0
	}
	
	return m
}

// Note: kubectl-based backup functions have been removed.
// The TUI now uses the CommandBuilder and Executor pattern
// to call cli-recover directly, ensuring consistency between
// TUI and CLI modes.
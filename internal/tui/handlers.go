package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cagojeiger/cli-restore/internal/kubernetes"
	"github.com/cagojeiger/cli-restore/internal/runner"
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
		m.optionCategory = (m.optionCategory + 1) % 3 // 0: compression, 1: excludes, 2: advanced
		m.optionSelected = 0
	}
	return m
}

// handleOptionToggle toggles backup options
func handleOptionToggle(m Model) Model {
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
	
	return m
}

// getMaxItems returns the number of items in current screen
func getMaxItems(m Model) int {
	switch m.screen {
	case ScreenMain:
		return 3 // Backup, Restore, Exit
	case ScreenNamespaceList:
		return len(m.namespaces)
	case ScreenPodList:
		return len(m.pods)
	case ScreenDirectoryBrowser:
		return len(m.directories)
	case ScreenBackupOptions:
		switch m.optionCategory {
		case 0: // Compression
			return 4 // gzip, bzip2, xz, none
		case 1: // Excludes
			return 6 // 5 patterns + VCS toggle
		case 2: // Advanced
			return 3 // verbose, totals, preserve
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
		debugLog("Starting backup flow - loading namespaces")
		namespaces, err := kubernetes.GetNamespaces(m.runner)
		if err != nil {
			debugLog("Error loading namespaces: %v", err)
			m.err = err
			return m
		}
		debugLog("Loaded %d namespaces", len(namespaces))
		m.namespaces = namespaces
		m.screen = ScreenNamespaceList
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

func handleBackupOptionsEnter(m Model) Model {
	m.screen = ScreenPathInput
	m.selected = 0
	return m
}

func handlePathInputEnter(m Model) Model {
	debugLog("handlePathInputEnter: executing backup")
	debugLog("  namespace: %s, pod: %s, path: %s", m.selectedNamespace, m.selectedPod, m.selectedPath)
	debugLog("  options: %+v", m.backupOptions)
	
	// Generate backup command
	command := kubernetes.GenerateBackupCommand(m.selectedPod, m.selectedNamespace, m.selectedPath, m.backupOptions)
	debugLog("Generated command: %s", command)
	
	// Generate output filename
	pathSuffix := generatePathSuffix(m.selectedPath)
	extension := getFileExtension(m.backupOptions.CompressionType)
	outputFile := fmt.Sprintf("backup-%s-%s-%s%s", m.selectedNamespace, m.selectedPod, pathSuffix, extension)
	
	debugLog("Output file: %s", outputFile)
	
	// Execute backup (simplified for TUI)
	err := executeBackupTUI(m.runner, command, outputFile)
	if err != nil {
		debugLog("Backup failed: %v", err)
		m.err = fmt.Errorf("backup failed: %w", err)
	} else {
		debugLog("Backup completed successfully")
		m.err = fmt.Errorf("backup completed successfully: %s", outputFile)
	}
	
	return m
}

// handleBack processes back navigation
func handleBack(m Model) Model {
	switch m.screen {
	case ScreenNamespaceList:
		m.screen = ScreenMain
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

// generatePathSuffix creates a safe filename suffix from a path
func generatePathSuffix(path string) string {
	if path == "/" {
		return "root"
	}
	// Remove leading slash and replace slashes with dashes
	suffix := strings.TrimPrefix(path, "/")
	suffix = strings.ReplaceAll(suffix, "/", "-")
	suffix = strings.ReplaceAll(suffix, " ", "-")
	suffix = strings.ReplaceAll(suffix, ".", "-")
	return suffix
}

// getFileExtension returns file extension based on compression type
func getFileExtension(compression string) string {
	switch compression {
	case "gzip":
		return ".tar.gz"
	case "bzip2":
		return ".tar.bz2"
	case "xz":
		return ".tar.xz"
	case "none":
		return ".tar"
	default:
		return ".tar.gz"
	}
}

// executeBackupTUI performs backup for TUI mode
func executeBackupTUI(runner runner.Runner, command, outputFile string) error {
	debugLog("executeBackupTUI: creating output file %s", outputFile)
	
	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputFile, err)
	}
	defer outFile.Close()
	
	// Execute kubectl command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	
	debugLog("executeBackupTUI: executing command with %d parts", len(parts))
	
	// Execute the command and get output
	output, err := runner.Run(parts[0], parts[1:]...)
	if err != nil {
		return fmt.Errorf("backup command failed: %w", err)
	}
	
	// Write output to file
	_, err = outFile.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}
	
	debugLog("executeBackupTUI: backup completed, %d bytes written", len(output))
	return nil
}
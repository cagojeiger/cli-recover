package tui

import (
	"path/filepath"

	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)

// handleBack processes back navigation
func handleBack(m Model) Model {
	// Special handling for directory browser - navigate within directories
	if m.screen == ScreenDirectoryBrowser && m.currentPath != "/" {
		// Go to parent directory
		parentPath := filepath.Dir(m.currentPath)
		directories, err := kubernetes.GetDirectoryContents(m.runner, m.selectedPod, m.selectedNamespace, parentPath, m.selectedContainer)
		if err != nil {
			m.err = err
			return m
		}
		m.currentPath = parentPath
		m.directories = directories
		m.selected = 0
		return m
	}
	
	// Use screen history for navigation
	return m.popScreen()
}

// getMaxItems returns the number of items in current screen
func getMaxItems(m Model) int {
	switch m.screen {
	case ScreenMain:
		return 3 // Backup, Restore, Exit
	case ScreenBackupType:
		return 1 // filesystem only
	case ScreenNamespaceList:
		return len(m.namespaces)
	case ScreenPodList:
		return len(m.pods)
	case ScreenContainerList:
		// Find selected pod and return container count
		for _, pod := range m.pods {
			if pod.Name == m.selectedPod {
				return len(pod.Containers)
			}
		}
		return 0
	case ScreenDirectoryBrowser:
		return len(m.directories)
	case ScreenBackupOptions:
		// Only filesystem backup is supported now
		switch m.optionCategory {
		case 0: // Compression
			return 4 // gzip, bzip2, xz, none
		case 1: // Excludes
			return 6 // 5 patterns + VCS toggle
		case 2: // Advanced
			return 3 // verbose, totals, preserve
		case 3: // Output
			return 1 // output file only
		}
		return 1
	default:
		return 1
	}
}

// getNumCategories returns the number of option categories for the current backup type
func getNumCategories(m Model) int {
	// Only filesystem backup is supported now
	return 4 // Compression, Excludes, Advanced, Output
}

// Note: kubectl-based backup functions have been removed.
// The TUI now uses the CommandBuilder and Executor pattern
// to call cli-recover directly, ensuring consistency between
// TUI and CLI modes.
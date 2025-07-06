package tui

import (
	"fmt"
)

// version will be set by ldflags during build (import from main)
var version = "dev"

// SetVersion sets the version for TUI display
func SetVersion(v string) {
	version = v
}

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit", m.err)
	}
	
	width, _ := getViewDimensions(m)
	
	var view string
	view += renderHeader(width)
	view += renderContent(m, width)
	view += renderCommand(m, width)
	view += renderFooter(m, width)
	
	return view
}

func getViewDimensions(m Model) (int, int) {
	width := m.width
	if width == 0 {
		width = 80
	}
	// Ensure minimum width for proper display
	if width < 60 {
		width = 60
	}
	height := m.height
	if height == 0 {
		height = 20
	}
	return width, height
}

func renderHeader(width int) string {
	header := fmt.Sprintf("CLI Recover v%s", version)
	return fmt.Sprintf("=== %s ===\n\n", header)
}

func renderContent(m Model, width int) string {
	switch m.screen {
	case ScreenMain:
		return viewMainMenu(m, width)
	case ScreenBackupType:
		return viewBackupType(m, width)
	case ScreenNamespaceList:
		return viewNamespaceList(m, width)
	case ScreenPodList:
		return viewPodList(m, width)
	case ScreenDirectoryBrowser:
		return viewDirectoryBrowser(m, width)
	case ScreenBackupOptions:
		return viewBackupOptions(m, width)
	case ScreenPathInput:
		return viewPathInput(m, width)
	}
	return ""
}

func renderCommand(m Model, width int) string {
	// Don't show command on main menu or error states
	if m.screen == ScreenMain {
		return ""
	}
	
	command := m.commandBuilder.Preview()
	if command == "cli-recover" {
		return ""
	}
	
	return fmt.Sprintf("\n---\nCommand: %s\n", command)
}

func renderFooter(m Model, width int) string {
	var instructions string
	
	switch m.screen {
	case ScreenDirectoryBrowser:
		instructions = "[↑/↓] Navigate  [Enter] Open  [Space] Select  [b] Back  [q] Quit"
	case ScreenBackupOptions:
		instructions = "[↑/↓] Navigate  [Space] Toggle  [Tab] Category  [Enter] OK  [b] Back"
	case ScreenPathInput:
		instructions = "[Enter] Execute  [b] Back  [q] Quit"
	default:
		instructions = "[↑/↓] Navigate  [Enter] Select  [b] Back  [q] Quit"
	}
	
	return fmt.Sprintf("\n%s\n", instructions)
}
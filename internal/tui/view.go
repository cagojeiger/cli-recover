package tui

import (
	"fmt"
	"strings"
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
	
	// Add text input overlay if in edit mode
	if m.editMode {
		view += renderTextInputOverlay(m, width)
	}
	
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
	case ScreenContainerList:
		return viewContainerList(m, width)
	case ScreenDirectoryBrowser:
		return viewDirectoryBrowser(m, width)
	case ScreenBackupOptions:
		return viewBackupOptions(m, width)
	case ScreenPathInput:
		return viewPathInput(m, width)
	case ScreenExecuting:
		return viewExecuting(m, width)
	case ScreenJobManager:
		return viewJobManager(m, width)
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
	case ScreenExecuting:
		if len(m.executeOutput) > 0 && strings.Contains(m.executeOutput[len(m.executeOutput)-1], "Press any key") {
			instructions = "[Any key] Continue"
		} else {
			instructions = "Executing backup..."
		}
	case ScreenJobManager:
		instructions = "[↑/↓] Navigate  [c] Cancel Job  [K] Cancel All  [b] Back"
	default:
		instructions = "[↑/↓] Navigate  [Enter] Select  [b] Back  [q] Quit"
	}
	
	// Add job status if any jobs are running
	if m.jobManager != nil && m.hasActiveJobs() {
		jobStatus := fmt.Sprintf(" | Jobs: %s", m.getJobSummary())
		instructions += jobStatus
	}
	
	// Add shortcut hint
	if m.screen != ScreenJobManager && m.jobManager != nil {
		instructions += " [J] Job Manager"
	}
	
	return fmt.Sprintf("\n%s\n", instructions)
}

// renderTextInputOverlay renders the text input overlay
func renderTextInputOverlay(m Model, width int) string {
	// Create a simple overlay with the text input
	overlay := "\n\n"
	overlay += strings.Repeat("=", width/2) + "\n"
	overlay += m.textInput.View() + "\n"
	overlay += "\n[Enter] Save  [Esc] Cancel\n"
	overlay += strings.Repeat("=", width/2) + "\n"
	
	return overlay
}
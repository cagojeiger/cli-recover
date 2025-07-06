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
	view += renderFooterBorder(width)
	view += renderFooterInstructions(m)
	
	return view
}

func getViewDimensions(m Model) (int, int) {
	width := m.width
	if width == 0 {
		width = 50
	}
	height := m.height
	if height == 0 {
		height = 20
	}
	return width, height
}

func renderHeader(width int) string {
	header := fmt.Sprintf("CLI Recover v%s", version)
	headerPadding := width - len(header) - 4
	if headerPadding < 0 {
		headerPadding = 0
	}
	return "┌─ " + header + " " + strings.Repeat("─", headerPadding) + "┐\n"
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

func renderFooterBorder(width int) string {
	return "└" + strings.Repeat("─", width-2) + "┘\n"
}

func renderFooterInstructions(m Model) string {
	switch m.screen {
	case ScreenDirectoryBrowser:
		return "↑/↓ Navigate  Enter Open  Space Select  b Back  q Quit"
	case ScreenBackupOptions:
		return "↑/↓ Navigate  Space Toggle  Tab Category  Enter OK  b Back"
	default:
		return "↑/↓ Navigate  Enter Select  b Back  q Quit"
	}
}
package tui

import (
	"fmt"
	"strings"
	
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)

// viewMainMenu renders the main menu
func viewMainMenu(m Model, width int) string {
	items := []string{"Backup", "Restore", "Exit"}
	
	var view string
	contentWidth := width - 2 // Account for borders
	
	// Title
	title := "Main Menu"
	titlePadding := contentWidth - len(title)
	view += "│ " + title + strings.Repeat(" ", titlePadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// Menu items
	for i, item := range items {
		var line string
		if i == m.selected {
			line = fmt.Sprintf(" > %s", item)
		} else {
			line = fmt.Sprintf("   %s", item)
		}
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}

// viewNamespaceList renders namespace selection
func viewNamespaceList(m Model, width int) string {
	var view string
	contentWidth := width - 2
	
	// Title
	title := "Select Namespace"
	titlePadding := contentWidth - len(title)
	view += "│ " + title + strings.Repeat(" ", titlePadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// Command preview
	preview := m.commandBuilder.Preview()
	if len(preview) > contentWidth {
		preview = preview[:contentWidth-3] + "..."
	}
	previewPadding := contentWidth - len(preview)
	if previewPadding < 0 {
		previewPadding = 0
	}
	view += "│" + preview + strings.Repeat(" ", previewPadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// Namespace list
	for i, ns := range m.namespaces {
		var line string
		if i == m.selected {
			line = fmt.Sprintf(" > %s", ns)
		} else {
			line = fmt.Sprintf("   %s", ns)
		}
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}

// viewPodList renders pod selection
func viewPodList(m Model, width int) string {
	var view string
	contentWidth := width - 2
	
	// Title
	title := fmt.Sprintf("Pods in %s", m.selectedNamespace)
	titlePadding := contentWidth - len(title)
	if titlePadding < 0 {
		titlePadding = 0
	}
	view += "│ " + title + strings.Repeat(" ", titlePadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// Command preview
	preview := m.commandBuilder.Preview()
	if len(preview) > contentWidth {
		preview = preview[:contentWidth-3] + "..."
	}
	previewPadding := contentWidth - len(preview)
	if previewPadding < 0 {
		previewPadding = 0
	}
	view += "│" + preview + strings.Repeat(" ", previewPadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// Pod list
	for i, pod := range m.pods {
		// Format pod info to fit width
		nameLen := contentWidth / 3
		if nameLen > len(pod.Name) {
			nameLen = len(pod.Name)
		}
		name := pod.Name
		if len(name) > nameLen {
			name = name[:nameLen-3] + "..."
		}
		
		display := fmt.Sprintf("%-*s %s %s", nameLen, name, pod.Status, pod.Ready)
		
		var line string
		if i == m.selected {
			line = fmt.Sprintf(" > %s", display)
		} else {
			line = fmt.Sprintf("   %s", display)
		}
		
		if len(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}
		
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}

// viewDirectoryBrowser renders directory browsing screen
func viewDirectoryBrowser(m Model, width int) string {
	contentWidth := width - 2
	var view string
	
	view += renderDirectoryTitle(m.currentPath, contentWidth)
	view += renderCommandPreview(m, contentWidth)
	view += renderDirectoryEntries(m, contentWidth)
	view += renderDirectoryInstructions(contentWidth)
	
	return view
}

func renderDirectoryTitle(currentPath string, contentWidth int) string {
	title := fmt.Sprintf("Browse: %s", currentPath)
	titlePadding := contentWidth - len(title)
	if titlePadding < 0 {
		titlePadding = 0
		title = title[:contentWidth-3] + "..."
	}
	view := "│ " + title + strings.Repeat(" ", titlePadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	return view
}

func renderCommandPreview(m Model, contentWidth int) string {
	preview := m.commandBuilder.Preview()
	if len(preview) > contentWidth {
		preview = preview[:contentWidth-3] + "..."
	}
	previewPadding := contentWidth - len(preview)
	if previewPadding < 0 {
		previewPadding = 0
	}
	view := "│" + preview + strings.Repeat(" ", previewPadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	return view
}

func renderDirectoryEntries(m Model, contentWidth int) string {
	var view string
	for i, entry := range m.directories {
		view += renderDirectoryEntry(entry, i, m.selected, contentWidth)
	}
	return view
}

func renderDirectoryEntry(entry kubernetes.DirectoryEntry, index, selected, contentWidth int) string {
	prefix := "   "
	if index == selected {
		prefix = " > "
	}
	
	icon := getEntryIcon(entry.Type)
	entryName := formatEntryName(prefix, icon, entry.Name, contentWidth)
	typeInfo := getEntryTypeInfo(entry)
	
	line := fmt.Sprintf("%-*s %s", contentWidth-12, entryName, typeInfo)
	if len(line) > contentWidth {
		line = line[:contentWidth]
	}
	
	padding := contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}
	return "│" + line + strings.Repeat(" ", padding) + "│\n"
}

func getEntryIcon(entryType string) string {
	if entryType == "dir" {
		return "📁"
	}
	return "📄"
}

func formatEntryName(prefix, icon, name string, contentWidth int) string {
	entryName := fmt.Sprintf("%s%s %s", prefix, icon, name)
	if len(entryName) > contentWidth-15 {
		entryName = entryName[:contentWidth-18] + "..."
	}
	return entryName
}

func getEntryTypeInfo(entry kubernetes.DirectoryEntry) string {
	if entry.Type == "file" {
		return entry.Size
	}
	return entry.Type
}

func renderDirectoryInstructions(contentWidth int) string {
	view := "├" + strings.Repeat("─", contentWidth) + "┤\n"
	instructions := " Enter: Open dir  Space: Select current path  b: Back"
	if len(instructions) > contentWidth {
		instructions = instructions[:contentWidth-3] + "..."
	}
	instrPadding := contentWidth - len(instructions)
	view += "│" + instructions + strings.Repeat(" ", instrPadding) + "│\n"
	return view
}

// viewBackupType renders the backup type selection screen
func viewBackupType(m Model, width int) string {
	backupTypes := []struct {
		name        string
		description string
	}{
		{"filesystem", "Backup files and directories from pod filesystem"},
		{"minio", "Backup objects from MinIO object storage"},
		{"mongodb", "Backup collections from MongoDB database"},
	}
	
	var view string
	contentWidth := width - 2
	
	// Title
	title := "Select Backup Type"
	titlePadding := contentWidth - len(title)
	view += "│ " + title + strings.Repeat(" ", titlePadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// Backup type options
	for i, bt := range backupTypes {
		var line string
		if i == m.selected {
			line = fmt.Sprintf(" > %-12s - %s", bt.name, bt.description)
		} else {
			line = fmt.Sprintf("   %-12s - %s", bt.name, bt.description)
		}
		
		// Truncate if too long
		if len(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}
		
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}
package tui

import (
	"fmt"
	"strings"
)

// viewBackupOptions renders backup options configuration screen
func viewBackupOptions(m Model, width int) string {
	var view string
	contentWidth := width - 2
	
	// Title
	title := "Backup Options"
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
	
	// Category tabs
	categories := []string{"Compression", "Excludes", "Advanced"}
	var tabLine string
	for i, cat := range categories {
		if i == m.optionCategory {
			tabLine += fmt.Sprintf(" [%s] ", cat)
		} else {
			tabLine += fmt.Sprintf("  %s  ", cat)
		}
	}
	tabPadding := contentWidth - len(tabLine)
	if tabPadding < 0 {
		tabPadding = 0
	}
	view += "│" + tabLine + strings.Repeat(" ", tabPadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// Category content
	switch m.optionCategory {
	case 0: // Compression
		view += viewCompressionOptions(m, contentWidth)
	case 1: // Excludes
		view += viewExcludeOptions(m, contentWidth)
	case 2: // Advanced
		view += viewAdvancedOptions(m, contentWidth)
	}
	
	return view
}

// viewCompressionOptions renders compression options
func viewCompressionOptions(m Model, contentWidth int) string {
	var view string
	compressionTypes := []string{"gzip", "bzip2", "xz", "none"}
	
	for i, compType := range compressionTypes {
		var line string
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		checkbox := "[ ]"
		if m.backupOptions.CompressionType == compType {
			checkbox = "[x]"
		}
		
		line = fmt.Sprintf("%s%s %s", prefix, checkbox, compType)
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}

// viewExcludeOptions renders exclude pattern options
func viewExcludeOptions(m Model, contentWidth int) string {
	var view string
	excludeOptions := []string{"*.log", "tmp/*", ".git", "node_modules/*", "*.tmp"}
	
	for i, pattern := range excludeOptions {
		var line string
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		checkbox := "[ ]"
		for _, existing := range m.backupOptions.ExcludePatterns {
			if existing == pattern {
				checkbox = "[x]"
				break
			}
		}
		
		line = fmt.Sprintf("%s%s %s", prefix, checkbox, pattern)
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	// VCS exclusion
	var line string
	prefix := "   "
	if len(excludeOptions) == m.optionSelected {
		prefix = " > "
	}
	
	checkbox := "[ ]"
	if m.backupOptions.ExcludeVCS {
		checkbox = "[x]"
	}
	
	line = fmt.Sprintf("%s%s Exclude VCS (.git, .svn)", prefix, checkbox)
	padding := contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}
	view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	
	return view
}

// viewAdvancedOptions renders advanced options
func viewAdvancedOptions(m Model, contentWidth int) string {
	var view string
	options := []struct {
		name  string
		value bool
	}{
		{"Verbose output", m.backupOptions.Verbose},
		{"Show totals", m.backupOptions.ShowTotals},
		{"Preserve permissions", m.backupOptions.PreservePerms},
	}
	
	for i, option := range options {
		var line string
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		checkbox := "[ ]"
		if option.value {
			checkbox = "[x]"
		}
		
		line = fmt.Sprintf("%s%s %s", prefix, checkbox, option.name)
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}

// viewPathInput renders path input screen
func viewPathInput(m Model, width int) string {
	contentWidth := width - 2
	var view string
	
	view += renderConfigHeader(contentWidth)
	view += renderConfigDetails(m, contentWidth)
	view += renderCommandComparison(m, contentWidth)
	view += renderInstructions(contentWidth)
	
	return view
}

func renderConfigHeader(contentWidth int) string {
	title := "Backup Configuration"
	titlePadding := contentWidth - len(title)
	view := "│ " + title + strings.Repeat(" ", titlePadding) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	return view
}

func renderConfigDetails(m Model, contentWidth int) string {
	var view string
	configs := []struct {
		label string
		value string
	}{
		{"Namespace", m.selectedNamespace},
		{"Pod", m.selectedPod},
		{"Path", m.selectedPath},
	}
	
	for _, config := range configs {
		line := fmt.Sprintf(" %s: %s", config.label, config.value)
		if len(line) > contentWidth {
			line = line[:contentWidth-3] + "..."
		}
		padding := contentWidth - len(line)
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	return view
}

func renderCommandComparison(m Model, contentWidth int) string {
	var view string
	
	// Command header
	cmdTitle := " Command to execute:"
	view += "│" + cmdTitle + strings.Repeat(" ", contentWidth-len(cmdTitle)) + "│\n"
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	
	// cli-restore command section
	view += renderCliRestoreCommand(m, contentWidth)
	
	return view
}

// renderKubectlCommand removed - we now use cli-restore directly

func renderCliRestoreCommand(m Model, contentWidth int) string {
	var view string
	
	// Use the actual command from CommandBuilder
	cliCmd := m.commandBuilder.Preview()
	cliLine := " $ " + cliCmd
	
	view += wrapCommand(cliLine, contentWidth)
	return view
}

func wrapCommand(command string, contentWidth int) string {
	var view string
	cmdLine := command
	
	for len(cmdLine) > 0 {
		lineLen := contentWidth
		if len(cmdLine) < lineLen {
			lineLen = len(cmdLine)
		}
		
		line := cmdLine[:lineLen]
		cmdLine = cmdLine[lineLen:]
		
		padding := contentWidth - len(line)
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
		
		if len(cmdLine) > 0 {
			cmdLine = "   " + cmdLine
		}
	}
	
	return view
}

func renderInstructions(contentWidth int) string {
	var view string
	view += "├" + strings.Repeat("─", contentWidth) + "┤\n"
	instructions := " Press Enter to execute, b to go back"
	instrPadding := contentWidth - len(instructions)
	view += "│" + instructions + strings.Repeat(" ", instrPadding) + "│\n"
	return view
}
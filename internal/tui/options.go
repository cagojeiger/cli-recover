package tui

import (
	"fmt"
	"strings"
)

// viewBackupOptions renders backup options configuration screen
func viewBackupOptions(m Model, width int) string {
	// Route to appropriate options view based on backup type
	switch m.selectedBackupType {
	case "minio":
		return viewMinioOptions(m, width)
	case "mongodb":
		return viewMongoOptions(m, width)
	default: // filesystem
		return viewFilesystemOptions(m, width)
	}
}

// viewFilesystemOptions renders filesystem backup options
func viewFilesystemOptions(m Model, width int) string {
	var view string
	contentWidth := width - 2
	
	// Title
	title := "Filesystem Backup Options"
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

// viewMinioOptions renders MinIO backup options
func viewMinioOptions(m Model, width int) string {
	var view string
	contentWidth := width - 2
	
	// Title
	title := "MinIO Backup Options"
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
	
	// Category tabs for MinIO
	categories := []string{"Connection", "Backup Settings"}
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
	case 0: // Connection
		view += viewMinioConnectionOptions(m, contentWidth)
	case 1: // Backup Settings
		view += viewMinioBackupSettings(m, contentWidth)
	}
	
	return view
}

// viewMinioConnectionOptions renders MinIO connection options
func viewMinioConnectionOptions(m Model, contentWidth int) string {
	var view string
	options := []struct {
		name  string
		value string
	}{
		{"Endpoint", m.minioOptions.Endpoint},
		{"Access Key", m.minioOptions.AccessKey},
		{"Secret Key", m.minioOptions.SecretKey},
	}
	
	for i, option := range options {
		var line string
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		value := option.value
		if value == "" {
			value = "(not set)"
		}
		if option.name == "Secret Key" && value != "(not set)" {
			value = "********"
		}
		
		line = fmt.Sprintf("%s%s: %s", prefix, option.name, value)
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}

// viewMinioBackupSettings renders MinIO backup settings
func viewMinioBackupSettings(m Model, contentWidth int) string {
	var view string
	
	// Format selection
	formats := []string{"tar", "zip"}
	for i, format := range formats {
		var line string
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		checkbox := "[ ]"
		if m.minioOptions.Format == format {
			checkbox = "[x]"
		}
		
		line = fmt.Sprintf("%s%s Format: %s", prefix, checkbox, format)
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	// Recursive option
	var line string
	prefix := "   "
	if m.optionSelected == len(formats) {
		prefix = " > "
	}
	
	checkbox := "[ ]"
	if m.minioOptions.Recursive {
		checkbox = "[x]"
	}
	
	line = fmt.Sprintf("%s%s Recursive backup", prefix, checkbox)
	padding := contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}
	view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	
	return view
}

// viewMongoOptions renders MongoDB backup options
func viewMongoOptions(m Model, width int) string {
	var view string
	contentWidth := width - 2
	
	// Title
	title := "MongoDB Backup Options"
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
	
	// Category tabs for MongoDB
	categories := []string{"Connection", "Auth", "Backup Settings"}
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
	case 0: // Connection
		view += viewMongoConnectionOptions(m, contentWidth)
	case 1: // Auth
		view += viewMongoAuthOptions(m, contentWidth)
	case 2: // Backup Settings
		view += viewMongoBackupSettings(m, contentWidth)
	}
	
	return view
}

// viewMongoConnectionOptions renders MongoDB connection options
func viewMongoConnectionOptions(m Model, contentWidth int) string {
	var view string
	
	var line string
	prefix := "   "
	if m.optionSelected == 0 {
		prefix = " > "
	}
	
	line = fmt.Sprintf("%sHost: %s", prefix, m.mongoOptions.Host)
	padding := contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}
	view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	
	return view
}

// viewMongoAuthOptions renders MongoDB auth options
func viewMongoAuthOptions(m Model, contentWidth int) string {
	var view string
	options := []struct {
		name  string
		value string
	}{
		{"Username", m.mongoOptions.Username},
		{"Password", m.mongoOptions.Password},
		{"Auth DB", m.mongoOptions.AuthDB},
	}
	
	for i, option := range options {
		var line string
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		value := option.value
		if value == "" {
			value = "(not set)"
		}
		if option.name == "Password" && value != "(not set)" {
			value = "********"
		}
		
		line = fmt.Sprintf("%s%s: %s", prefix, option.name, value)
		padding := contentWidth - len(line)
		if padding < 0 {
			padding = 0
		}
		view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	}
	
	return view
}

// viewMongoBackupSettings renders MongoDB backup settings
func viewMongoBackupSettings(m Model, contentWidth int) string {
	var view string
	
	// Gzip option
	var line string
	prefix := "   "
	if m.optionSelected == 0 {
		prefix = " > "
	}
	
	checkbox := "[ ]"
	if m.mongoOptions.Gzip {
		checkbox = "[x]"
	}
	
	line = fmt.Sprintf("%s%s Compress with gzip", prefix, checkbox)
	padding := contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}
	view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	
	// Oplog option
	prefix = "   "
	if m.optionSelected == 1 {
		prefix = " > "
	}
	
	checkbox = "[ ]"
	if m.mongoOptions.Oplog {
		checkbox = "[x]"
	}
	
	line = fmt.Sprintf("%s%s Include oplog", prefix, checkbox)
	padding = contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}
	view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	
	// Collections
	prefix = "   "
	if m.optionSelected == 2 {
		prefix = " > "
	}
	
	collections := "all"
	if len(m.mongoOptions.Collections) > 0 {
		collections = strings.Join(m.mongoOptions.Collections, ", ")
	}
	
	line = fmt.Sprintf("%sCollections: %s", prefix, collections)
	padding = contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}
	view += "│" + line + strings.Repeat(" ", padding) + "│\n"
	
	return view
}
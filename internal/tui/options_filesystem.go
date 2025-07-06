package tui

import (
	"fmt"
	"strings"
)

// viewFilesystemOptions renders filesystem backup options
func viewFilesystemOptions(m Model, width int) string {
	var view string
	
	// Title
	view += "Filesystem Backup Options:\n\n"
	
	// Category tabs
	categories := []string{"Compression", "Excludes", "Advanced", "Output"}
	var tabLine string
	for i, cat := range categories {
		if i == m.optionCategory {
			tabLine += fmt.Sprintf(" [%s] ", cat)
		} else {
			tabLine += fmt.Sprintf("  %s  ", cat)
		}
	}
	view += tabLine + "\n"
	view += strings.Repeat("-", len(tabLine)) + "\n\n"
	
	// Category content
	switch m.optionCategory {
	case 0: // Compression
		view += viewCompressionOptions(m)
	case 1: // Excludes
		view += viewExcludeOptions(m)
	case 2: // Advanced
		view += viewAdvancedOptions(m)
	case 3: // Output
		view += viewOutputOptions(m)
	}
	
	return view
}

// viewCompressionOptions renders compression options
func viewCompressionOptions(m Model) string {
	var view string
	compressionTypes := []string{"gzip", "bzip2", "xz", "none"}
	
	for i, compType := range compressionTypes {
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		checkbox := "[ ]"
		if m.backupOptions.CompressionType == compType {
			checkbox = "[x]"
		}
		
		view += fmt.Sprintf("%s%s %s\n", prefix, checkbox, compType)
	}
	
	return view
}

// viewExcludeOptions renders exclude pattern options
func viewExcludeOptions(m Model) string {
	var view string
	excludeOptions := []string{"*.log", "tmp/*", ".git", "node_modules/*", "*.tmp"}
	
	for i, pattern := range excludeOptions {
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
		
		view += fmt.Sprintf("%s%s %s\n", prefix, checkbox, pattern)
	}
	
	// VCS exclusion
	prefix := "   "
	if len(excludeOptions) == m.optionSelected {
		prefix = " > "
	}
	
	checkbox := "[ ]"
	if m.backupOptions.ExcludeVCS {
		checkbox = "[x]"
	}
	
	view += fmt.Sprintf("%s%s Exclude VCS (.git, .svn)\n", prefix, checkbox)
	
	return view
}

// viewAdvancedOptions renders advanced options
func viewAdvancedOptions(m Model) string {
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
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		checkbox := "[ ]"
		if option.value {
			checkbox = "[x]"
		}
		
		view += fmt.Sprintf("%s%s %s\n", prefix, checkbox, option.name)
	}
	
	return view
}

// viewOutputOptions renders output-related options
func viewOutputOptions(m Model) string {
	var view string
	options := []struct {
		name  string
		desc  string
		value string
	}{
		{"Output file", "Custom output filename", m.backupOptions.OutputFile},
	}
	
	// Text input options
	for i, option := range options {
		prefix := "   "
		if i == m.optionSelected {
			prefix = " > "
		}
		
		value := option.value
		if value == "" {
			value = "(default)"
		}
		
		// Add edit hint for selected item
		editHint := ""
		if i == m.optionSelected {
			editHint = " (Press Space to edit)"
		}
		
		view += fmt.Sprintf("%s%s: %s%s\n", prefix, option.name, value, editHint)
		view += fmt.Sprintf("     %s\n", option.desc)
	}
	
	// Remove dry-run toggle as requested by user
	
	return view
}
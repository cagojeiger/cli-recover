package flags

import (
	"fmt"
	"reflect"
	"strings"
)

// init runs at startup to validate flag shortcuts
func init() {
	if err := ValidateNoDuplicates(); err != nil {
		panic(fmt.Sprintf("Flag shortcut conflict detected: %v", err))
	}
}

// ValidateNoDuplicates checks that no two flags share the same shortcut
func ValidateNoDuplicates() error {
	seen := make(map[string][]string) // shortcut -> list of field names

	// Check Registry struct
	v := reflect.ValueOf(Registry)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		shortcut := v.Field(i).String()

		if shortcut == "" {
			continue // Skip flags without shortcuts
		}

		seen[shortcut] = append(seen[shortcut], fieldName)
	}

	// Check for conflicts
	var conflicts []string
	for shortcut, fields := range seen {
		if len(fields) > 1 {
			conflicts = append(conflicts, fmt.Sprintf("'-%s' used by: %s",
				shortcut, strings.Join(fields, ", ")))
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("flag conflicts found:\n%s", strings.Join(conflicts, "\n"))
	}

	return nil
}

// GetShortcutReport generates a human-readable report of all flag shortcuts
func GetShortcutReport() string {
	var report strings.Builder
	report.WriteString("Flag Shortcut Registry Report\n")
	report.WriteString("=============================\n\n")

	// Group by shortcut
	shortcuts := make(map[string][]string)

	v := reflect.ValueOf(Registry)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		shortcut := v.Field(i).String()

		if shortcut != "" {
			shortcuts[shortcut] = append(shortcuts[shortcut], fieldName)
		}
	}

	// Sort and display
	report.WriteString("Shortcuts in use:\n")
	for shortcut := 'a'; shortcut <= 'z'; shortcut++ {
		key := string(shortcut)
		if fields, exists := shortcuts[key]; exists {
			report.WriteString(fmt.Sprintf("  -%s: %s\n", key, strings.Join(fields, ", ")))
		}
	}

	// Check uppercase
	for shortcut := 'A'; shortcut <= 'Z'; shortcut++ {
		key := string(shortcut)
		if fields, exists := shortcuts[key]; exists {
			report.WriteString(fmt.Sprintf("  -%s: %s\n", key, strings.Join(fields, ", ")))
		}
	}

	// List flags without shortcuts
	report.WriteString("\nFlags without shortcuts:\n")
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		shortcut := v.Field(i).String()

		if shortcut == "" {
			report.WriteString(fmt.Sprintf("  %s\n", fieldName))
		}
	}

	return report.String()
}

// IsShortcutAvailable checks if a shortcut is available for use
func IsShortcutAvailable(shortcut string) bool {
	if len(shortcut) != 1 {
		return false
	}

	v := reflect.ValueOf(Registry)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == shortcut {
			return false
		}
	}
	return true
}

// SuggestShortcut suggests an available shortcut based on the flag name
func SuggestShortcut(flagName string) string {
	// Try first letter
	firstLetter := strings.ToLower(string(flagName[0]))
	if IsShortcutAvailable(firstLetter) {
		return firstLetter
	}

	// Try uppercase first letter
	upperFirst := strings.ToUpper(firstLetter)
	if IsShortcutAvailable(upperFirst) {
		return upperFirst
	}

	// Try other letters in the name
	for _, char := range strings.ToLower(flagName[1:]) {
		shortcut := string(char)
		if IsShortcutAvailable(shortcut) {
			return shortcut
		}
	}

	// No available shortcut found
	return ""
}

package flags

import (
	"strings"
	"testing"
)

func TestValidateNoDuplicates(t *testing.T) {
	// This test runs the actual validation
	err := ValidateNoDuplicates()
	if err != nil {
		t.Errorf("Flag validation failed: %v", err)
	}
}

func TestGetShortcutReport(t *testing.T) {
	report := GetShortcutReport()

	// Check that report contains expected sections
	if !strings.Contains(report, "Flag Shortcut Registry Report") {
		t.Error("Report missing header")
	}

	if !strings.Contains(report, "Shortcuts in use:") {
		t.Error("Report missing shortcuts section")
	}

	// Check for some known shortcuts
	expectedShortcuts := []string{"-n:", "-v:", "-o:", "-T:", "-C:"}
	for _, expected := range expectedShortcuts {
		if !strings.Contains(report, expected) {
			t.Errorf("Report missing expected shortcut %s", expected)
		}
	}
}

func TestIsShortcutAvailable(t *testing.T) {
	tests := []struct {
		shortcut string
		want     bool
	}{
		{"n", false},  // Used by Namespace
		{"v", false},  // Used by Verbose
		{"x", true},   // Not used
		{"y", true},   // Not used
		{"T", false},  // Used by Totals
		{"C", false},  // Used by Container
		{"", false},   // Empty
		{"ab", false}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.shortcut, func(t *testing.T) {
			got := IsShortcutAvailable(tt.shortcut)
			if got != tt.want {
				t.Errorf("IsShortcutAvailable(%q) = %v, want %v", tt.shortcut, got, tt.want)
			}
		})
	}
}

func TestSuggestShortcut(t *testing.T) {
	tests := []struct {
		flagName string
		wantAny  []string // Any of these is acceptable
	}{
		{"xray", []string{"x"}},                     // First letter available
		{"yellow", []string{"y"}},                   // First letter available
		{"namespace", []string{"a", "m", "e", "N"}}, // 'n' taken, try others including uppercase
		{"verbose", []string{"r", "b", "e", "V"}},   // 'v' taken, try others including uppercase
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			got := SuggestShortcut(tt.flagName)

			// Check if the suggestion is valid
			if got != "" && !IsShortcutAvailable(got) {
				t.Errorf("SuggestShortcut(%q) = %q, but that shortcut is not available", tt.flagName, got)
			}

			// For flags that should get suggestions, verify it's reasonable
			if got != "" {
				found := false
				for _, acceptable := range tt.wantAny {
					if got == acceptable {
						found = true
						break
					}
				}
				if len(tt.wantAny) > 0 && !found {
					t.Errorf("SuggestShortcut(%q) = %q, expected one of %v", tt.flagName, got, tt.wantAny)
				}
			}
		})
	}
}

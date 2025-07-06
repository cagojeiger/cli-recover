package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"

	"github.com/cagojeiger/cli-recover/internal/runner"
	"github.com/cagojeiger/cli-recover/internal/tui"
)

func init() {
	// CI 환경에서 일관된 색상 프로파일 사용
	if os.Getenv("CI") == "true" {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

// Test TUI full output
func TestTUIFullOutput(t *testing.T) {
	// Use golden runner for predictable data
	os.Setenv("USE_GOLDEN", "true")
	runner := runner.NewRunner()
	tui.SetVersion("test")
	model := tui.InitialModel(runner)
	
	// Create test model with fixed terminal size
	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(50, 20),
	)
	
	// Wait for initial render
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Main Menu"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Send quit command (on main screen, q moves to exit, then enter quits)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for program to finish
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	
	// Get final output
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Fatal(err)
	}
	
	// Check that output contains expected content
	// The output may contain ANSI escape sequences, so we check for content
	outStr := string(out)
	if !strings.Contains(outStr, "Backup") {
		t.Errorf("Output should contain 'Backup' option, got:\n%s", outStr)
	}
	
	if !strings.Contains(outStr, "Exit") {
		t.Errorf("Output should contain 'Exit' option, got:\n%s", outStr)
	}
	
	// Exit should be selected (indicated by ">")
	if !strings.Contains(outStr, "> Exit") {
		t.Errorf("Exit should be selected with '>', got:\n%s", outStr)
	}
}

// Test navigation scenario
func TestTUINavigation(t *testing.T) {
	os.Setenv("USE_GOLDEN", "true")
	runner := runner.NewRunner()
	tui.SetVersion("test")
	model := tui.InitialModel(runner)
	
	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(50, 20),
	)
	
	// Wait for main menu
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Main Menu"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	
	// Press 'j' to move down
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	
	// Wait for cursor to move to Restore
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("> Restore"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	
	// Press 'k' to move up
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	
	// Wait for cursor to move back to Backup
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("> Backup"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	
	// Quit (q moves to exit, then enter quits)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	
	// Wait for program to finish
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// Test backup flow scenario
func TestBackupFlowScenario(t *testing.T) {
	os.Setenv("USE_GOLDEN", "true")
	runner := runner.NewRunner()
	tui.SetVersion("test")
	model := tui.InitialModel(runner)
	
	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(50, 25),
	)
	
	// Wait for main menu
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Main Menu"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	
	// Select Backup (press Enter)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	
	// Wait for namespace list and verify content
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Select Namespace") &&
				   (strings.Contains(output, "default") || strings.Contains(output, "production"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	
	// Select first namespace (Enter)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	
	// Wait for pod list and verify content
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Pods in default") &&
				   (strings.Contains(output, "nginx-") || strings.Contains(output, "mongo-"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	
	// Test back navigation
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	
	// Should be back to namespace list
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Select Namespace"))
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	
	// Quit (q moves to exit, then enter quits)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	
	// Wait for program to finish
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// Test directory browsing scenario
func TestDirectoryBrowsingScenario(t *testing.T) {
	os.Setenv("USE_GOLDEN", "true")
	runner := runner.NewRunner()
	tui.SetVersion("test")
	model := tui.InitialModel(runner)

	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 30),
	)

	// Wait for main menu
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return strings.Contains(string(bts), "Main Menu")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Select Backup
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for namespace list and select default
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Select Namespace") &&
				   strings.Contains(output, "default")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for pod list and select first pod
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Pods in default") &&
				   strings.Contains(output, "nginx-")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for directory browser (root directory)
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Browse: /") &&
				   strings.Contains(output, "var") &&
				   strings.Contains(output, "📁")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Select current directory (/) for backup using Space
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	// Wait for backup options screen
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Backup Options") &&
				   strings.Contains(output, "Compression")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Press Enter to go to command comparison
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for command comparison screen
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Command Comparison") &&
				   strings.Contains(output, "kubectl exec") &&
				   strings.Contains(output, "cli-restore backup")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Quit (q moves to exit, then enter quits)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for program to finish
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

// Test backup options functionality
func TestBackupOptionsScenario(t *testing.T) {
	os.Setenv("USE_GOLDEN", "true")
	runner := runner.NewRunner()
	tui.SetVersion("test")
	model := tui.InitialModel(runner)

	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 30),
	)

	// Navigate to backup options screen (abbreviated path)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Select Backup
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Select Namespace")
	}, teatest.WithCheckInterval(time.Millisecond*10), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Select default namespace
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Pods in default")
	}, teatest.WithCheckInterval(time.Millisecond*10), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Select first pod
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Browse: /")
	}, teatest.WithCheckInterval(time.Millisecond*10), teatest.WithDuration(time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}) // Select current path
	
	// Wait for backup options screen
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Backup Options") &&
				   strings.Contains(output, "[Compression]") &&
				   strings.Contains(output, "[x] gzip")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Test tab navigation to Excludes
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "[Excludes]") &&
				   strings.Contains(output, "*.log")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Test tab navigation to Advanced
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "[Advanced]") &&
				   strings.Contains(output, "Verbose output")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Go to final command screen
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			output := string(bts)
			return strings.Contains(output, "Command Comparison")
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second),
	)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
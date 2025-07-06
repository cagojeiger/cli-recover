package tui

import (
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
)

// HandleKey processes keyboard input
func HandleKey(m Model, key string) Model {
	// Special handling for text input mode
	if m.editMode {
		return handleTextInputKey(m, key)
	}
	
	// Special handling for execution screen
	if m.screen == ScreenExecuting {
		if len(m.executeOutput) > 0 && strings.Contains(m.executeOutput[len(m.executeOutput)-1], "Press any key") {
			// Any key continues from execution screen
			m.screen = ScreenMain
			m.selected = 0
			m.executeOutput = nil
		}
		return m
	}
	
	switch key {
	case "q", "ctrl+c":
		return handleQuit(m)
	case "j", "down":
		return handleDownNavigation(m)
	case "k", "up":
		return handleUpNavigation(m)
	case "enter":
		return handleEnter(m)
	case " ", "space":
		return handleSpace(m)
	case "tab":
		return handleTab(m)
	case "b", "esc":
		return handleBack(m)
	}
	
	return m
}

// handleTextInputKey processes keyboard input in text input mode
func handleTextInputKey(m Model, key string) Model {
	switch key {
	case "enter":
		// Save the input
		return saveTextInput(m)
	case "esc":
		// Cancel the input
		return cancelTextInput(m)
	default:
		// Pass the key to the text input component
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		_ = cmd // We don't need to handle the command for now
		return m
	}
}

func handleQuit(m Model) Model {
	if m.screen == ScreenMain && m.selected == 2 {
		m.quit = true
		return m
	}
	if m.screen != ScreenMain {
		m.quit = true
		return m
	}
	m.selected = 2
	return m
}

func handleDownNavigation(m Model) Model {
	if m.screen == ScreenBackupOptions {
		m.optionSelected++
		maxItems := getMaxItems(m)
		if m.optionSelected >= maxItems {
			m.optionSelected = maxItems - 1
		}
	} else {
		m.selected++
		maxItems := getMaxItems(m)
		if m.selected >= maxItems {
			m.selected = maxItems - 1
		}
	}
	return m
}

func handleUpNavigation(m Model) Model {
	if m.screen == ScreenBackupOptions {
		m.optionSelected--
		if m.optionSelected < 0 {
			m.optionSelected = 0
		}
	} else {
		m.selected--
		if m.selected < 0 {
			m.selected = 0
		}
	}
	return m
}

func handleTab(m Model) Model {
	if m.screen != ScreenBackupOptions {
		return m
	}
	
	// Cycle through categories based on backup type
	numCategories := getNumCategories(m)
	m.optionCategory = (m.optionCategory + 1) % numCategories
	m.optionSelected = 0
	
	return m
}
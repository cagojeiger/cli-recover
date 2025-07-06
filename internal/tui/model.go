package tui

import (
	"fmt"
	"os"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/cagojeiger/cli-recover/internal/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/runner"
)

// Global debug flag
var debugMode bool

// SetDebug sets the global debug mode
func SetDebug(debug bool) {
	debugMode = debug
	if debug {
		// Create or append to debug log file
		logFile, err := os.OpenFile("cli-recover-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			fmt.Fprintf(logFile, "=== Debug session started ===\n")
			logFile.Close()
		}
	}
}

// debugLog writes a debug message to both console and log file
func debugLog(format string, args ...interface{}) {
	if !debugMode {
		return
	}
	
	message := fmt.Sprintf(format, args...)
	
	// Write to log file
	logFile, err := os.OpenFile("cli-recover-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		fmt.Fprintf(logFile, "TUI: %s\n", message)
		logFile.Close()
	}
}

// TUI Types and Constants
type Screen int

const (
	ScreenMain Screen = iota
	ScreenBackupType
	ScreenNamespaceList
	ScreenPodList
	ScreenContainerList
	ScreenDirectoryBrowser
	ScreenBackupOptions
	ScreenPathInput
	ScreenExecuting
)

// Model represents the TUI state
type Model struct {
	runner     runner.Runner
	screen     Screen
	selected   int
	namespaces []string
	pods       []kubernetes.Pod
	
	// Directory browsing
	currentPath string
	directories []kubernetes.DirectoryEntry
	
	// Backup configuration
	selectedBackupType string
	selectedNamespace  string
	selectedPod        string
	selectedContainer  string
	selectedPath       string
	backupOptions      kubernetes.BackupOptions
	
	// Backup options UI state
	optionCategory int // 0: compression, 1: excludes, 2: advanced
	optionSelected int // selected item within category
	
	// Command building
	commandBuilder *CommandBuilder
	executor       Executor
	
	// Execution state
	executeOutput []string
	
	// Text input state
	editMode      bool
	editField     string // "container", "output", "minio-endpoint", etc.
	textInput     textinput.Model
	originalValue string // for cancellation
	
	// UI state
	err    error
	width  int
	height int
	quit   bool
}

// InitialModel creates the initial TUI model
func InitialModel(runner runner.Runner) Model {
	cb := NewCommandBuilder()
	cb.SetAction("backup") // Default to backup action
	
	// Create executor
	executor, err := NewRealExecutor()
	if err != nil {
		// Create a model with error state
		return Model{
			runner:         runner,
			screen:         ScreenMain,
			selected:       0,
			commandBuilder: cb,
			err:            fmt.Errorf("failed to initialize: %w", err),
			backupOptions: kubernetes.BackupOptions{
				CompressionType: "gzip",
				ExcludePatterns: []string{}, // 기본적으로 아무것도 제외하지 않음
				ExcludeVCS:      false,     // 기본적으로 VCS 제외하지 않음
				Verbose:         false,     // 기본적으로 verbose 끔
				ShowTotals:      false,
				PreservePerms:   false,     // 기본적으로 권한 보존하지 않음
			},
		}
	}
	
	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 100
	ti.Width = 50
	
	return Model{
		runner:         runner,
		screen:         ScreenMain,
		selected:       0,
		commandBuilder: cb,
		executor:       executor,
		textInput:      ti,
		editMode:       false,
		backupOptions: kubernetes.BackupOptions{
			CompressionType: "gzip",
			ExcludePatterns: []string{}, // 기본적으로 아무것도 제외하지 않음
			ExcludeVCS:      false,     // 기본적으로 VCS 제외하지 않음
			Verbose:         false,     // 기본적으로 verbose 끔
			ShowTotals:      false,
			PreservePerms:   false,     // 기본적으로 권한 보존하지 않음
		},
	}
}

// Init is the Bubble Tea initialization function
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle text input in edit mode
		if m.editMode {
			// Update text input component
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			
			// Handle special keys
			switch msg.String() {
			case "enter":
				m = saveTextInput(m)
			case "esc":
				m = cancelTextInput(m)
			}
			
			return m, cmd
		}
		
		m = HandleKey(m, msg.String())
		// Check if we should quit
		if m.quit {
			return m, tea.Quit
		}
		return m, nil
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	
	return m, nil
}
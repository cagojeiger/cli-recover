package tui

import (
	"fmt"
	"os"
	
	tea "github.com/charmbracelet/bubbletea"

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
	selectedPath       string
	backupOptions      kubernetes.BackupOptions
	minioOptions       kubernetes.MinioBackupOptions
	mongoOptions       kubernetes.MongoBackupOptions
	
	// Backup options UI state
	optionCategory int // 0: compression, 1: excludes, 2: advanced
	optionSelected int // selected item within category
	
	// Command building
	commandBuilder *CommandBuilder
	executor       Executor
	
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
				ExcludePatterns: []string{"*.log", "tmp/*", ".git"},
				ExcludeVCS:      true,
				Verbose:         true,
				ShowTotals:      false,
				PreservePerms:   true,
			},
			minioOptions: kubernetes.MinioBackupOptions{
				Endpoint:  "http://localhost:9000",
				Recursive: true,
				Format:    "tar",
			},
			mongoOptions: kubernetes.MongoBackupOptions{
				Host:   "localhost:27017",
				AuthDB: "admin",
				Gzip:   true,
			},
		}
	}
	
	return Model{
		runner:         runner,
		screen:         ScreenMain,
		selected:       0,
		commandBuilder: cb,
		executor:       executor,
		backupOptions: kubernetes.BackupOptions{
			CompressionType: "gzip",
			ExcludePatterns: []string{"*.log", "tmp/*", ".git"},
			ExcludeVCS:      true,
			Verbose:         true,
			ShowTotals:      false,
			PreservePerms:   true,
		},
		minioOptions: kubernetes.MinioBackupOptions{
			Endpoint:  "http://localhost:9000",
			Recursive: true,
			Format:    "tar",
		},
		mongoOptions: kubernetes.MongoBackupOptions{
			Host:   "localhost:27017",
			AuthDB: "admin",
			Gzip:   true,
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
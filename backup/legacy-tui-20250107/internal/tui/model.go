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
	ScreenJobManager
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
	
	// Multi-backup support
	program      *tea.Program  // Reference to tea.Program for sending messages
	jobManager   *JobManager   // Replaces JobScheduler
	activeJobID  string        // Currently selected job in job manager
	screenStack  []Screen      // Screen history for navigation
	
	// Job Manager view state
	jobDetailView bool         // Whether showing job detail
	selectedJobIndex int       // Index in job list
	pendingBackupJob *BackupJob // Job waiting to be executed
	
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
			jobManager:     NewJobManager(3),
			screenStack:    make([]Screen, 0),
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
	
	// Create job manager with max 3 concurrent jobs
	jobManager := NewJobManager(3)
	
	return Model{
		runner:         runner,
		screen:         ScreenMain,
		selected:       0,
		commandBuilder: cb,
		executor:       executor,
		jobManager:     jobManager,
		screenStack:    make([]Screen, 0),
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
	// No background goroutines! Following Bubble Tea patterns
	// Job execution will be handled through tea.Cmd
	return nil
}

// SetProgram sets the tea.Program reference
// This should be called after creating the program but before running it
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
}

// pushScreen saves current screen and switches to new screen
func (m Model) pushScreen(screen Screen) Model {
	m.screenStack = append(m.screenStack, m.screen)
	m.screen = screen
	m.selected = 0 // Reset selection for new screen
	return m
}

// popScreen returns to previous screen
func (m Model) popScreen() Model {
	if len(m.screenStack) > 0 {
		m.screen = m.screenStack[len(m.screenStack)-1]
		m.screenStack = m.screenStack[:len(m.screenStack)-1]
		m.selected = 0 // Reset selection
	}
	return m
}

// hasActiveJobs checks if there are any active jobs
func (m Model) hasActiveJobs() bool {
	return len(m.jobManager.GetActive()) > 0
}

// getJobSummary returns a summary of job states
func (m Model) getJobSummary() string {
	active := len(m.jobManager.GetActive())
	queued := len(m.jobManager.GetQueued())
	
	if active == 0 && queued == 0 {
		return "No active jobs"
	}
	
	summary := fmt.Sprintf("%d active", active)
	if queued > 0 {
		summary += fmt.Sprintf(", %d queued", queued)
	}
	
	return summary
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Add panic recovery to prevent crashes
	defer func() {
		if r := recover(); r != nil {
			debugLog("PANIC in Update: %v", r)
			m.err = fmt.Errorf("internal error: %v", r)
			// Clear potentially problematic state
			m.jobDetailView = false
			m.activeJobID = ""
		}
	}()
	
	// Check if we have a pending backup job to execute
	if m.pendingBackupJob != nil {
		job := m.pendingBackupJob
		m.pendingBackupJob = nil
		
		// Execute the job if we can run more
		if m.jobManager.CanRunMore() {
			return m, func() tea.Msg {
				return JobExecuteMsg{Job: job}
			}
		}
		// Otherwise it's queued and will run later
	}
	
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
		
	// Multi-backup message handling
	case BackupSubmitMsg:
		return m.handleBackupSubmit(msg)
		
	case BackupStartMsg:
		return m.handleBackupStart(msg)
		
	case BackupProgressMsg:
		return m.handleBackupProgress(msg)
		
	case BackupCompleteMsg:
		return m.handleBackupComplete(msg)
		
	case BackupErrorMsg:
		return m.handleBackupError(msg)
		
	case BackupCancelMsg:
		return m.handleBackupCancel(msg)
		
	case JobExecuteMsg:
		return m.handleJobExecute(msg)
		
	case ScreenJobManagerMsg:
		m = showJobManager(m)
		return m, nil
		
	case NavigateBackMsg:
		m = m.popScreen()
		return m, nil
		
	case JobDetailMsg:
		m.jobDetailView = true
		m.activeJobID = msg.JobID
		return m, nil
		
	case RefreshMsg:
		// Check for next job to run
		return m, waitForNextJobCmd(m.jobManager, m.program)
		
	case JobListUpdateMsg:
		// Force refresh
		return m, nil
	}
	
	return m, nil
}
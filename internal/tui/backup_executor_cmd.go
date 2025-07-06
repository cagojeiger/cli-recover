package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// executeBackupCmd creates a tea.Cmd that executes a backup job
// This follows the Bubble Tea pattern - no goroutines outside of tea.Cmd
func executeBackupCmd(job *BackupJob, program *tea.Program) tea.Cmd {
	return func() tea.Msg {
		// Get the path of the currently running executable (cli-recover itself)
		selfPath, err := os.Executable()
		if err != nil {
			return BackupErrorMsg{
				JobID: job.ID,
				Error: fmt.Errorf("cannot find self executable: %w", err),
			}
		}
		
		// Resolve any symlinks to get the real path
		selfPath, err = filepath.EvalSymlinks(selfPath)
		if err != nil {
			return BackupErrorMsg{
				JobID: job.ID,
				Error: fmt.Errorf("cannot resolve executable path: %w", err),
			}
		}

		// Parse command arguments (job.Command contains the arguments without "cli-recover")
		args := strings.Fields(job.Command)
		if len(args) == 0 {
			return BackupErrorMsg{
				JobID: job.ID,
				Error: fmt.Errorf("invalid command: empty"),
			}
		}

		debugLog("executeBackupCmd: starting %s with args: %v", selfPath, args)

		// Create command with context using self path
		cmd := exec.CommandContext(job.Context(), selfPath, args...)
		
		// Set process group for proper cleanup
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		// Get stdout and stderr pipes
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return BackupErrorMsg{
				JobID: job.ID,
				Error: fmt.Errorf("failed to create stdout pipe: %w", err),
			}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return BackupErrorMsg{
				JobID: job.ID,
				Error: fmt.Errorf("failed to create stderr pipe: %w", err),
			}
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return BackupErrorMsg{
				JobID: job.ID,
				Error: fmt.Errorf("failed to start command: %w", err),
			}
		}

		// Send start message
		program.Send(BackupStartMsg{
			JobID:   job.ID,
			Command: job.Command,
		})

		// Create channels for output collection
		outputDone := make(chan bool, 2)
		
		// Read stdout (in goroutine within tea.Cmd is OK)
		go func() {
			defer func() { outputDone <- true }()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				debugLog("stdout: %s", line)
				
				// Send progress update
				program.Send(BackupProgressMsg{
					JobID:     job.ID,
					Output:    line,
					Progress:  parseProgress(line),
					Timestamp: time.Now(),
				})
			}
			if err := scanner.Err(); err != nil {
				debugLog("stdout scanner error: %v", err)
			}
		}()

		// Read stderr
		go func() {
			defer func() { outputDone <- true }()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				debugLog("stderr: %s", line)
				
				// Send error output as progress
				program.Send(BackupProgressMsg{
					JobID:     job.ID,
					Output:    "[ERROR] " + line,
					Progress:  -1,
					Timestamp: time.Now(),
				})
			}
			if err := scanner.Err(); err != nil {
				debugLog("stderr scanner error: %v", err)
			}
		}()

		// Wait for output readers to finish
		<-outputDone
		<-outputDone

		// Wait for command to complete
		err = cmd.Wait()
		
		debugLog("executeBackupCmd: command completed with err=%v", err)

		// Check if cancelled
		if job.Context().Err() != nil {
			return BackupCompleteMsg{
				JobID:    job.ID,
				Success:  false,
				Error:    fmt.Errorf("job cancelled"),
				Duration: time.Since(job.StartTime),
			}
		}

		// Return completion message
		return BackupCompleteMsg{
			JobID:    job.ID,
			Success:  err == nil,
			Error:    err,
			Duration: time.Since(job.StartTime),
		}
	}
}

// monitorJobCmd creates a tea.Cmd that monitors a job's context for cancellation
func monitorJobCmd(job *BackupJob) tea.Cmd {
	return func() tea.Msg {
		<-job.Context().Done()
		return BackupCancelMsg{JobID: job.ID}
	}
}

// waitForNextJobCmd creates a tea.Cmd that checks for the next queued job
func waitForNextJobCmd(manager *JobManager, program *tea.Program) tea.Cmd {
	return func() tea.Msg {
		// Small delay to avoid busy loop
		time.Sleep(100 * time.Millisecond)
		
		// Check if we can run more jobs
		if manager.CanRunMore() {
			if nextJob := manager.GetNextQueued(); nextJob != nil {
				return JobExecuteMsg{Job: nextJob}
			}
		}
		
		return RefreshMsg{}
	}
}

// parseCommand is no longer used - keeping for reference
// We now use strings.Fields() for simple argument splitting
// and os.Executable() to get the cli-recover path

// parseProgress extracts progress percentage from output line
func parseProgress(line string) int {
	// Look for common progress patterns
	patterns := []struct {
		prefix  string
		suffix  string
		extract func(string) int
	}{
		// "50%"
		{"", "%", func(s string) int {
			var pct int
			fmt.Sscanf(s, "%d%%", &pct)
			return pct
		}},
		// "Progress: 50%"
		{"Progress:", "%", func(s string) int {
			var pct int
			fmt.Sscanf(s, "Progress: %d%%", &pct)
			return pct
		}},
		// "[50%]"
		{"[", "%]", func(s string) int {
			var pct int
			fmt.Sscanf(s, "[%d%%]", &pct)
			return pct
		}},
		// "50/100"
		{"", "/", func(s string) int {
			var current, total int
			if _, err := fmt.Sscanf(s, "%d/%d", &current, &total); err == nil && total > 0 {
				return (current * 100) / total
			}
			return -1
		}},
	}

	for _, pattern := range patterns {
		if strings.Contains(line, pattern.suffix) {
			if pct := pattern.extract(line); pct >= 0 && pct <= 100 {
				return pct
			}
		}
	}

	return -1
}

// killProcessGroup attempts to kill a process and all its children
func killProcessGroup(cmd *exec.Cmd, signal syscall.Signal) error {
	if cmd.Process == nil {
		return fmt.Errorf("process not started")
	}

	pid := cmd.Process.Pid
	
	// Try to kill the process group
	if err := syscall.Kill(-pid, signal); err != nil {
		// If process group kill fails, try killing just the process
		if err := cmd.Process.Signal(signal); err != nil {
			return fmt.Errorf("failed to send signal: %w", err)
		}
	}
	
	return nil
}

// forceKillCmd creates a tea.Cmd that force kills a process after timeout
func forceKillCmd(cmd *exec.Cmd, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(timeout)
		
		// Force kill with SIGKILL
		if err := killProcessGroup(cmd, syscall.SIGKILL); err != nil {
			debugLog("force kill failed: %v", err)
		}
		
		return nil
	}
}

// Helper to create a safe copy of output for display
func sanitizeOutput(output string) string {
	// Remove control characters except newline
	var result strings.Builder
	for _, r := range output {
		if r == '\n' || (r >= 32 && r < 127) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
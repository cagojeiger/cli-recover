package tui

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// AsyncBackupExecutor executes backup jobs asynchronously
type AsyncBackupExecutor struct {
	processManager ProcessManager
	program        *tea.Program
}

// NewAsyncBackupExecutor creates a new async backup executor
func NewAsyncBackupExecutor(program *tea.Program) *AsyncBackupExecutor {
	return &AsyncBackupExecutor{
		processManager: NewRealProcessManager(),
		program:        program,
	}
}

// Execute runs a backup job and sends progress updates
func (e *AsyncBackupExecutor) Execute(job *BackupJob) error {
	// Parse command
	args := parseBackupCommand(job.Command)
	if len(args) == 0 {
		return fmt.Errorf("invalid command")
	}

	// Start the process
	cmd, err := e.processManager.Start(job.Context(), args[0], args[1:])
	if err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Set up output streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start output readers
	go e.streamOutput(job, stdout, false)
	go e.streamOutput(job, stderr, true)

	// Wait for process to complete
	err = cmd.Wait()
	
	// Handle cancellation
	if job.Context().Err() != nil {
		return fmt.Errorf("job cancelled")
	}

	return err
}

// streamOutput reads from a pipe and sends progress updates
func (e *AsyncBackupExecutor) streamOutput(job *BackupJob, reader io.Reader, isError bool) {
	scanner := bufio.NewScanner(reader)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Add to job output
		job.AddOutput(line)
		
		// Parse progress if possible
		percent := parseProgress(line)
		if percent >= 0 {
			job.UpdateProgress(percent, line)
		}
		
		// Send update to TUI
		if e.program != nil {
			e.program.Send(BackupProgressMsg{
				JobID:   job.ID,
				Output:  line,
				Percent: job.GetProgress(),
			})
		}
	}
}

// ExecuteBackupAsync creates a tea.Cmd that executes a backup asynchronously
func ExecuteBackupAsync(job *BackupJob, executor JobExecutor) tea.Cmd {
	return func() tea.Msg {
		// Start the job
		err := job.Start()
		if err != nil {
			return BackupCompleteMsg{
				JobID:   job.ID,
				Success: false,
				Error:   err,
			}
		}

		// Execute the backup
		err = executor.Execute(job)
		
		// Mark job as complete
		job.Complete(err)

		// Return completion message
		return BackupCompleteMsg{
			JobID:   job.ID,
			Success: err == nil,
			Error:   err,
		}
	}
}

// parseBackupCommand parses a command string into arguments
func parseBackupCommand(command string) []string {
	// Simple parsing - in production, use a proper shell parser
	args := strings.Fields(command)
	
	// Handle quoted strings (basic implementation)
	var result []string
	var current string
	inQuote := false
	
	for _, arg := range args {
		if strings.HasPrefix(arg, "\"") {
			inQuote = true
			current = strings.TrimPrefix(arg, "\"")
			if strings.HasSuffix(current, "\"") && len(current) > 0 {
				result = append(result, strings.TrimSuffix(current, "\""))
				current = ""
				inQuote = false
			}
		} else if inQuote {
			if strings.HasSuffix(arg, "\"") {
				current += " " + strings.TrimSuffix(arg, "\"")
				result = append(result, current)
				current = ""
				inQuote = false
			} else {
				current += " " + arg
			}
		} else {
			result = append(result, arg)
		}
	}
	
	return result
}

// parseProgress tries to extract progress percentage from output line
func parseProgress(line string) int {
	// Look for common progress patterns
	// Examples: "50%", "Progress: 50%", "[50%]", etc.
	
	patterns := []string{
		`(\d+)%`,           // Simple percentage
		`\[(\d+)%\]`,       // Bracketed percentage
		`Progress:\s*(\d+)%`, // Progress label
	}
	
	for _, pattern := range patterns {
		if matches := findStringSubmatch(pattern, line); len(matches) > 1 {
			var percent int
			fmt.Sscanf(matches[1], "%d", &percent)
			return percent
		}
	}
	
	return -1 // No progress found
}

// findStringSubmatch is a simple regex-like matcher
// In production, use regexp package
func findStringSubmatch(pattern, text string) []string {
	// This is a simplified implementation
	// Real implementation would use regexp
	
	// For now, just look for simple percentage
	for i := 0; i < len(text); i++ {
		if i+1 < len(text) && text[i] >= '0' && text[i] <= '9' {
			j := i
			for j < len(text) && text[j] >= '0' && text[j] <= '9' {
				j++
			}
			if j < len(text) && text[j] == '%' {
				return []string{text[i:j+1], text[i:j]}
			}
		}
	}
	
	return nil
}
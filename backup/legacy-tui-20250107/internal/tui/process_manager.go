package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// debugLogPM is a simple logger for debugging
func debugLogPM(format string, args ...interface{}) {
	// For now, just print to stdout when needed
	// In production, this would check a debug flag
	// fmt.Printf("PM: "+format+"\n", args...)
}

// ProcessManager handles process lifecycle and output
type ProcessManager interface {
	Start(ctx context.Context, cmd string, args []string) (*exec.Cmd, error)
	Wait(cmd *exec.Cmd) error
	Kill(cmd *exec.Cmd, force bool) error
	ReadOutput(onOutput func(string)) error
}

// RealProcessManager implements ProcessManager for actual process execution
type RealProcessManager struct {
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// NewRealProcessManager creates a new process manager
func NewRealProcessManager() *RealProcessManager {
	return &RealProcessManager{}
}

// Start starts a new process with proper group handling
func (pm *RealProcessManager) Start(ctx context.Context, cmd string, args []string) (*exec.Cmd, error) {
	command := exec.CommandContext(ctx, cmd, args...)
	
	// Set up process group for proper cleanup
	if runtime.GOOS != "windows" {
		command.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true, // Create new process group
		}
	}
	
	// Set up pipes for stdout and stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	pm.stdout = stdout
	
	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	pm.stderr = stderr
	
	// Start the command
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}
	
	return command, nil
}

// Wait waits for the process to complete
func (pm *RealProcessManager) Wait(cmd *exec.Cmd) error {
	return cmd.Wait()
}

// Kill terminates the process and its children
func (pm *RealProcessManager) Kill(cmd *exec.Cmd, force bool) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	
	if runtime.GOOS == "windows" {
		// On Windows, just kill the process directly
		return cmd.Process.Kill()
	}
	
	// On Unix-like systems, kill the entire process group
	pid := cmd.Process.Pid
	
	if !force {
		// Try graceful termination first (SIGTERM)
		if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
			// If process group kill fails, try killing just the process
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				return fmt.Errorf("failed to send SIGTERM: %w", err)
			}
		}
		
		// Give process time to terminate gracefully
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		
		select {
		case <-done:
			// Process terminated gracefully
			return nil
		case <-time.After(2 * time.Second):
			// Process didn't terminate in time, fall through to force kill
		}
	}
	
	// Force kill (SIGKILL) the entire process group
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		// If process group kill fails, try killing just the process
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	
	// Wait for process to actually terminate
	cmd.Wait()
	
	return nil
}

// ReadOutput reads stdout and stderr, calling onOutput for each line
func (pm *RealProcessManager) ReadOutput(onOutput func(string)) error {
	if pm.stdout == nil && pm.stderr == nil {
		return fmt.Errorf("no output pipes available")
	}
	
	// Read stdout
	if pm.stdout != nil {
		go pm.readPipe(pm.stdout, onOutput)
	}
	
	// Read stderr
	if pm.stderr != nil {
		go pm.readPipe(pm.stderr, onOutput)
	}
	
	return nil
}

// readPipe reads from a pipe and calls onOutput for each line
func (pm *RealProcessManager) readPipe(pipe io.ReadCloser, onOutput func(string)) {
	defer pipe.Close()
	
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		if onOutput != nil {
			onOutput(line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		debugLogPM("Error reading pipe: %v", err)
	}
}

// SafeKillProcess provides a reliable way to terminate a process
func SafeKillProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	
	// Stage 1: Send SIGTERM for graceful shutdown
	if runtime.GOOS != "windows" {
		cmd.Process.Signal(syscall.SIGTERM)
		
		// Wait up to 2 seconds for graceful shutdown
		done := make(chan bool)
		go func() {
			cmd.Wait()
			done <- true
		}()
		
		select {
		case <-done:
			return nil
		case <-time.After(2 * time.Second):
			// Continue to force kill
		}
	}
	
	// Stage 2: Force kill
	if err := cmd.Process.Kill(); err != nil {
		// Process might already be dead
		if !isProcessAlive(cmd.Process.Pid) {
			return nil
		}
		return fmt.Errorf("failed to kill process: %w", err)
	}
	
	// Stage 3: Kill process group (Unix only)
	if runtime.GOOS != "windows" {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	
	// Wait for process to terminate
	cmd.Wait()
	
	return nil
}

// isProcessAlive checks if a process is still running
func isProcessAlive(pid int) bool {
	if runtime.GOOS == "windows" {
		// On Windows, we can't easily check, so assume it's alive
		return true
	}
	
	// On Unix, sending signal 0 checks if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}
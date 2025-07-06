package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Executor interface for command execution
type Executor interface {
	Execute(args []string, writer io.Writer) error
}

// RealExecutor executes actual cli-recover commands
type RealExecutor struct{
	selfPath string
}

// NewRealExecutor creates a new real executor
func NewRealExecutor() (*RealExecutor, error) {
	// Get the path of the currently running executable
	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot find self executable: %w", err)
	}
	
	// Resolve any symlinks to get the real path
	selfPath, err = filepath.EvalSymlinks(selfPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve executable path: %w", err)
	}
	
	debugLog("RealExecutor: using self path: %s", selfPath)
	
	return &RealExecutor{selfPath: selfPath}, nil
}

// Execute runs the cli-recover command with given arguments
func (r *RealExecutor) Execute(args []string, writer io.Writer) error {
	debugLog("RealExecutor.Execute: %s %v", r.selfPath, args)
	
	// Create the command using self path
	cmd := exec.Command(r.selfPath, args...)
	
	// Set up output handling
	if writer != nil {
		cmd.Stdout = writer
		cmd.Stderr = writer
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	// Execute the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	
	debugLog("RealExecutor.Execute: completed successfully")
	return nil
}

// StreamingExecutor executes commands with real-time output streaming
type StreamingExecutor struct {
	selfPath string
	OnOutput func(line string)
}

// NewStreamingExecutor creates a new streaming executor
func NewStreamingExecutor(onOutput func(line string)) (*StreamingExecutor, error) {
	// Get the path of the currently running executable
	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot find self executable: %w", err)
	}
	
	// Resolve any symlinks to get the real path
	selfPath, err = filepath.EvalSymlinks(selfPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve executable path: %w", err)
	}
	
	return &StreamingExecutor{
		selfPath: selfPath,
		OnOutput: onOutput,
	}, nil
}

// Execute runs the command with streaming output
func (s *StreamingExecutor) Execute(args []string, writer io.Writer) error {
	debugLog("StreamingExecutor.Execute: %s %v", s.selfPath, args)
	
	cmd := exec.Command(s.selfPath, args...)
	
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	
	// Read output in real-time
	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			output := string(buf[:n])
			if writer != nil {
				writer.Write(buf[:n])
			}
			if s.OnOutput != nil {
				s.OnOutput(output)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stdout: %w", err)
		}
	}
	
	// Also read stderr
	stderrBuf := make([]byte, 1024)
	for {
		n, err := stderr.Read(stderrBuf)
		if n > 0 {
			if writer != nil {
				writer.Write(stderrBuf[:n])
			}
		}
		if err == io.EOF {
			break
		}
	}
	
	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	
	debugLog("StreamingExecutor.Execute: completed successfully")
	return nil
}
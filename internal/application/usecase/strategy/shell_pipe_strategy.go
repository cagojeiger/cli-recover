package strategy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
)

// ShellPipeStrategy executes pipeline using Unix pipes
type ShellPipeStrategy struct {
	LogDir        string // Directory for logging (optional)
	CaptureOutput bool   // Whether to capture output
	Output        string // Captured output
}

// Execute implements ExecutionStrategy for shell pipes
func (s *ShellPipeStrategy) Execute(pipeline *entity.Pipeline) error {
	// Validate that pipeline is linear
	if !isSimpleLinear(pipeline) {
		return fmt.Errorf("non-linear pipeline cannot be executed with shell pipes")
	}

	// Determine execution mode
	if s.LogDir != "" {
		return s.executeWithLogging(pipeline)
	}
	return s.executeSimple(pipeline)
}

// executeSimple runs the pipeline without logging
func (s *ShellPipeStrategy) executeSimple(pipeline *entity.Pipeline) error {
	// Build shell command
	shellCmd, err := buildShellCommand(pipeline)
	if err != nil {
		return fmt.Errorf("failed to build shell command: %w", err)
	}

	// Execute using bash
	cmd := exec.Command("bash", "-c", shellCmd)
	
	// Handle output capture
	if s.CaptureOutput {
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
		
		s.Output = out.String()
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	return nil
}

// executeWithLogging runs the pipeline with logging enabled
func (s *ShellPipeStrategy) executeWithLogging(pipeline *entity.Pipeline) error {
	// Ensure log directory exists
	if err := os.MkdirAll(s.LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Build shell script with logging
	shellScript, err := buildShellCommandWithLogging(pipeline, s.LogDir)
	if err != nil {
		return fmt.Errorf("failed to build shell script: %w", err)
	}

	// Create temporary script file
	scriptFile := filepath.Join(s.LogDir, "pipeline.sh")
	if err := os.WriteFile(scriptFile, []byte(shellScript), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}
	defer os.Remove(scriptFile)

	// Execute the script
	cmd := exec.Command("bash", scriptFile)
	
	if s.CaptureOutput {
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pipeline execution failed: %w", err)
		}
		
		s.Output = out.String()
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pipeline execution failed: %w", err)
		}
	}

	return nil
}
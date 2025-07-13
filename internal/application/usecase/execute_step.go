package usecase

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/cagojeiger/cli-recover/internal/domain/service"
)

// ExecuteStep handles the execution of a single pipeline step
type ExecuteStep struct {
	streamManager *service.StreamManager
	logWriter     io.Writer
}

// NewExecuteStep creates a new ExecuteStep use case
func NewExecuteStep(streamManager *service.StreamManager) *ExecuteStep {
	return &ExecuteStep{
		streamManager: streamManager,
		logWriter:     io.Discard, // Default to discarding logs
	}
}

// SetLogWriter sets the writer for capturing logs
func (e *ExecuteStep) SetLogWriter(w io.Writer) {
	e.logWriter = w
}

// Execute runs a single step
func (e *ExecuteStep) Execute(step *entity.Step) error {
	// Setup output first so reader can be created before command starts
	var outputWriter io.WriteCloser
	if step.Output != "" {
		writer, err := e.streamManager.CreateStream(step.Output)
		if err != nil {
			return fmt.Errorf("failed to create output stream '%s': %w", step.Output, err)
		}
		outputWriter = writer
		defer func() {
			if outputWriter != nil {
				outputWriter.Close()
			}
		}()
	}
	
	// Create command
	cmd := exec.Command("sh", "-c", step.Run)
	
	// Setup input if specified
	if step.Input != "" {
		reader, err := e.streamManager.GetStream(step.Input)
		if err != nil {
			return fmt.Errorf("failed to get input stream '%s': %w", step.Input, err)
		}
		cmd.Stdin = reader
	}
	
	// Setup output
	if outputWriter != nil {
		// If we have a log writer, tee stdout to both output and log
		if e.logWriter != io.Discard {
			cmd.Stdout = io.MultiWriter(outputWriter, e.logWriter)
		} else {
			cmd.Stdout = outputWriter
		}
	} else if e.logWriter != io.Discard {
		// No output stream, but we have a log writer
		cmd.Stdout = e.logWriter
	}
	
	// Always capture stderr to log writer
	if e.logWriter != io.Discard {
		cmd.Stderr = e.logWriter
	}
	
	// Run the command
	err := cmd.Run()
	
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	
	return nil
}
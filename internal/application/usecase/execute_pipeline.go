package usecase

import (
	"fmt"
	"io"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/cagojeiger/cli-recover/internal/domain/service"
)

// ExecutePipeline handles the execution of an entire pipeline
type ExecutePipeline struct {
	stepExecutor  *ExecuteStep
	streamManager *service.StreamManager
	logWriter     io.Writer
}

// NewExecutePipeline creates a new ExecutePipeline use case
func NewExecutePipeline(stepExecutor *ExecuteStep, streamManager *service.StreamManager) *ExecutePipeline {
	return &ExecutePipeline{
		stepExecutor:  stepExecutor,
		streamManager: streamManager,
		logWriter:     io.Discard,
	}
}

// SetLogWriter sets the writer for capturing logs
func (e *ExecutePipeline) SetLogWriter(w io.Writer) {
	e.logWriter = w
	// Also set log writer for step executor
	e.stepExecutor.SetLogWriter(w)
}

// Execute runs the entire pipeline
func (e *ExecutePipeline) Execute(pipeline *entity.Pipeline) error {
	// Validate pipeline first
	if err := pipeline.Validate(); err != nil {
		return fmt.Errorf("pipeline validation failed: %w", err)
	}
	
	// Log pipeline execution start
	if e.logWriter != io.Discard {
		fmt.Fprintf(e.logWriter, "Executing pipeline: %s\n", pipeline.Name)
		if pipeline.Description != "" {
			fmt.Fprintf(e.logWriter, "Description: %s\n", pipeline.Description)
		}
	}
	
	// Execute steps sequentially
	for i, step := range pipeline.Steps {
		if e.logWriter != io.Discard {
			fmt.Fprintf(e.logWriter, "\n[Step %d/%d] %s\n", i+1, len(pipeline.Steps), step.Name)
			fmt.Fprintf(e.logWriter, "Command: %s\n", step.Run)
			if step.Input != "" {
				fmt.Fprintf(e.logWriter, "Input: %s\n", step.Input)
			}
			if step.Output != "" {
				fmt.Fprintf(e.logWriter, "Output: %s\n", step.Output)
			}
		}
		
		// Execute the step
		if err := e.stepExecutor.Execute(step); err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}
		
		if e.logWriter != io.Discard {
			fmt.Fprintf(e.logWriter, "Step '%s' completed successfully\n", step.Name)
		}
	}
	
	// Don't close all streams here - let the caller handle cleanup
	// This allows reading from output streams after pipeline execution
	
	if e.logWriter != io.Discard {
		fmt.Fprintf(e.logWriter, "\nPipeline '%s' completed successfully\n", pipeline.Name)
	}
	
	return nil
}
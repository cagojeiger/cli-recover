package usecase

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/cagojeiger/cli-recover/internal/domain/service"
)

func TestExecutePipeline_SimpleExecution(t *testing.T) {
	// Setup
	sm := service.NewStreamManager()
	stepExecutor := NewExecuteStep(sm)
	executor := NewExecutePipeline(stepExecutor, sm)
	
	// Create a simple pipeline without output
	pipeline, _ := entity.NewPipeline("test", "Test pipeline")
	
	step1, _ := entity.NewStep("echo", "echo hello > /dev/null")
	// Don't set output - just run the command
	pipeline.AddStep(step1)
	
	// Create log buffer
	var logBuffer bytes.Buffer
	executor.SetLogWriter(&logBuffer)
	
	// Execute
	err := executor.Execute(pipeline)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	// Check logs
	logs := logBuffer.String()
	t.Logf("Logs: %s", logs)
	if !strings.Contains(logs, "Pipeline 'test' completed successfully") {
		t.Error("Logs should indicate successful completion")
	}
}

func TestExecutePipeline_EmptyPipeline(t *testing.T) {
	// Setup
	sm := service.NewStreamManager()
	stepExecutor := NewExecuteStep(sm)
	executor := NewExecutePipeline(stepExecutor, sm)
	
	// Create empty pipeline
	pipeline, _ := entity.NewPipeline("empty", "Empty pipeline")
	
	// Execute
	err := executor.Execute(pipeline)
	if err == nil {
		t.Error("Execute() should fail for empty pipeline")
	}
}

func TestExecutePipeline_InvalidPipeline(t *testing.T) {
	// Setup
	sm := service.NewStreamManager()
	stepExecutor := NewExecuteStep(sm)
	executor := NewExecutePipeline(stepExecutor, sm)
	
	// Create invalid pipeline
	pipeline, _ := entity.NewPipeline("invalid", "Invalid pipeline")
	step, _ := entity.NewStep("bad", "cat")
	step.SetInput("nonexistent")
	pipeline.AddStep(step)
	
	// Execute
	err := executor.Execute(pipeline)
	if err == nil {
		t.Error("Execute() should fail for invalid pipeline")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Error should mention validation, got: %v", err)
	}
}
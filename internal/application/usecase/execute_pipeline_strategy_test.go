package usecase

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/cagojeiger/cli-recover/internal/domain/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutePipeline_WithStrategy(t *testing.T) {
	// Setup
	sm := service.NewStreamManager()
	stepExecutor := NewExecuteStep(sm)
	executor := NewExecutePipeline(stepExecutor, sm)

	t.Run("simple linear pipeline uses shell strategy", func(t *testing.T) {
		// Create simple linear pipeline
		pipeline, _ := entity.NewPipeline("simple", "Simple pipeline")
		step1, _ := entity.NewStep("echo", "echo hello")
		step1.SetOutput("text")
		step2, _ := entity.NewStep("upper", "tr '[:lower:]' '[:upper:]'")
		step2.SetInput("text")
		pipeline.AddStep(step1)
		pipeline.AddStep(step2)

		// Execute with strategy selection
		options := ExecuteOptions{
			UseStrategy: true,
		}
		err := executor.ExecuteWithOptions(pipeline, options)
		assert.NoError(t, err)
	})

	t.Run("complex pipeline uses go stream strategy", func(t *testing.T) {
		// Create complex pipeline with branch
		pipeline, _ := entity.NewPipeline("complex", "Complex pipeline")
		step1, _ := entity.NewStep("source", "echo data")
		step1.SetOutput("data")
		step2, _ := entity.NewStep("branch1", "grep foo || true")
		step2.SetInput("data")
		step3, _ := entity.NewStep("branch2", "grep bar || true")
		step3.SetInput("data")
		pipeline.AddStep(step1)
		pipeline.AddStep(step2)
		pipeline.AddStep(step3)

		// Create log buffer to verify go stream execution
		var logBuffer bytes.Buffer
		executor.SetLogWriter(&logBuffer)

		// Execute with strategy selection
		options := ExecuteOptions{
			UseStrategy: true,
		}
		err := executor.ExecuteWithOptions(pipeline, options)
		assert.NoError(t, err)

		// Should see go stream style logging
		logs := logBuffer.String()
		assert.Contains(t, logs, "[Step 1/3]")
		assert.Contains(t, logs, "[Step 2/3]")
		assert.Contains(t, logs, "[Step 3/3]")
	})

	t.Run("shell strategy with logging", func(t *testing.T) {
		// Create temp directory for logs
		tempDir, err := os.MkdirTemp("", "cli-pipe-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create simple pipeline
		pipeline, _ := entity.NewPipeline("logged", "Logged pipeline")
		step1, _ := entity.NewStep("echo", "echo test output")
		step1.SetOutput("text")
		step2, _ := entity.NewStep("upper", "tr '[:lower:]' '[:upper:]'")
		step2.SetInput("text")
		pipeline.AddStep(step1)
		pipeline.AddStep(step2)

		// Execute with strategy and logging
		options := ExecuteOptions{
			UseStrategy: true,
			LogDir:      tempDir,
		}
		err = executor.ExecuteWithOptions(pipeline, options)
		assert.NoError(t, err)

		// Verify log files exist
		assert.FileExists(t, filepath.Join(tempDir, "echo.out"))
		assert.FileExists(t, filepath.Join(tempDir, "upper.out"))
		assert.FileExists(t, filepath.Join(tempDir, "pipeline.status"))

		// Check output content
		echoOut, _ := os.ReadFile(filepath.Join(tempDir, "echo.out"))
		assert.Equal(t, "test output\n", string(echoOut))

		upperOut, _ := os.ReadFile(filepath.Join(tempDir, "upper.out"))
		assert.Equal(t, "TEST OUTPUT\n", string(upperOut))
	})

	t.Run("force specific strategy", func(t *testing.T) {
		// Create simple pipeline
		pipeline, _ := entity.NewPipeline("forced", "Forced strategy")
		step, _ := entity.NewStep("echo", "echo hello")
		pipeline.AddStep(step)

		// Force go stream strategy even for simple pipeline
		options := ExecuteOptions{
			UseStrategy:   true,
			ForceStrategy: "go-stream",
		}

		var logBuffer bytes.Buffer
		executor.SetLogWriter(&logBuffer)

		err := executor.ExecuteWithOptions(pipeline, options)
		assert.NoError(t, err)

		// Should see go stream style logging
		logs := logBuffer.String()
		assert.Contains(t, logs, "[Step 1/1]")
	})

	t.Run("backward compatibility without strategy", func(t *testing.T) {
		// Create pipeline
		pipeline, _ := entity.NewPipeline("legacy", "Legacy execution")
		step, _ := entity.NewStep("echo", "echo legacy > /dev/null")
		pipeline.AddStep(step)

		var logBuffer bytes.Buffer
		executor.SetLogWriter(&logBuffer)

		// Execute without strategy (original behavior)
		err := executor.Execute(pipeline)
		assert.NoError(t, err)

		// Should see original logging style
		logs := logBuffer.String()
		assert.Contains(t, logs, "Pipeline 'legacy' completed successfully")
	})
}

func TestExecutePipeline_StrategyErrorHandling(t *testing.T) {
	// Setup
	sm := service.NewStreamManager()
	stepExecutor := NewExecuteStep(sm)
	executor := NewExecutePipeline(stepExecutor, sm)

	t.Run("handle shell strategy failure", func(t *testing.T) {
		// Create failing pipeline
		pipeline, _ := entity.NewPipeline("fail", "Failing pipeline")
		step, _ := entity.NewStep("fail", "false")
		pipeline.AddStep(step)

		options := ExecuteOptions{
			UseStrategy: true,
		}
		err := executor.ExecuteWithOptions(pipeline, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exit status")
	})

	t.Run("reject non-linear pipeline for shell strategy", func(t *testing.T) {
		// This should be rejected at strategy determination
		pipeline, _ := entity.NewPipeline("branch", "Branching pipeline")
		step1, _ := entity.NewStep("source", "echo data")
		step1.SetOutput("data")
		step2, _ := entity.NewStep("branch1", "cat")
		step2.SetInput("data")
		step3, _ := entity.NewStep("branch2", "cat")
		step3.SetInput("data")
		pipeline.AddStep(step1)
		pipeline.AddStep(step2)
		pipeline.AddStep(step3)

		// Force shell strategy on non-linear pipeline
		options := ExecuteOptions{
			UseStrategy:   true,
			ForceStrategy: "shell-pipe",
		}
		err := executor.ExecuteWithOptions(pipeline, options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-linear")
	})
}
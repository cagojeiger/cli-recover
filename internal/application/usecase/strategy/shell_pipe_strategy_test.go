package strategy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellPipeStrategy_Execute(t *testing.T) {
	strategy := &ShellPipeStrategy{}

	t.Run("execute simple pipeline", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "simple-test",
			Steps: []*entity.Step{
				{Name: "echo", Run: "echo hello world", Output: "text"},
				{Name: "upper", Run: "tr '[:lower:]' '[:upper:]'", Input: "text"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.NoError(t, err)
	})

	t.Run("execute single step", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "single-step",
			Steps: []*entity.Step{
				{Name: "list", Run: "ls -la /tmp"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.NoError(t, err)
	})

	t.Run("handle command failure", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "failing-pipeline",
			Steps: []*entity.Step{
				{Name: "fail", Run: "false"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exit status")
	})

	t.Run("handle piped command failure", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "failing-pipe",
			Steps: []*entity.Step{
				{Name: "echo", Run: "echo test", Output: "text"},
				{Name: "fail", Run: "grep -q nonexistent && false", Input: "text"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.Error(t, err)
	})

	t.Run("reject non-linear pipeline", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "branching",
			Steps: []*entity.Step{
				{Name: "source", Run: "echo data", Output: "data"},
				{Name: "branch1", Run: "grep foo", Input: "data"},
				{Name: "branch2", Run: "grep bar", Input: "data"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-linear")
	})
}

func TestShellPipeStrategy_ExecuteWithLogging(t *testing.T) {
	// Create temporary directory for logs
	tempDir, err := os.MkdirTemp("", "cli-pipe-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	strategy := &ShellPipeStrategy{
		LogDir: tempDir,
	}

	t.Run("execute with logging enabled", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "logged-pipeline",
			Steps: []*entity.Step{
				{Name: "echo", Run: "echo 'test output'", Output: "text"},
				{Name: "upper", Run: "tr '[:lower:]' '[:upper:]'", Input: "text"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.NoError(t, err)

		// Check log files were created
		echoOut := filepath.Join(tempDir, "echo.out")
		upperOut := filepath.Join(tempDir, "upper.out")
		statusFile := filepath.Join(tempDir, "pipeline.status")

		assert.FileExists(t, echoOut)
		assert.FileExists(t, upperOut)
		assert.FileExists(t, statusFile)

		// Verify output content
		echoContent, err := os.ReadFile(echoOut)
		assert.NoError(t, err)
		assert.Equal(t, "test output\n", string(echoContent))

		upperContent, err := os.ReadFile(upperOut)
		assert.NoError(t, err)
		assert.Equal(t, "TEST OUTPUT\n", string(upperContent))

		// Check status file
		statusContent, err := os.ReadFile(statusFile)
		assert.NoError(t, err)
		assert.Contains(t, string(statusContent), "Pipeline exit code: 0")
	})

	t.Run("capture stderr in logs", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "stderr-test",
			Steps: []*entity.Step{
				{Name: "stderr", Run: "echo 'error message' >&2 && echo 'normal output'"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.NoError(t, err)

		// Check stderr was captured
		stderrFile := filepath.Join(tempDir, "stderr.err")
		assert.FileExists(t, stderrFile)

		stderrContent, err := os.ReadFile(stderrFile)
		assert.NoError(t, err)
		assert.Equal(t, "error message\n", string(stderrContent))
	})

	t.Run("log failed pipeline status", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "fail-logged",
			Steps: []*entity.Step{
				{Name: "fail", Run: "exit 42"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.Error(t, err)

		// Check status file shows failure
		statusFile := filepath.Join(tempDir, "pipeline.status")
		statusContent, err := os.ReadFile(statusFile)
		assert.NoError(t, err)
		assert.Contains(t, string(statusContent), "Pipeline exit code: 42")
	})
}

func TestShellPipeStrategy_OutputCapture(t *testing.T) {
	strategy := &ShellPipeStrategy{
		CaptureOutput: true,
	}

	t.Run("capture pipeline output", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "capture-test",
			Steps: []*entity.Step{
				{Name: "echo", Run: "echo 'captured'", Output: "text"},
				{Name: "trim", Run: "tr -d '\n'", Input: "text"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.NoError(t, err)
		assert.Equal(t, "captured", strings.TrimSpace(strategy.Output))
	})

	t.Run("capture single step output", func(t *testing.T) {
		pipeline := &entity.Pipeline{
			Name: "single-capture",
			Steps: []*entity.Step{
				{Name: "echo", Run: "echo 'single output'"},
			},
		}

		err := strategy.Execute(pipeline)
		assert.NoError(t, err)
		assert.Equal(t, "single output", strings.TrimSpace(strategy.Output))
	})
}
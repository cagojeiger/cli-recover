package pipeline

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	// Test default executor
	e1 := NewExecutor()
	assert.NotNil(t, e1)
	assert.Equal(t, "", e1.logDir)
	assert.Equal(t, os.Stdout, e1.logWriter)

	// Test with options
	var buf bytes.Buffer
	e2 := NewExecutor(
		WithLogDir("/tmp/logs"),
		WithLogWriter(&buf),
	)
	assert.Equal(t, "/tmp/logs", e2.logDir)
	assert.Equal(t, &buf, e2.logWriter)
}

func TestExecutor_Execute_Simple(t *testing.T) {
	var buf bytes.Buffer
	executor := NewExecutor(WithLogWriter(&buf))

	pipeline := &Pipeline{
		Name: "test",
		Steps: []Step{
			{Name: "echo", Run: "echo hello"},
		},
	}

	err := executor.Execute(pipeline)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Executing pipeline: test")
	assert.Contains(t, output, "Pipeline completed successfully")
}

func TestExecutor_Execute_WithDescription(t *testing.T) {
	var buf bytes.Buffer
	executor := NewExecutor(WithLogWriter(&buf))

	pipeline := &Pipeline{
		Name:        "test",
		Description: "Test pipeline",
		Steps: []Step{
			{Name: "echo", Run: "echo hello"},
		},
	}

	err := executor.Execute(pipeline)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Description: Test pipeline")
}

func TestExecutor_Execute_InvalidPipeline(t *testing.T) {
	executor := NewExecutor()

	// Empty pipeline
	pipeline := &Pipeline{
		Name: "test",
	}

	err := executor.Execute(pipeline)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline must have at least one step")
}

func TestExecutor_Execute_CommandFailure(t *testing.T) {
	var buf bytes.Buffer
	executor := NewExecutor(WithLogWriter(&buf))

	pipeline := &Pipeline{
		Name: "test",
		Steps: []Step{
			{Name: "fail", Run: "exit 1"},
		},
	}

	err := executor.Execute(pipeline)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")

	output := buf.String()
	assert.Contains(t, output, "Pipeline failed")
}

func TestExecutor_Execute_WithLogging(t *testing.T) {
	tmpDir := t.TempDir()
	var buf bytes.Buffer
	executor := NewExecutor(
		WithLogDir(tmpDir),
		WithLogWriter(&buf),
	)

	pipeline := &Pipeline{
		Name: "test",
		Steps: []Step{
			{Name: "echo", Run: "echo hello"},
		},
	}

	err := executor.Execute(pipeline)
	require.NoError(t, err)

	// Check that log directory was created
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Greater(t, len(entries), 0)

	// Find the created log directory
	var logDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "test_") {
			logDir = entry.Name()
			break
		}
	}
	assert.NotEmpty(t, logDir)

	// Check log files exist
	logPath := filepath.Join(tmpDir, logDir)
	assert.FileExists(t, filepath.Join(logPath, "pipeline.sh"))
	assert.FileExists(t, filepath.Join(logPath, "summary.txt"))
	assert.FileExists(t, filepath.Join(logPath, "pipeline.status"))

	// Check summary content
	summary, err := os.ReadFile(filepath.Join(logPath, "summary.txt"))
	require.NoError(t, err)
	assert.Contains(t, string(summary), "Pipeline: test")
	assert.Contains(t, string(summary), "Status: Success")
}

func TestExecutor_CaptureOutput(t *testing.T) {
	executor := NewExecutor()

	tests := []struct {
		name     string
		pipeline *Pipeline
		want     string
		wantErr  bool
	}{
		{
			name: "simple echo",
			pipeline: &Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "echo", Run: "echo hello world"},
				},
			},
			want:    "hello world",
			wantErr: false,
		},
		{
			name: "pipeline with pipes",
			pipeline: &Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "echo", Run: "echo 'HELLO WORLD'"},
					{Name: "tr", Run: "tr A-Z a-z"},
				},
			},
			want:    "hello world",
			wantErr: false,
		},
		{
			name: "command failure",
			pipeline: &Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "fail", Run: "exit 1"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executor.CaptureOutput(tt.pipeline)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestExecutor_Execute_MultiStep(t *testing.T) {
	var buf bytes.Buffer
	executor := NewExecutor(WithLogWriter(&buf))

	pipeline := &Pipeline{
		Name: "multi-step",
		Steps: []Step{
			{Name: "echo", Run: "echo hello", Output: "data"},
			{Name: "upper", Run: "tr a-z A-Z", Input: "data", Output: "upper"},
			{Name: "count", Run: "wc -c", Input: "upper"},
		},
	}

	err := executor.Execute(pipeline)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Command: echo hello | tr a-z A-Z | wc -c")
	assert.Contains(t, output, "Pipeline completed successfully")
}
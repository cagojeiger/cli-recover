package pipeline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *Pipeline
		want     string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "single step",
			pipeline: &Pipeline{
				Steps: []Step{
					{Name: "echo", Run: "echo hello"},
				},
			},
			want:    "echo hello",
			wantErr: false,
		},
		{
			name: "two steps",
			pipeline: &Pipeline{
				Steps: []Step{
					{Name: "echo", Run: "echo hello", Output: "data"},
					{Name: "cat", Run: "cat", Input: "data"},
				},
			},
			want:    "echo hello | cat",
			wantErr: false,
		},
		{
			name: "multiline command",
			pipeline: &Pipeline{
				Steps: []Step{
					{Name: "multi", Run: "echo line1\necho line2"},
				},
			},
			want:    "(echo line1\necho line2)",
			wantErr: false,
		},
		{
			name: "empty pipeline",
			pipeline: &Pipeline{
				Steps: []Step{},
			},
			wantErr: true,
			errMsg:  "empty pipeline",
		},
		{
			name: "non-linear pipeline",
			pipeline: &Pipeline{
				Steps: []Step{
					{Name: "source", Run: "echo data", Output: "data"},
					{Name: "branch1", Run: "cat", Input: "data"},
					{Name: "branch2", Run: "wc", Input: "data"},
				},
			},
			wantErr: true,
			errMsg:  "non-linear pipeline cannot be converted to shell command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildCommand(tt.pipeline)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuildCommandWithLogging(t *testing.T) {
	pipeline := &Pipeline{
		Name:        "test-pipeline",
		Description: "Test description",
		Steps: []Step{
			{Name: "echo", Run: "echo hello", Output: "data"},
			{Name: "cat", Run: "cat", Input: "data"},
		},
	}

	script, err := BuildCommandWithLogging(pipeline, "/tmp/logs")
	require.NoError(t, err)

	// Check script contains expected elements
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "set -o pipefail")
	assert.Contains(t, script, "LOGDIR=\"/tmp/logs\"")
	assert.Contains(t, script, "mkdir -p \"$LOGDIR\"")
	assert.Contains(t, script, "Pipeline: test-pipeline")
	assert.Contains(t, script, "Test description")
	assert.Contains(t, script, "echo.err")
	assert.Contains(t, script, "cat.out")
	assert.Contains(t, script, "pipeline.status")
}

func TestWrapCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{
			name: "simple command",
			cmd:  "echo hello",
			want: "echo hello",
		},
		{
			name: "multiline command",
			cmd:  "echo line1\necho line2",
			want: "(echo line1\necho line2)",
		},
		{
			name: "already wrapped",
			cmd:  "(echo hello)",
			want: "(echo hello)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapCommand(tt.cmd)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildCommandWithLogging_SingleStep(t *testing.T) {
	pipeline := &Pipeline{
		Name: "single-step",
		Steps: []Step{
			{Name: "echo", Run: "echo hello"},
		},
	}

	script, err := BuildCommandWithLogging(pipeline, "/tmp/logs")
	require.NoError(t, err)

	// For single step, should use tee for both stdout and stderr
	assert.Contains(t, script, "tee \"$LOGDIR/echo.out\"")
	assert.Contains(t, script, "2>\"$LOGDIR/echo.err\"")
	
	// Should not contain pipe continuation
	assert.NotContains(t, script, " | \\")
}

func TestBuildCommandWithLogging_EmptyPipeline(t *testing.T) {
	pipeline := &Pipeline{
		Name:  "empty",
		Steps: []Step{},
	}

	_, err := BuildCommandWithLogging(pipeline, "/tmp/logs")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty pipeline")
}

func TestBuildCommandWithLogging_MultiStep(t *testing.T) {
	pipeline := &Pipeline{
		Name: "multi-step",
		Steps: []Step{
			{Name: "step1", Run: "echo hello"},
			{Name: "step2", Run: "tr a-z A-Z"},
			{Name: "step3", Run: "cat"},
		},
	}

	script, err := BuildCommandWithLogging(pipeline, "/tmp/logs")
	require.NoError(t, err)

	// Check the script structure
	lines := strings.Split(script, "\n")
	
	// Find the pipeline command
	var pipelineCmd string
	for i, line := range lines {
		if strings.Contains(line, "step1") {
			// Collect the full command (may span multiple lines)
			for j := i; j < len(lines); j++ {
				pipelineCmd += lines[j] + "\n"
				if !strings.HasSuffix(strings.TrimSpace(lines[j]), "\\") {
					break
				}
			}
			break
		}
	}

	// First step should only log stderr
	assert.Contains(t, pipelineCmd, "step1.err")
	assert.NotContains(t, pipelineCmd, "step1.out")
	
	// Middle step should only log stderr
	assert.Contains(t, pipelineCmd, "step2.err")
	assert.NotContains(t, pipelineCmd, "step2.out")
	
	// Last step should log both
	assert.Contains(t, pipelineCmd, "step3.err")
	assert.Contains(t, pipelineCmd, "step3.out")
}
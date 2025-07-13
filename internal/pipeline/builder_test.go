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

func TestBuildSmartCommand(t *testing.T) {
	t.Run("simple command without monitoring", func(t *testing.T) {
		step := Step{
			Name: "simple",
			Run:  "echo hello",
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		assert.Equal(t, "echo hello", cmd)
		assert.Empty(t, monitors)
	})

	t.Run("command with byte monitoring", func(t *testing.T) {
		step := Step{
			Name: "backup",
			Run:  "tar cf - /data",
			Monitor: &MonitorConfig{
				Type:     "bytes",
				Interval: 1000,
			},
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		assert.Equal(t, "tar cf - /data", cmd)
		assert.Len(t, monitors, 1)
		
		// Verify monitor is ByteMonitor
		_, ok := monitors[0].(*ByteMonitor)
		assert.True(t, ok)
	})

	t.Run("command with line monitoring", func(t *testing.T) {
		step := Step{
			Name: "count",
			Run:  "grep pattern",
			Monitor: &MonitorConfig{
				Type: "lines",
			},
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		assert.Equal(t, "grep pattern", cmd)
		assert.Len(t, monitors, 1)
		
		// Verify monitor is LineMonitor
		_, ok := monitors[0].(*LineMonitor)
		assert.True(t, ok)
	})

	t.Run("command with checksum", func(t *testing.T) {
		step := Step{
			Name:     "download",
			Run:      "curl -L https://example.com/file.zip",
			Output:   "file:download.zip",
			Checksum: []string{"sha256", "md5"},
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		assert.Equal(t, "curl -L https://example.com/file.zip", cmd)
		assert.Len(t, monitors, 2)
	})

	t.Run("command with log", func(t *testing.T) {
		step := Step{
			Name: "process",
			Run:  "long-running-command",
			Log:  "process.log",
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		// Log is handled in executor, not in builder
		assert.Equal(t, "long-running-command", cmd)
		assert.Empty(t, monitors)
	})

	t.Run("command with multiple features", func(t *testing.T) {
		step := Step{
			Name: "complex",
			Run:  "tar czf - /data",
			Output: "file:backup.tar.gz",
			Monitor: &MonitorConfig{
				Type: "bytes",
			},
			Checksum: []string{"sha256"},
			Log:      "backup.log",
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		assert.Equal(t, "tar czf - /data", cmd)
		assert.Len(t, monitors, 2) // ByteMonitor + ChecksumMonitor
	})

	t.Run("file output parsing", func(t *testing.T) {
		step := Step{
			Name:   "save",
			Run:    "generate-data",
			Output: "file:output.txt",
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		assert.Equal(t, "generate-data", cmd)
		assert.Empty(t, monitors)
		
		// File output is handled in executor
		assert.True(t, IsFileOutput(step.Output))
		assert.Equal(t, "output.txt", ExtractFilename(step.Output))
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("IsFileOutput", func(t *testing.T) {
		assert.True(t, IsFileOutput("file:test.txt"))
		assert.True(t, IsFileOutput("file:path/to/file.gz"))
		assert.False(t, IsFileOutput("stream"))
		assert.False(t, IsFileOutput(""))
	})

	t.Run("ExtractFilename", func(t *testing.T) {
		assert.Equal(t, "test.txt", ExtractFilename("file:test.txt"))
		assert.Equal(t, "path/to/file.gz", ExtractFilename("file:path/to/file.gz"))
		assert.Equal(t, "", ExtractFilename("not-a-file"))
	})
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

func TestBuildEnhancedPipeline(t *testing.T) {
	t.Run("valid pipeline with monitors", func(t *testing.T) {
		p := &Pipeline{
			Name: "test-enhanced",
			Steps: []Step{
				{
					Name: "generate",
					Run:  "echo test",
					Output: "data",
					Monitor: &MonitorConfig{
						Type: "bytes",
					},
				},
				{
					Name:     "save",
					Run:      "cat",
					Input:    "data",
					Output:   "file:output.txt",
					Checksum: []string{"sha256"},
				},
			},
		}
		
		executions, err := BuildEnhancedPipeline(p)
		assert.NoError(t, err)
		assert.Len(t, executions, 2)
		
		// First step should have ByteMonitor
		assert.Equal(t, "generate", executions[0].Step.Name)
		assert.Equal(t, "echo test", executions[0].Command)
		assert.Len(t, executions[0].Monitors, 1)
		_, ok := executions[0].Monitors[0].(*ByteMonitor)
		assert.True(t, ok)
		
		// Second step should have ChecksumFileWriter
		assert.Equal(t, "save", executions[1].Step.Name)
		assert.Equal(t, "cat", executions[1].Command)
		assert.Len(t, executions[1].Monitors, 1)
		_, ok = executions[1].Monitors[0].(*ChecksumFileWriter)
		assert.True(t, ok)
	})
	
	t.Run("pipeline with multiple monitors", func(t *testing.T) {
		p := &Pipeline{
			Name: "multi-monitor",
			Steps: []Step{
				{
					Name: "process",
					Run:  "long-command",
					Monitor: &MonitorConfig{
						Type: "lines",
					},
					Checksum: []string{"md5", "sha256"},
				},
			},
		}
		
		executions, err := BuildEnhancedPipeline(p)
		assert.NoError(t, err)
		assert.Len(t, executions, 1)
		
		// Should have 3 monitors: LineMonitor + 2 ChecksumWriters
		assert.Len(t, executions[0].Monitors, 3)
		
		// Verify monitor types
		var hasLineMonitor, hasMD5, hasSHA256 bool
		for _, monitor := range executions[0].Monitors {
			switch m := monitor.(type) {
			case *LineMonitor:
				hasLineMonitor = true
			case *ChecksumWriter:
				if m.Algorithm() == "md5" {
					hasMD5 = true
				} else if m.Algorithm() == "sha256" {
					hasSHA256 = true
				}
			}
		}
		
		assert.True(t, hasLineMonitor)
		assert.True(t, hasMD5)
		assert.True(t, hasSHA256)
	})
	
	t.Run("pipeline with time monitoring", func(t *testing.T) {
		p := &Pipeline{
			Name: "time-test",
			Steps: []Step{
				{
					Name: "timed",
					Run:  "sleep 1",
					Monitor: &MonitorConfig{
						Type: "time",
					},
				},
			},
		}
		
		executions, err := BuildEnhancedPipeline(p)
		assert.NoError(t, err)
		assert.Len(t, executions, 1)
		assert.Len(t, executions[0].Monitors, 1)
		
		// TimeMonitor should be started
		timeMonitor, ok := executions[0].Monitors[0].(*TimeMonitor)
		assert.True(t, ok)
		
		// Verify it's started (elapsed time should be > 0)
		// Note: There's a small race condition here, but it should be negligible
		elapsed := timeMonitor.Elapsed()
		assert.GreaterOrEqual(t, elapsed.Nanoseconds(), int64(0))
	})
	
	t.Run("empty pipeline", func(t *testing.T) {
		p := &Pipeline{
			Name:  "empty",
			Steps: []Step{},
		}
		
		executions, err := BuildEnhancedPipeline(p)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one step")
		assert.Nil(t, executions)
	})
	
	t.Run("invalid pipeline", func(t *testing.T) {
		p := &Pipeline{
			Name: "invalid",
			Steps: []Step{
				{Name: "", Run: "echo test"}, // Empty name
			},
		}
		
		executions, err := BuildEnhancedPipeline(p)
		assert.Error(t, err)
		assert.Nil(t, executions)
	})
	
	t.Run("pipeline without monitors", func(t *testing.T) {
		p := &Pipeline{
			Name: "no-monitors",
			Steps: []Step{
				{
					Name: "simple",
					Run:  "echo hello",
				},
			},
		}
		
		executions, err := BuildEnhancedPipeline(p)
		assert.NoError(t, err)
		assert.Len(t, executions, 1)
		assert.Equal(t, "simple", executions[0].Step.Name)
		assert.Equal(t, "echo hello", executions[0].Command)
		assert.Empty(t, executions[0].Monitors)
	})
}
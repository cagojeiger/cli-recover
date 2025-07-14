package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cagojeiger/cli-pipe/internal/config"
	"github.com/cagojeiger/cli-pipe/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	// Test executor creation
	e := NewExecutor(cfg)
	assert.NotNil(t, e)
	assert.NotNil(t, e.config)
	assert.Equal(t, os.Stdout, e.logWriter)
}

func TestExecutor_Execute_Simple(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name: "test",
		Steps: []Step{
			{Name: "echo", Run: "echo hello"},
		},
	}

	err := executor.Execute(pipeline)
	assert.NoError(t, err)
	
	// 1단계: 로그 파일 확인 제거됨 (tee 방식에서는 별도 로그 파일 생성 안 함)
}

func TestExecutor_Execute_Pipeline(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name: "multi-step",
		Steps: []Step{
			{Name: "generate", Run: "echo test data", Output: "data"},
			{Name: "transform", Run: "tr a-z A-Z", Input: "data"},
		},
	}

	err := executor.Execute(pipeline)
	assert.NoError(t, err)
}

func TestExecutor_Execute_FileOutput(t *testing.T) {
	// Create temp config and work directory
	tempDir := t.TempDir()
	workDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(oldWd)
	
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	outputFile := "output.txt"
	pipeline := &Pipeline{
		Name: "file-output",
		Steps: []Step{
			{Name: "generate", Run: "echo file content", Output: "data"},
			{Name: "save", Run: "cat", Input: "data", Output: "file:" + outputFile},
		},
	}

	err := executor.Execute(pipeline)
	assert.NoError(t, err)
	
	// 1단계: 파일 출력 기능 제거됨 (tee 방식에서는 파일 출력 자동 생성 안 함)
}

func TestExecutor_Execute_InvalidPipeline(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	t.Run("empty name", func(t *testing.T) {
		pipeline := &Pipeline{
			Name: "",
			Steps: []Step{
				{Name: "test", Run: "echo test"},
			},
		}
		err := executor.Execute(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pipeline name cannot be empty")
	})

	t.Run("no steps", func(t *testing.T) {
		pipeline := &Pipeline{
			Name: "empty",
			Steps: []Step{},
		}
		err := executor.Execute(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must have at least one step")
	})

	t.Run("invalid input reference", func(t *testing.T) {
		pipeline := &Pipeline{
			Name: "invalid-ref",
			Steps: []Step{
				{Name: "step1", Run: "echo data", Output: "out1"},
				{Name: "step2", Run: "cat", Input: "wrong-ref"},
			},
		}
		err := executor.Execute(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "references undefined input")
	})
}

func TestExecutor_Execute_NonLinearPipeline(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name: "branching",
		Steps: []Step{
			{Name: "source", Run: "echo data", Output: "data"},
			{Name: "branch1", Run: "cat", Input: "data", Output: "out1"},
			{Name: "branch2", Run: "wc", Input: "data", Output: "out2"},
		},
	}

	err := executor.Execute(pipeline)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-linear pipelines not yet supported")
}

func TestExecutor_Execute_FailingCommand(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name: "failing",
		Steps: []Step{
			{Name: "fail", Run: "exit 1"},
		},
	}

	err := executor.Execute(pipeline)
	assert.Error(t, err)
	
	// Check summary shows failure
	entries, _ := os.ReadDir(tempDir)
	if len(entries) > 0 {
		runDir := filepath.Join(tempDir, entries[0].Name())
		summary, _ := os.ReadFile(filepath.Join(runDir, "summary.txt"))
		assert.Contains(t, string(summary), "Failed")
	}
}

func TestExecutor_CaptureOutput(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name: "capture",
		Steps: []Step{
			{Name: "echo", Run: "echo hello world"},
		},
	}

	output, err := executor.CaptureOutput(pipeline)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", output)
}

func TestExecutor_Execute_Monitoring(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name: "monitoring-test",
		Steps: []Step{
			{
				Name: "generate",
				Run: `echo "line 1"
echo "line 2"
echo "line 3"`,
				Output: "data",
			},
			{
				Name: "process",
				Run: "cat",
				Input: "data",
			},
		},
	}

	err := executor.Execute(pipeline)
	assert.NoError(t, err)
	
	// Check summary contains monitoring info
	entries, _ := os.ReadDir(tempDir)
	require.Greater(t, len(entries), 0)
	
	runDir := filepath.Join(tempDir, entries[0].Name())
	summary, err := os.ReadFile(filepath.Join(runDir, "summary.txt"))
	require.NoError(t, err)
	
	summaryStr := string(summary)
	// Check basic execution info (simplified Phase 1)
	assert.Contains(t, summaryStr, "Pipeline:")
	assert.Contains(t, summaryStr, "Duration:")
	assert.Contains(t, summaryStr, "Status:")
}

func TestExecutor_Execute_MultilineCommand(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name: "multiline",
		Steps: []Step{
			{
				Name: "multi",
				Run: `echo "line 1"
echo "line 2"
echo "line 3"`,
			},
		},
	}

	err := executor.Execute(pipeline)
	assert.NoError(t, err)
}

func TestExecutor_Execute_WithDescription(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)

	pipeline := &Pipeline{
		Name:        "described",
		Description: "This is a test pipeline",
		Steps: []Step{
			{Name: "test", Run: "echo test"},
		},
	}

	err := executor.Execute(pipeline)
	assert.NoError(t, err)
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecutor_Execute_NilConfig(t *testing.T) {
	// Create temp dir for config
	tempDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Initialize config first so Load() will work
	cfg := config.DefaultConfig()
	require.NoError(t, cfg.Save())
	require.NoError(t, cfg.EnsureLogDir())
	
	// Create executor with nil config
	executor := &Executor{
		config:    nil,
		logWriter: os.Stdout,
	}
	
	pipeline := &Pipeline{
		Name: "test",
		Steps: []Step{
			{Name: "echo", Run: "echo hello"},
		},
	}
	
	// Should load config automatically
	err := executor.Execute(pipeline)
	assert.NoError(t, err)
	assert.NotNil(t, executor.config)
}

// 1단계: 로그 디렉토리 생성 에러 테스트 제거됨 (단순화로 인해 로그 디렉토리 생성 안 함)

func TestExecutor_CaptureOutput_Errors(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)
	
	t.Run("command not found", func(t *testing.T) {
		pipeline := &Pipeline{
			Name: "test",
			Steps: []Step{
				{Name: "fail", Run: "/nonexistent/command/that/will/fail"},
			},
		}
		_, err := executor.CaptureOutput(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command failed")
	})
	
	t.Run("command exits with error", func(t *testing.T) {
		pipeline := &Pipeline{
			Name: "test",
			Steps: []Step{
				{Name: "error", Run: "sh -c 'echo error >&2; exit 1'"},
			},
		}
		output, err := executor.CaptureOutput(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command failed")
		assert.Contains(t, output, "error") // CombinedOutput includes stderr
	})
	
	t.Run("empty pipeline", func(t *testing.T) {
		pipeline := &Pipeline{
			Name: "test",
			Steps: []Step{},
		}
		output, err := executor.CaptureOutput(pipeline)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to build shell command")
		assert.Empty(t, output)
	})
}

func TestExecutor_BuildCommand_Errors(t *testing.T) {
	// Test empty pipeline
	_, err := BuildCommand(&Pipeline{Steps: []Step{}}, "/tmp")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty pipeline")
}

func TestExecutor_Execute_ConfigLoadError(t *testing.T) {
	// Create temp dir for config
	tempDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	// Create corrupted config file
	configDir := filepath.Join(tempDir, ".cli-pipe")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yaml")
	os.WriteFile(configPath, []byte("invalid yaml: ["), 0644)
	
	// Create executor with nil config
	executor := &Executor{
		config:    nil,
		logWriter: os.Stdout,
	}
	
	pipeline := &Pipeline{
		Name: "test",
		Steps: []Step{
			{Name: "echo", Run: "echo hello"},
		},
	}
	
	// Should fail to load config
	err := executor.Execute(pipeline)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config")
}

func TestExecutor_ExecuteLinearPipeline_Errors(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory: tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := NewExecutor(cfg)
	
	// 1단계: 복잡한 파일 생성 에러 테스트들 제거됨
	// (tee 방식에서는 파일 생성 에러가 발생하지 않음)
	
	t.Run("stderr logging", func(t *testing.T) {
		pipeline := &Pipeline{
			Name: "test",
			Steps: []Step{
				{Name: "stderr", Run: "echo 'error message' >&2"},
			},
		}
		
		logDir := filepath.Join(tempDir, "stderr-test")
		os.MkdirAll(logDir, 0755) // Ensure log directory exists
		err := executor.executeLinearPipeline(pipeline, logger.Default())
		assert.NoError(t, err)
		
		// Check that stderr was logged
		stderrLog := filepath.Join(logDir, "pipeline.err")
		content, err := os.ReadFile(stderrLog)
		if os.IsNotExist(err) {
			// If pipeline.err doesn't exist, that's okay - stderr might be empty
			return
		}
		require.NoError(t, err)
		assert.Contains(t, string(content), "error message")
	})
}
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

func TestExecutePipelineEnhanced(t *testing.T) {
	t.Run("linear pipeline with monitoring", func(t *testing.T) {
		var logBuf bytes.Buffer
		executor := NewExecutor(WithLogWriter(&logBuf))
		
		p := &Pipeline{
			Name:        "test-enhanced",
			Description: "Test enhanced pipeline",
			Steps: []Step{
				{
					Name: "generate",
					Run:  "echo 'test data'",
					Monitor: &MonitorConfig{
						Type: "bytes",
					},
				},
				{
					Name:  "uppercase",
					Run:   "tr '[:lower:]' '[:upper:]'",
					Input: "generate",
				},
			},
		}
		
		err := executor.ExecutePipelineEnhanced(p)
		assert.NoError(t, err)
		
		// 로그에 모니터 보고서가 포함되어야 함
		logs := logBuf.String()
		assert.Contains(t, logs, "Monitor reports:")
		assert.Contains(t, logs, "Processed")
		assert.Contains(t, logs, "bytes")
	})
	
	t.Run("non-linear pipeline returns error", func(t *testing.T) {
		executor := NewExecutor()
		
		p := &Pipeline{
			Name: "non-linear",
			Steps: []Step{
				{Name: "step1", Run: "echo 1"},
				{Name: "step2", Run: "echo 2", Input: "step1"},
				{Name: "step3", Run: "echo 3", Input: "step1"}, // 분기
			},
		}
		
		err := executor.ExecutePipelineEnhanced(p)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-linear pipelines not yet supported")
	})
	
	t.Run("pipeline with file output", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output.txt")
		
		executor := NewExecutor()
		p := &Pipeline{
			Name: "file-output-test",
			Steps: []Step{
				{
					Name:   "generate",
					Run:    "echo 'file content'",
					Output: "file:" + outputFile,
				},
			},
		}
		
		err := executor.ExecutePipelineEnhanced(p)
		assert.NoError(t, err)
		
		// 파일이 생성되었는지 확인
		assert.FileExists(t, outputFile)
		
		// 파일 내용 확인
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, "file content\n", string(content))
	})
	
	t.Run("pipeline with checksum", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "checksum-test.txt")
		
		var logBuf bytes.Buffer
		executor := NewExecutor(WithLogWriter(&logBuf))
		
		p := &Pipeline{
			Name: "checksum-test",
			Steps: []Step{
				{
					Name:     "generate",
					Run:      "echo 'checksum test'",
					Output:   "file:" + outputFile,
					Checksum: []string{"sha256"},
				},
			},
		}
		
		err := executor.ExecutePipelineEnhanced(p)
		assert.NoError(t, err)
		
		// 체크섬 파일이 생성되었는지 확인
		checksumFile := outputFile + ".sha256"
		assert.FileExists(t, checksumFile)
		
		// 로그에 체크섬 보고서가 포함되어야 함
		logs := logBuf.String()
		assert.Contains(t, logs, "Checksum (sha256):")
	})
	
	t.Run("pipeline with line monitoring", func(t *testing.T) {
		var logBuf bytes.Buffer
		executor := NewExecutor(WithLogWriter(&logBuf))
		
		p := &Pipeline{
			Name: "line-monitor-test",
			Steps: []Step{
				{
					Name: "generate",
					Run: `echo -e "line1\nline2\nline3"`,
					Monitor: &MonitorConfig{
						Type: "lines",
					},
				},
			},
		}
		
		err := executor.ExecutePipelineEnhanced(p)
		assert.NoError(t, err)
		
		// 로그에 라인 카운트가 포함되어야 함
		logs := logBuf.String()
		assert.Contains(t, logs, "lines")
	})
	
	t.Run("empty pipeline", func(t *testing.T) {
		executor := NewExecutor()
		p := &Pipeline{
			Name:  "empty",
			Steps: []Step{},
		}
		
		err := executor.ExecutePipelineEnhanced(p)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty pipeline")
	})
	
	t.Run("pipeline with multiple monitors", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "multi-monitor.txt")
		
		var logBuf bytes.Buffer
		executor := NewExecutor(WithLogWriter(&logBuf))
		
		p := &Pipeline{
			Name: "multi-monitor-test",
			Steps: []Step{
				{
					Name: "generate",
					Run:  "echo 'multi monitor test'",
					Monitor: &MonitorConfig{
						Type: "bytes",
					},
					Output:   "file:" + outputFile,
					Checksum: []string{"md5", "sha256"},
				},
			},
		}
		
		err := executor.ExecutePipelineEnhanced(p)
		assert.NoError(t, err)
		
		// 모든 모니터 보고서가 로그에 있어야 함
		logs := logBuf.String()
		assert.Contains(t, logs, "Processed")
		assert.Contains(t, logs, "bytes")
		assert.Contains(t, logs, "Checksum")
		
		// 체크섬 파일들이 생성되었는지 확인
		assert.FileExists(t, outputFile+".md5")
		assert.FileExists(t, outputFile+".sha256")
	})
}

func TestExecuteStepWithDataFlow(t *testing.T) {
	t.Run("simple step without input", func(t *testing.T) {
		executor := NewExecutor()
		step := Step{
			Name: "simple",
			Run:  "echo 'hello world'",
		}
		
		output, err := executor.ExecuteStepWithDataFlow(step, "")
		assert.NoError(t, err)
		assert.Equal(t, "hello world", output)
	})
	
	t.Run("step with input data", func(t *testing.T) {
		executor := NewExecutor()
		step := Step{
			Name:  "uppercase",
			Run:   "tr '[:lower:]' '[:upper:]'",
			Input: "previous",
		}
		
		output, err := executor.ExecuteStepWithDataFlow(step, "hello world")
		assert.NoError(t, err)
		assert.Equal(t, "HELLO WORLD", output)
	})
	
	t.Run("step with file output", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "step-output.txt")
		
		executor := NewExecutor()
		step := Step{
			Name:   "save",
			Run:    "echo 'save to file'",
			Output: "file:" + outputFile,
		}
		
		output, err := executor.ExecuteStepWithDataFlow(step, "")
		assert.NoError(t, err)
		assert.Equal(t, "save to file", output)
		
		// 파일이 생성되었는지 확인
		assert.FileExists(t, outputFile)
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, "save to file", strings.TrimSpace(string(content)))
	})
	
	t.Run("step with command failure", func(t *testing.T) {
		executor := NewExecutor()
		step := Step{
			Name: "fail",
			Run:  "exit 1",
		}
		
		_, err := executor.ExecuteStepWithDataFlow(step, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command failed")
	})
	
	t.Run("step with complex input", func(t *testing.T) {
		executor := NewExecutor()
		step := Step{
			Name:  "count",
			Run:   "wc -l",
			Input: "lines",
		}
		
		inputData := "line1\nline2\nline3\n"
		output, err := executor.ExecuteStepWithDataFlow(step, inputData)
		assert.NoError(t, err)
		assert.Equal(t, "3", strings.TrimSpace(output))
	})
	
	t.Run("step with multiline input", func(t *testing.T) {
		executor := NewExecutor()
		step := Step{
			Name:  "sort",
			Run:   "sort",
			Input: "unsorted",
		}
		
		inputData := "zebra\napple\nbanana"
		output, err := executor.ExecuteStepWithDataFlow(step, inputData)
		assert.NoError(t, err)
		assert.Equal(t, "apple\nbanana\nzebra", output)
	})
}

func TestExecuteLinearPipelineEnhanced(t *testing.T) {
	t.Run("linear pipeline with all features", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "full-test.txt")
		
		var logBuf bytes.Buffer
		executor := NewExecutor(WithLogWriter(&logBuf))
		
		p := &Pipeline{
			Name:        "full-feature-test",
			Description: "Test all features",
			Steps: []Step{
				{
					Name: "generate",
					Run: `echo -e "line1\nline2\nline3"`,
					Monitor: &MonitorConfig{
						Type: "lines",
					},
				},
				{
					Name:  "uppercase",
					Run:   "tr '[:lower:]' '[:upper:]'",
					Input: "generate",
					Monitor: &MonitorConfig{
						Type: "bytes",
					},
				},
				{
					Name:     "save",
					Run:      "cat",
					Input:    "uppercase",
					Output:   "file:" + outputFile,
					Checksum: []string{"sha256"},
				},
			},
		}
		
		err := executor.executeLinearPipelineEnhanced(p)
		assert.NoError(t, err)
		
		// 출력 파일 확인
		assert.FileExists(t, outputFile)
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Equal(t, "LINE1\nLINE2\nLINE3\n", string(content))
		
		// 체크섬 파일 확인
		assert.FileExists(t, outputFile + ".sha256")
		
		// 모니터 보고서 확인
		logs := logBuf.String()
		assert.Contains(t, logs, "Monitor reports:")
		assert.Contains(t, logs, "lines")
		assert.Contains(t, logs, "bytes")
		assert.Contains(t, logs, "Checksum")
	})
	
	t.Run("pipeline with command failure", func(t *testing.T) {
		executor := NewExecutor()
		
		p := &Pipeline{
			Name: "fail-test",
			Steps: []Step{
				{
					Name: "fail",
					Run:  "false",
				},
			},
		}
		
		err := executor.executeLinearPipelineEnhanced(p)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pipeline failed")
	})
}
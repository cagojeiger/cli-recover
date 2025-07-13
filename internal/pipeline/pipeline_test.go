package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestPipeline_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pipeline Pipeline
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pipeline",
			pipeline: Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "step1", Run: "echo hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "step1", Run: "echo hello"},
				},
			},
			wantErr: true,
			errMsg:  "pipeline name cannot be empty",
		},
		{
			name: "no steps",
			pipeline: Pipeline{
				Name: "test",
			},
			wantErr: true,
			errMsg:  "pipeline must have at least one step",
		},
		{
			name: "duplicate step names",
			pipeline: Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "step1", Run: "echo hello"},
					{Name: "step1", Run: "echo world"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate step name: step1",
		},
		{
			name: "empty step name",
			pipeline: Pipeline{
				Name: "test",
				Steps: []Step{
					{Run: "echo hello"},
				},
			},
			wantErr: true,
			errMsg:  "step name cannot be empty",
		},
		{
			name: "empty step command",
			pipeline: Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "step1"},
				},
			},
			wantErr: true,
			errMsg:  "step 'step1' has empty command",
		},
		{
			name: "undefined input",
			pipeline: Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data"},
					{Name: "step2", Run: "cat", Input: "unknown"},
				},
			},
			wantErr: true,
			errMsg:  "step 'step2' references undefined input 'unknown'",
		},
		{
			name: "valid with input/output",
			pipeline: Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data"},
					{Name: "step2", Run: "cat", Input: "data"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pipeline.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPipeline_IsLinear(t *testing.T) {
	tests := []struct {
		name     string
		pipeline Pipeline
		want     bool
	}{
		{
			name: "single step",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "step1", Run: "echo hello"},
				},
			},
			want: true,
		},
		{
			name: "linear chain",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data1"},
					{Name: "step2", Run: "cat", Input: "data1", Output: "data2"},
					{Name: "step3", Run: "wc", Input: "data2"},
				},
			},
			want: true,
		},
		{
			name: "branching pipeline",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data"},
					{Name: "step2", Run: "cat", Input: "data"},
					{Name: "step3", Run: "wc", Input: "data"},
				},
			},
			want: false,
		},
		{
			name: "non-sequential",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data1"},
					{Name: "step2", Run: "echo world", Output: "data2"},
					{Name: "step3", Run: "cat", Input: "data1"},
				},
			},
			want: false, // Not linear because step2's output is not used
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pipeline.IsLinear()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStep_EnhancedFields(t *testing.T) {
	t.Run("parses monitor config", func(t *testing.T) {
		yaml := `
name: test-step
run: tar cf - /data
output: archive
monitor:
  type: bytes
  interval: 1000
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.NotNil(t, step.Monitor)
		assert.Equal(t, "bytes", step.Monitor.Type)
		assert.Equal(t, 1000, step.Monitor.Interval)
	})

	t.Run("parses checksum array", func(t *testing.T) {
		yaml := `
name: test-step
run: tar cf - /data
output: file:backup.tar
checksum: [sha256, md5]
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.Len(t, step.Checksum, 2)
		assert.Contains(t, step.Checksum, "sha256")
		assert.Contains(t, step.Checksum, "md5")
	})

	t.Run("parses log field", func(t *testing.T) {
		yaml := `
name: test-step
run: long-running-command
log: output.log
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.Equal(t, "output.log", step.Log)
	})

	t.Run("parses progress field", func(t *testing.T) {
		yaml := `
name: test-step
run: gzip -9
progress: true
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.True(t, step.Progress)
	})

	t.Run("handles empty optional fields", func(t *testing.T) {
		yaml := `
name: test-step
run: echo hello
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.Nil(t, step.Monitor)
		assert.Nil(t, step.Checksum)
		assert.Empty(t, step.Log)
		assert.False(t, step.Progress)
	})

	t.Run("backward compatibility", func(t *testing.T) {
		// 기존 Step이 여전히 작동하는지 확인
		step := Step{
			Name:   "old-step",
			Run:    "echo test",
			Input:  "input",
			Output: "output",
		}
		
		// 새 필드들은 기본값을 가져야 함
		assert.Nil(t, step.Monitor)
		assert.Nil(t, step.Checksum)
		assert.Empty(t, step.Log)
		assert.False(t, step.Progress)
	})
}

// parseYAMLString is a helper function for testing YAML parsing
func parseYAMLString(yamlStr string, v interface{}) error {
	return yaml.Unmarshal([]byte(yamlStr), v)
}

func TestPipeline_IsLinear_EdgeCases(t *testing.T) {
	t.Run("empty pipeline", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{},
		}
		// Empty pipeline is considered linear
		assert.True(t, p.IsLinear())
	})
	
	t.Run("multiple outputs same name", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{
				{Name: "step1", Run: "echo 1", Output: "data"},
				{Name: "step2", Run: "echo 2", Output: "data"}, // Reuses same output name
				{Name: "step3", Run: "cat", Input: "data"},
			},
		}
		// This is still linear because step3 uses the last "data"
		assert.True(t, p.IsLinear())
	})
	
	t.Run("circular reference", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{
				{Name: "step1", Run: "echo 1", Input: "data3", Output: "data1"},
				{Name: "step2", Run: "cat", Input: "data1", Output: "data2"},
				{Name: "step3", Run: "wc", Input: "data2", Output: "data3"},
			},
		}
		// Circular reference but still linear flow
		assert.True(t, p.IsLinear())
	})
	
	t.Run("unused intermediate output", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{
				{Name: "step1", Run: "echo 1", Output: "data1"},
				{Name: "step2", Run: "echo 2", Output: "data2"},
				{Name: "step3", Run: "echo 3", Output: "data3"},
				{Name: "step4", Run: "cat", Input: "data3"},
			},
		}
		// Actually linear - IsLinear only checks if steps form a chain
		// It doesn't verify all outputs are used
		assert.True(t, p.IsLinear())
	})
	
	t.Run("complex branching", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{
				{Name: "source", Run: "echo data", Output: "raw"},
				{Name: "process1", Run: "tr a-z A-Z", Input: "raw", Output: "upper"},
				{Name: "process2", Run: "tr A-Z a-z", Input: "raw", Output: "lower"},
				{Name: "combine", Run: "cat", Input: "upper"}, // Only uses one branch
			},
		}
		// Not linear due to branching
		assert.False(t, p.IsLinear())
	})
	
	t.Run("linear with skipped step", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{
				{Name: "step1", Run: "echo 1", Output: "data1"},
				{Name: "step2", Run: "cat", Input: "data1", Output: "data2"},
				{Name: "independent", Run: "date"}, // No input/output
				{Name: "step3", Run: "wc", Input: "data2"},
			},
		}
		// Still linear despite independent step
		assert.True(t, p.IsLinear())
	})
}
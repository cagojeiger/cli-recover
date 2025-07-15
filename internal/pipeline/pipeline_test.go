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

func TestStep_BasicFields(t *testing.T) {
	t.Run("parses basic step", func(t *testing.T) {
		yaml := `
name: test-step
run: echo hello
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.Equal(t, "test-step", step.Name)
		assert.Equal(t, "echo hello", step.Run)
	})

	t.Run("parses step with input/output", func(t *testing.T) {
		yaml := `
name: test-step
run: cat
input: data
output: processed
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.Equal(t, "data", step.Input)
		assert.Equal(t, "processed", step.Output)
	})

	t.Run("parses step with file output", func(t *testing.T) {
		yaml := `
name: test-step
run: tar cf - /data
output: file:backup.tar
`
		var step Step
		err := parseYAMLString(yaml, &step)
		
		assert.NoError(t, err)
		assert.Equal(t, "file:backup.tar", step.Output)
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
	
	t.Run("non-linear unused output", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{
				{Name: "step1", Run: "echo 1", Output: "data1"},
				{Name: "step2", Run: "echo 2", Output: "data2"}, // data1 is not used by step2
				{Name: "step3", Run: "cat", Input: "data3"},     // and data1 is not used later
			},
		}
		// Not linear because data1 is never used
		assert.False(t, p.IsLinear())
	})
	
	t.Run("non-linear with mismatched chain", func(t *testing.T) {
		p := Pipeline{
			Steps: []Step{
				{Name: "step1", Run: "echo 1", Output: "data1"},
				{Name: "step2", Run: "cat", Input: "data2", Output: "data3"}, // Input doesn't match previous output
				{Name: "step3", Run: "wc", Input: "data3"},
			},
		}
		// Not linear because step2's input doesn't match step1's output
		assert.False(t, p.IsLinear())
	})
}

func TestPipeline_IsTree(t *testing.T) {
	tests := []struct {
		name     string
		pipeline Pipeline
		want     bool
	}{
		{
			name: "simple tree - one branch",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "root", Run: "echo data", Output: "data"},
					{Name: "branch1", Run: "cat", Input: "data"},
					{Name: "branch2", Run: "wc", Input: "data"},
				},
			},
			want: true,
		},
		{
			name: "tree with multiple levels",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "root", Run: "curl api.com", Output: "raw"},
					{Name: "backup", Run: "gzip", Input: "raw"},
					{Name: "process", Run: "jq .users", Input: "raw", Output: "users"},
					{Name: "count", Run: "wc -l", Input: "users"},
					{Name: "filter", Run: "grep active", Input: "users"},
				},
			},
			want: true,
		},
		{
			name: "not tree - multiple inputs (merge)",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "src1", Run: "echo 1", Output: "data1"},
					{Name: "src2", Run: "echo 2", Output: "data2"},
					{Name: "merge", Run: "cat", Input: "data1,data2"}, // Multiple inputs
				},
			},
			want: false,
		},
		{
			name: "not tree - circular reference",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "step1", Run: "cat", Input: "data2", Output: "data1"},
					{Name: "step2", Run: "cat", Input: "data1", Output: "data2"},
				},
			},
			want: false,
		},
		{
			name: "linear pipeline is also a tree",
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
			name: "disconnected components",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "tree1_root", Run: "echo 1", Output: "data1"},
					{Name: "tree1_leaf", Run: "cat", Input: "data1"},
					{Name: "tree2_root", Run: "echo 2", Output: "data2"},
					{Name: "tree2_leaf", Run: "wc", Input: "data2"},
				},
			},
			want: true, // Multiple trees are still valid
		},
		{
			name: "empty pipeline",
			pipeline: Pipeline{
				Steps: []Step{},
			},
			want: true,
		},
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
			name: "complex tree structure",
			pipeline: Pipeline{
				Steps: []Step{
					{Name: "fetch", Run: "curl api.com", Output: "api_data"},
					{Name: "parse", Run: "jq .", Input: "api_data", Output: "json"},
					{Name: "users", Run: "jq .users", Input: "json", Output: "user_list"},
					{Name: "logs", Run: "jq .logs", Input: "json", Output: "log_list"},
					{Name: "active_users", Run: "grep active", Input: "user_list"},
					{Name: "error_logs", Run: "grep ERROR", Input: "log_list"},
					{Name: "user_count", Run: "wc -l", Input: "user_list"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pipeline.IsTree()
			assert.Equal(t, tt.want, got)
		})
	}
}
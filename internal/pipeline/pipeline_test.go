package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
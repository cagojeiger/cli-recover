package strategy

import (
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestIsSimpleLinear(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *entity.Pipeline
		want     bool
	}{
		{
			name: "simple two-step linear pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "echo hello", Output: "text"},
					{Name: "step2", Run: "tr upper", Input: "text"},
				},
			},
			want: true,
		},
		{
			name: "three-step linear pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "cat file", Output: "content"},
					{Name: "step2", Run: "grep pattern", Input: "content", Output: "filtered"},
					{Name: "step3", Run: "wc -l", Input: "filtered"},
				},
			},
			want: true,
		},
		{
			name: "pipeline with branch (one output used by two steps)",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "echo hello", Output: "text"},
					{Name: "step2", Run: "tr upper", Input: "text", Output: "upper"},
					{Name: "step3", Run: "tr lower", Input: "text", Output: "lower"},
				},
			},
			want: false, // Not linear because 'text' is used by two steps
		},
		{
			name: "pipeline with merge (two outputs into one step)",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "echo hello", Output: "text1"},
					{Name: "step2", Run: "echo world", Output: "text2"},
					{Name: "step3", Run: "cat", Input: "text1"}, // Would need both inputs
				},
			},
			want: false, // Not properly linear
		},
		{
			name: "single step pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "echo hello"},
				},
			},
			want: true,
		},
		{
			name: "empty pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{},
			},
			want: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSimpleLinear(tt.pipeline)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequiresProgress(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *entity.Pipeline
		want     bool
	}{
		{
			name: "pipeline with large data operations",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "backup", Run: "tar cf - /data", Output: "archive"},
					{Name: "compress", Run: "gzip -9", Input: "archive"},
				},
			},
			want: true, // tar operations typically need progress
		},
		{
			name: "simple echo pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "echo", Run: "echo hello", Output: "text"},
					{Name: "upper", Run: "tr upper", Input: "text"},
				},
			},
			want: false,
		},
		{
			name: "download operation",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "download", Run: "curl -L https://example.com/file.tar.gz", Output: "file"},
				},
			},
			want: true, // Downloads need progress
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requiresProgress(tt.pipeline)
			assert.Equal(t, tt.want, got)
		})
	}
}
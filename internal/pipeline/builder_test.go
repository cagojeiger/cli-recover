package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *Pipeline
		expected string
		wantErr  bool
	}{
		{
			name: "simple linear pipeline",
			pipeline: &Pipeline{
				Name: "test",
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data"},
					{Name: "step2", Run: "cat", Input: "data"},
				},
			},
			expected: "echo hello | cat | tee /tmp/cli-pipe-debug.log",
			wantErr:  false,
		},
		{
			name: "single step pipeline",
			pipeline: &Pipeline{
				Name: "single",
				Steps: []Step{
					{Name: "only", Run: "ls -la"},
				},
			},
			expected: "ls -la",
			wantErr:  false,
		},
		{
			name: "three step pipeline",
			pipeline: &Pipeline{
				Name: "three-steps",
				Steps: []Step{
					{Name: "generate", Run: "echo test", Output: "text"},
					{Name: "transform", Run: "tr a-z A-Z", Input: "text", Output: "upper"},
					{Name: "count", Run: "wc -w", Input: "upper"},
				},
			},
			expected: "echo test | tr a-z A-Z | wc -w | tee /tmp/cli-pipe-debug.log",
			wantErr:  false,
		},
		{
			name: "pipeline with file output",
			pipeline: &Pipeline{
				Name: "file-output",
				Steps: []Step{
					{Name: "generate", Run: "echo data", Output: "stream"},
					{Name: "save", Run: "cat", Input: "stream", Output: "file:output.txt"},
				},
			},
			expected: "echo data | cat | tee /tmp/cli-pipe-debug.log",
			wantErr:  false,
		},
		{
			name: "multiline command",
			pipeline: &Pipeline{
				Name: "multiline",
				Steps: []Step{
					{
						Name: "multi",
						Run: `echo "line1"
echo "line2"`,
					},
				},
			},
			expected: `(echo "line1"
echo "line2")`,
			wantErr: false,
		},
		{
			name: "invalid pipeline - wrong input reference",
			pipeline: &Pipeline{
				Name: "invalid",
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data"},
					{Name: "step2", Run: "cat", Input: "wrong-ref"},
				},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "non-linear pipeline",
			pipeline: &Pipeline{
				Name: "branching",
				Steps: []Step{
					{Name: "source", Run: "echo test", Output: "data"},
					{Name: "branch1", Run: "cat", Input: "data", Output: "out1"},
					{Name: "branch2", Run: "wc", Input: "data", Output: "out2"},
				},
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildCommand(tt.pipeline)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}



func TestIsFileOutput(t *testing.T) {
	tests := []struct {
		output string
		want   bool
	}{
		{"file:output.txt", true},
		{"file:data.json", true},
		{"stream-name", false},
		{"", false},
		{"file:", true},
	}

	for _, tt := range tests {
		t.Run(tt.output, func(t *testing.T) {
			got := IsFileOutput(tt.output)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractFilename(t *testing.T) {
	tests := []struct {
		output string
		want   string
	}{
		{"file:output.txt", "output.txt"},
		{"file:data.json", "data.json"},
		{"stream-name", ""},
		{"", ""},
		{"file:", ""},
	}

	for _, tt := range tests {
		t.Run(tt.output, func(t *testing.T) {
			got := ExtractFilename(tt.output)
			assert.Equal(t, tt.want, got)
		})
	}
}


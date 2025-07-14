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
			expected: "echo hello | cat",
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
			expected: "echo test | tr a-z A-Z | wc -w",
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
			expected: "echo data | cat",
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



func TestBuildSmartCommand(t *testing.T) {
	t.Run("always returns UnifiedMonitor", func(t *testing.T) {
		step := Step{
			Name: "test",
			Run:  "echo hello",
		}
		
		cmd, monitors := BuildSmartCommand(step)
		
		assert.Equal(t, "echo hello", cmd)
		assert.Len(t, monitors, 1)
		
		// Verify monitor is UnifiedMonitor
		_, ok := monitors[0].(*UnifiedMonitor)
		assert.True(t, ok)
	})
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

func TestBuildEnhancedPipeline(t *testing.T) {
	t.Run("valid pipeline", func(t *testing.T) {
		p := &Pipeline{
			Name: "test",
			Steps: []Step{
				{Name: "step1", Run: "echo hello", Output: "data"},
				{Name: "step2", Run: "cat", Input: "data"},
			},
		}
		
		executions, err := BuildEnhancedPipeline(p)
		
		assert.NoError(t, err)
		assert.Len(t, executions, 2)
		
		// Verify each execution has UnifiedMonitor
		for _, exec := range executions {
			assert.Len(t, exec.Monitors, 1)
			_, ok := exec.Monitors[0].(*UnifiedMonitor)
			assert.True(t, ok)
		}
	})
	
	t.Run("invalid pipeline", func(t *testing.T) {
		p := &Pipeline{
			Name: "invalid",
			Steps: []Step{}, // Empty steps
		}
		
		_, err := BuildEnhancedPipeline(p)
		assert.Error(t, err)
	})
}


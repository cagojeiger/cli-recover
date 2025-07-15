package pipeline

import (
	"sort"
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
			expected: "echo hello | cat | tee /tmp/test-logs/pipeline.out",
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
			expected: "echo test | tr a-z A-Z | wc -w | tee /tmp/test-logs/pipeline.out",
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
			expected: "echo data | cat | tee /tmp/test-logs/pipeline.out",
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
			// Use temporary directory for test
			result, err := BuildCommand(tt.pipeline, "/tmp/test-logs")
			
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

func TestBuildDependencyGraph(t *testing.T) {
	tests := []struct {
		name     string
		steps    []Step
		expected map[string]struct {
			parents  []string
			children []string
		}
	}{
		{
			name: "simple linear chain",
			steps: []Step{
				{Name: "fetch", Run: "curl", Output: "raw"},
				{Name: "process", Run: "jq", Input: "raw", Output: "data"},
				{Name: "save", Run: "cat", Input: "data"},
			},
			expected: map[string]struct {
				parents  []string
				children []string
			}{
				"fetch":   {parents: []string{}, children: []string{"process"}},
				"process": {parents: []string{"fetch"}, children: []string{"save"}},
				"save":    {parents: []string{"process"}, children: []string{}},
			},
		},
		{
			name: "simple branching",
			steps: []Step{
				{Name: "fetch", Run: "curl", Output: "raw"},
				{Name: "backup", Run: "gzip", Input: "raw"},
				{Name: "process", Run: "jq", Input: "raw", Output: "data"},
				{Name: "analyze", Run: "python", Input: "data"},
			},
			expected: map[string]struct {
				parents  []string
				children []string
			}{
				"fetch":   {parents: []string{}, children: []string{"backup", "process"}},
				"backup":  {parents: []string{"fetch"}, children: []string{}},
				"process": {parents: []string{"fetch"}, children: []string{"analyze"}},
				"analyze": {parents: []string{"process"}, children: []string{}},
			},
		},
		{
			name: "multiple trees",
			steps: []Step{
				{Name: "tree1_root", Run: "echo 1", Output: "data1"},
				{Name: "tree1_leaf", Run: "cat", Input: "data1"},
				{Name: "tree2_root", Run: "echo 2", Output: "data2"},
				{Name: "tree2_leaf", Run: "wc", Input: "data2"},
			},
			expected: map[string]struct {
				parents  []string
				children []string
			}{
				"tree1_root": {parents: []string{}, children: []string{"tree1_leaf"}},
				"tree1_leaf": {parents: []string{"tree1_root"}, children: []string{}},
				"tree2_root": {parents: []string{}, children: []string{"tree2_leaf"}},
				"tree2_leaf": {parents: []string{"tree2_root"}, children: []string{}},
			},
		},
		{
			name: "complex tree",
			steps: []Step{
				{Name: "root", Run: "curl", Output: "api_data"},
				{Name: "parse", Run: "jq", Input: "api_data", Output: "json"},
				{Name: "users", Run: "jq .users", Input: "json", Output: "user_list"},
				{Name: "logs", Run: "jq .logs", Input: "json", Output: "log_list"},
				{Name: "active", Run: "grep active", Input: "user_list"},
				{Name: "errors", Run: "grep ERROR", Input: "log_list"},
				{Name: "count", Run: "wc -l", Input: "user_list"},
			},
			expected: map[string]struct {
				parents  []string
				children []string
			}{
				"root":   {parents: []string{}, children: []string{"parse"}},
				"parse":  {parents: []string{"root"}, children: []string{"users", "logs"}},
				"users":  {parents: []string{"parse"}, children: []string{"active", "count"}},
				"logs":   {parents: []string{"parse"}, children: []string{"errors"}},
				"active": {parents: []string{"users"}, children: []string{}},
				"errors": {parents: []string{"logs"}, children: []string{}},
				"count":  {parents: []string{"users"}, children: []string{}},
			},
		},
		{
			name: "isolated steps",
			steps: []Step{
				{Name: "isolated1", Run: "date"},
				{Name: "connected1", Run: "echo", Output: "data"},
				{Name: "connected2", Run: "cat", Input: "data"},
				{Name: "isolated2", Run: "whoami"},
			},
			expected: map[string]struct {
				parents  []string
				children []string
			}{
				"isolated1":  {parents: []string{}, children: []string{}},
				"connected1": {parents: []string{}, children: []string{"connected2"}},
				"connected2": {parents: []string{"connected1"}, children: []string{}},
				"isolated2":  {parents: []string{}, children: []string{}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := buildDependencyGraph(tt.steps)
			
			// Verify all steps are in the graph
			assert.Equal(t, len(tt.steps), len(graph))
			
			// Verify each node's relationships
			for stepName, expected := range tt.expected {
				node, exists := graph[stepName]
				assert.True(t, exists, "Step %s should exist in graph", stepName)
				
				// Sort for consistent comparison
				sort.Strings(node.Parents)
				sort.Strings(node.Children)
				sort.Strings(expected.parents)
				sort.Strings(expected.children)
				
				assert.Equal(t, expected.parents, node.Parents, 
					"Step %s parents mismatch", stepName)
				assert.Equal(t, expected.children, node.Children,
					"Step %s children mismatch", stepName)
			}
		})
	}
}


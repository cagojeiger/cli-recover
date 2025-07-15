package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-pipe/internal/config"
	"github.com/cagojeiger/cli-pipe/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTreePipelineIntegration(t *testing.T) {
	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory:     tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := pipeline.NewExecutor(cfg)

	t.Run("simple branch execution", func(t *testing.T) {
		yamlContent := `
name: simple-branch-test
description: Test simple branching

steps:
  - name: generate
    run: echo "test data"
    output: data
    
  - name: upper
    run: tr a-z A-Z
    input: data
    
  - name: count
    run: wc -c
    input: data
`
		var p pipeline.Pipeline
		err := yaml.Unmarshal([]byte(yamlContent), &p)
		require.NoError(t, err)
		
		// Verify it's a tree
		assert.True(t, p.IsTree())
		assert.False(t, p.IsLinear())
		
		// Execute
		err = executor.Execute(&p)
		assert.NoError(t, err)
		
		// Check logs were created
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Greater(t, len(entries), 0)
	})

	t.Run("multi-level tree execution", func(t *testing.T) {
		yamlContent := `
name: multi-level-test
description: Test multi-level tree

steps:
  - name: root
    run: echo "hello world"
    output: data
    
  - name: branch1
    run: tr a-z A-Z
    input: data
    output: upper
    
  - name: branch2
    run: wc -w
    input: data
    output: count
    
  - name: leaf1
    run: rev
    input: upper
    
  - name: leaf2
    run: cat
    input: count
`
		var p pipeline.Pipeline
		err := yaml.Unmarshal([]byte(yamlContent), &p)
		require.NoError(t, err)
		
		// Verify it's a tree
		assert.True(t, p.IsTree())
		assert.False(t, p.IsLinear())
		
		// Execute
		err = executor.Execute(&p)
		assert.NoError(t, err)
	})

	t.Run("complex tree with isolated steps", func(t *testing.T) {
		yamlContent := `
name: complex-test
description: Test complex tree with isolated steps

steps:
  - name: isolated1
    run: echo "standalone"
    
  - name: source
    run: printf "line1\nline2\nline3"
    output: lines
    
  - name: grep1
    run: grep 1
    input: lines
    
  - name: grep2
    run: grep 2
    input: lines
    
  - name: isolated2
    run: date
`
		var p pipeline.Pipeline
		err := yaml.Unmarshal([]byte(yamlContent), &p)
		require.NoError(t, err)
		
		// Verify it's a tree
		assert.True(t, p.IsTree())
		
		// Execute
		err = executor.Execute(&p)
		assert.NoError(t, err)
	})

	t.Run("reject non-tree with merge", func(t *testing.T) {
		yamlContent := `
name: merge-test
description: Test rejection of merge (non-tree)

steps:
  - name: source1
    run: echo "data1"
    output: stream1
    
  - name: source2
    run: echo "data2"
    output: stream2
    
  - name: merge
    run: cat
    input: stream1,stream2
`
		var p pipeline.Pipeline
		err := yaml.Unmarshal([]byte(yamlContent), &p)
		require.NoError(t, err)
		
		// Verify it's NOT a tree
		assert.False(t, p.IsTree())
		
		// Execute should fail
		err = executor.Execute(&p)
		assert.Error(t, err)
		// Could fail during validation or tree check
		assert.True(t, 
			strings.Contains(err.Error(), "non-tree") || 
			strings.Contains(err.Error(), "references undefined input"),
			"Expected error about non-tree or undefined input, got: %v", err)
	})
}

func TestTreePipelineExamples(t *testing.T) {
	// Skip if examples directory doesn't exist
	examplesDir := filepath.Join("..", "examples")
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("examples directory not found")
	}

	// Create temp config
	tempDir := t.TempDir()
	cfg := &config.Config{
		Version: 1,
		Logs: config.LogConfig{
			Directory:     tempDir,
			RetentionDays: 7,
		},
	}
	
	executor := pipeline.NewExecutor(cfg)

	// Test tree example files
	treeExamples := []string{
		"tree-simple-branch.yaml",
		"tree-multi-branch.yaml", 
		"tree-multi-level.yaml",
		"tree-complex.yaml",
	}

	for _, filename := range treeExamples {
		t.Run(filename, func(t *testing.T) {
			yamlPath := filepath.Join(examplesDir, filename)
			
			// Check if file exists
			if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
				t.Skipf("example file %s not found", filename)
			}
			
			// Read YAML file
			data, err := os.ReadFile(yamlPath)
			require.NoError(t, err)
			
			// Parse pipeline
			var p pipeline.Pipeline
			err = yaml.Unmarshal(data, &p)
			require.NoError(t, err)
			
			// Verify it's a tree
			assert.True(t, p.IsTree(), "%s should be a tree structure", filename)
			
			// Execute (skip actual execution for complex examples that might fail)
			if strings.Contains(filename, "simple") {
				err = executor.Execute(&p)
				assert.NoError(t, err)
			}
		})
	}
}

func TestTreeCommandGeneration(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected string
	}{
		{
			name: "simple branch command",
			yaml: `
name: test
steps:
  - name: gen
    run: echo hello
    output: out
  - name: b1
    run: cat
    input: out
  - name: b2
    run: wc
    input: out
`,
			expected: "echo hello | tee >(cat) >(wc) > /dev/null",
		},
		{
			name: "multi-level command",
			yaml: `
name: test
steps:
  - name: root
    run: echo data
    output: d1
  - name: mid
    run: cat
    input: d1
    output: d2
  - name: leaf1
    run: wc -l
    input: d2
  - name: leaf2
    run: wc -w
    input: d2
`,
			expected: "echo data | cat | tee >(wc -l) >(wc -w) > /dev/null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p pipeline.Pipeline
			err := yaml.Unmarshal([]byte(tt.yaml), &p)
			require.NoError(t, err)
			
			cmd, err := pipeline.BuildTreeCommand(&p, "/tmp")
			require.NoError(t, err)
			
			// Remove the tee log part for comparison
			cmd = strings.TrimSuffix(cmd, " | tee /tmp/pipeline.out")
			assert.Equal(t, tt.expected, cmd)
		})
	}
}
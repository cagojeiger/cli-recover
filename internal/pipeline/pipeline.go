package pipeline

import (
	"errors"
	"fmt"
	"strings"
)

// Pipeline represents a series of steps to be executed
type Pipeline struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Steps       []Step `yaml:"steps"`
}

// Step represents a single command in a pipeline
type Step struct {
	Name   string `yaml:"name"`
	Run    string `yaml:"run"`
	Input  string `yaml:"input,omitempty"`
	Output string `yaml:"output,omitempty"`
}

// Validate validates the pipeline structure
func (p *Pipeline) Validate() error {
	if p.Name == "" {
		return errors.New("pipeline name cannot be empty")
	}

	if len(p.Steps) == 0 {
		return errors.New("pipeline must have at least one step")
	}

	// Check for duplicate step names
	stepNames := make(map[string]bool)
	for _, step := range p.Steps {
		if step.Name == "" {
			return errors.New("step name cannot be empty")
		}
		if step.Run == "" {
			return fmt.Errorf("step '%s' has empty command", step.Name)
		}
		if stepNames[step.Name] {
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}
		stepNames[step.Name] = true
	}

	// Build output map
	outputs := make(map[string]bool)
	for _, step := range p.Steps {
		if step.Output != "" {
			outputs[step.Output] = true
		}
	}

	// Check for orphaned inputs
	for _, step := range p.Steps {
		if step.Input != "" && !outputs[step.Input] {
			return fmt.Errorf("step '%s' references undefined input '%s'", step.Name, step.Input)
		}
	}

	return nil
}

// IsLinear checks if the pipeline is a simple linear chain
func (p *Pipeline) IsLinear() bool {
	if len(p.Steps) <= 1 {
		return true
	}

	// Track output usage count
	outputUsage := make(map[string]int)

	// Count how many times each output is used as input
	for _, step := range p.Steps {
		if step.Input != "" {
			outputUsage[step.Input]++
		}
	}

	// Check if any output is used more than once (branching)
	for _, count := range outputUsage {
		if count > 1 {
			return false
		}
	}

	// Check for proper chaining
	for i := 0; i < len(p.Steps)-1; i++ {
		currentOutput := p.Steps[i].Output
		nextInput := p.Steps[i+1].Input

		// Skip if current step has no output
		if currentOutput == "" {
			continue
		}

		// If next step has input, it should match current output for linear flow
		if nextInput != "" && nextInput != currentOutput {
			// Check if this output is used by any later step
			usedLater := false
			for j := i + 1; j < len(p.Steps); j++ {
				if p.Steps[j].Input == currentOutput {
					usedLater = true
					break
				}
			}
			if !usedLater {
				return false
			}
		}
	}

	return true
}

// IsTree checks if the pipeline forms a tree structure (no merges, no cycles)
func (p *Pipeline) IsTree() bool {
	// Empty or single step pipelines are trees
	if len(p.Steps) <= 1 {
		return true
	}

	// Check 1: Each step can have at most one input (no merges)
	for _, step := range p.Steps {
		if step.Input != "" && strings.Contains(step.Input, ",") {
			// Multiple inputs detected (comma-separated)
			return false
		}
	}

	// Check 2: No circular dependencies
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	// Build adjacency list
	adj := make(map[string][]string)
	for _, step := range p.Steps {
		if step.Input != "" {
			// Find the step that produces this input
			for _, producer := range p.Steps {
				if producer.Output == step.Input {
					adj[producer.Name] = append(adj[producer.Name], step.Name)
				}
			}
		}
	}
	
	// Check for cycles using DFS
	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		
		for _, neighbor := range adj[node] {
			if !visited[neighbor] {
				if hasCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}
		
		recStack[node] = false
		return false
	}
	
	// Check all nodes
	for _, step := range p.Steps {
		if !visited[step.Name] {
			if hasCycle(step.Name) {
				return false
			}
		}
	}
	
	return true
}
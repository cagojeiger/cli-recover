package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BuildCommand converts a pipeline to a shell command string with logging to specific directory
func BuildCommand(p *Pipeline, logDir string) (string, error) {
	if len(p.Steps) == 0 {
		return "", fmt.Errorf("empty pipeline")
	}

	// Check if pipeline is linear
	if !p.IsLinear() {
		return "", fmt.Errorf("non-linear pipeline cannot be converted to shell command")
	}

	// For single step, just return the command (no tee needed for single step)
	if len(p.Steps) == 1 {
		return wrapCommand(p.Steps[0].Run), nil
	}

	// Build the pipe chain
	var commands []string
	for _, step := range p.Steps {
		commands = append(commands, wrapCommand(step.Run))
	}

	pipelineCmd := strings.Join(commands, " | ")
	
	// Add tee for pipeline logging to specific directory
	pipelineLog := filepath.Join(logDir, "pipeline.out")
	return pipelineCmd + " | tee " + pipelineLog, nil
}


// wrapCommand wraps multiline commands in parentheses
func wrapCommand(cmd string) string {
	// If command contains newlines, wrap it in parentheses
	if strings.Contains(cmd, "\n") {
		return fmt.Sprintf("(%s)", cmd)
	}
	return cmd
}

// IsFileOutput checks if the output is a file (starts with "file:")
func IsFileOutput(output string) bool {
	return strings.HasPrefix(output, "file:")
}

// ExtractFilename extracts the filename from a file output specifier
func ExtractFilename(output string) string {
	if !IsFileOutput(output) {
		return ""
	}
	return strings.TrimPrefix(output, "file:")
}

// Node represents a step in the dependency graph
type Node struct {
	Step     Step
	Parents  []string
	Children []string
}

// buildDependencyGraph builds a dependency graph from pipeline steps
func buildDependencyGraph(steps []Step) map[string]*Node {
	graph := make(map[string]*Node)
	
	// Initialize nodes
	for _, step := range steps {
		graph[step.Name] = &Node{
			Step:     step,
			Parents:  []string{},
			Children: []string{},
		}
	}
	
	// Build output to producer mapping
	outputProducers := make(map[string]string)
	for _, step := range steps {
		if step.Output != "" {
			outputProducers[step.Output] = step.Name
		}
	}
	
	// Build relationships
	for _, step := range steps {
		if step.Input != "" {
			if producer, exists := outputProducers[step.Input]; exists {
				// Add parent relationship
				graph[step.Name].Parents = append(graph[step.Name].Parents, producer)
				// Add child relationship
				graph[producer].Children = append(graph[producer].Children, step.Name)
			}
		}
	}
	
	return graph
}
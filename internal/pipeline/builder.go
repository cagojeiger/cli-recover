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

// BuildTreeCommand builds a shell command for tree-structured pipelines
func BuildTreeCommand(p *Pipeline, logDir string) (string, error) {
	// Validate it's a tree
	if !p.IsTree() {
		return "", fmt.Errorf("pipeline is not a tree structure")
	}
	
	// If it's linear, use the simple builder
	if p.IsLinear() {
		return BuildCommand(p, logDir)
	}
	
	// Build dependency graph
	graph := buildDependencyGraph(p.Steps)
	
	// Process steps in order, maintaining original step order where possible
	var commands []string
	processed := make(map[string]bool)
	
	// Process each step in the original order
	for _, step := range p.Steps {
		if processed[step.Name] {
			continue
		}
		
		node := graph[step.Name]
		
		// If it's a root node (no parents), process its subtree
		if len(node.Parents) == 0 {
			cmd := buildSubTree(step.Name, graph, processed)
			if cmd != "" {
				commands = append(commands, cmd)
			}
		}
	}
	
	// Join commands
	result := ""
	if len(commands) == 1 {
		result = commands[0]
	} else if len(commands) > 1 {
		// Multiple isolated trees/commands - join with &&
		result = strings.Join(commands, " && ")
	}
	
	// Add final tee for logging
	if result != "" {
		// For multiple commands joined with &&, wrap in parentheses before piping
		if strings.Contains(result, " && ") {
			result = "(" + result + ") | tee " + filepath.Join(logDir, "pipeline.out")
		} else {
			result = result + " | tee " + filepath.Join(logDir, "pipeline.out")
		}
	}
	
	return result, nil
}

// buildSubTree builds command for a subtree starting from given node
func buildSubTree(nodeName string, graph map[string]*Node, processed map[string]bool) string {
	if processed[nodeName] {
		return ""
	}
	
	node := graph[nodeName]
	processed[nodeName] = true
	
	// For isolated nodes (no parents, no children), just return the command
	if len(node.Parents) == 0 && len(node.Children) == 0 {
		return wrapCommand(node.Step.Run)
	}
	
	// Build the command chain
	return buildChain(nodeName, graph, processed)
}

// buildChain builds a linear chain or branching command
func buildChain(nodeName string, graph map[string]*Node, processed map[string]bool) string {
	node := graph[nodeName]
	cmd := wrapCommand(node.Step.Run)
	
	// If this node has exactly one child, continue the chain
	if len(node.Children) == 1 {
		childName := node.Children[0]
		if !processed[childName] {
			processed[childName] = true
			childCmd := buildChain(childName, graph, processed)
			return cmd + " | " + childCmd
		}
	} else if len(node.Children) > 1 {
		// Multiple children - use tee
		var branches []string
		for _, childName := range node.Children {
			if !processed[childName] {
				processed[childName] = true
				// Recursively build the full subtree for each child
				childCmd := buildChain(childName, graph, processed)
				branches = append(branches, fmt.Sprintf(">(%s)", childCmd))
			}
		}
		
		if len(branches) > 0 {
			cmd = cmd + " | tee " + strings.Join(branches, " ") + " > /dev/null"
		}
	}
	
	return cmd
}


// AnalyzeStructure analyzes the pipeline structure in O(n) time
func AnalyzeStructure(p *Pipeline) *Structure {
	structure := &Structure{
		Type:      Linear,
		BranchMap: make(map[string][]string),
	}
	
	// Single pass to build consumer mapping
	for _, step := range p.Steps {
		if step.Input != "" {
			structure.BranchMap[step.Input] = append(structure.BranchMap[step.Input], step.Name)
		}
	}
	
	// Check if any output has multiple consumers
	for _, consumers := range structure.BranchMap {
		if len(consumers) > 1 {
			structure.Type = Tree
			break
		}
	}
	
	return structure
}

// BuildUnifiedCommand builds a command for any pipeline structure
func BuildUnifiedCommand(p *Pipeline, logDir string) (string, error) {
	if len(p.Steps) == 0 {
		return "", fmt.Errorf("empty pipeline")
	}
	
	// Analyze structure
	structure := AnalyzeStructure(p)
	
	// Route to appropriate builder
	if structure.Type == Linear || p.IsLinear() {
		return BuildCommand(p, logDir)
	}
	
	// Use tree builder for tree structures
	return BuildTreeCommand(p, logDir)
}
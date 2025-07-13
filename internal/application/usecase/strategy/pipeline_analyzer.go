package strategy

import (
	"strings"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
)

// isSimpleLinear checks if the pipeline is a simple linear chain without branches
func isSimpleLinear(pipeline *entity.Pipeline) bool {
	if len(pipeline.Steps) <= 1 {
		return true
	}
	
	// Track output usage count
	outputUsage := make(map[string]int)
	
	// Count how many times each output is used as input
	for _, step := range pipeline.Steps {
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
	
	// Check for proper chaining (each step's output should be next step's input)
	for i := 0; i < len(pipeline.Steps)-1; i++ {
		currentOutput := pipeline.Steps[i].Output
		nextInput := pipeline.Steps[i+1].Input
		
		// Skip if current step has no output (might be final step)
		if currentOutput == "" {
			continue
		}
		
		// If next step has input, it should match current output for linear flow
		if nextInput != "" && nextInput != currentOutput {
			// Check if this output is used by any later step
			usedLater := false
			for j := i + 1; j < len(pipeline.Steps); j++ {
				if pipeline.Steps[j].Input == currentOutput {
					usedLater = true
					break
				}
			}
			if !usedLater {
				return false // Output not used, not properly linear
			}
		}
	}
	
	return true
}

// requiresProgress checks if the pipeline contains operations that typically need progress reporting
func requiresProgress(pipeline *entity.Pipeline) bool {
	progressCommands := []string{
		"tar",      // Archive operations
		"gzip",     // Compression
		"curl",     // Downloads
		"wget",     // Downloads
		"scp",      // File transfers
		"rsync",    // Synchronization
		"dd",       // Disk operations
		"pv",       // Pipe viewer itself
	}
	
	for _, step := range pipeline.Steps {
		// Check if the command starts with any progress-needing command
		for _, cmd := range progressCommands {
			if strings.HasPrefix(step.Run, cmd+" ") || step.Run == cmd {
				return true
			}
			// Also check if command contains the operation after pipe or semicolon
			if strings.Contains(step.Run, " "+cmd+" ") ||
			   strings.Contains(step.Run, "|"+cmd+" ") ||
			   strings.Contains(step.Run, ";"+cmd+" ") {
				return true
			}
		}
		
		// Check for large file operations by looking for paths
		if strings.Contains(step.Run, "/data") || 
		   strings.Contains(step.Run, "/backup") ||
		   strings.Contains(step.Run, "http://") ||
		   strings.Contains(step.Run, "https://") {
			return true
		}
	}
	
	return false
}
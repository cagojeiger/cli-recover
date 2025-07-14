package pipeline

import (
	"fmt"
	"strings"
)

// BuildCommand converts a pipeline to a shell command string
func BuildCommand(p *Pipeline) (string, error) {
	if len(p.Steps) == 0 {
		return "", fmt.Errorf("empty pipeline")
	}

	// Check if pipeline is linear
	if !p.IsLinear() {
		return "", fmt.Errorf("non-linear pipeline cannot be converted to shell command")
	}

	// For single step, just return the command
	if len(p.Steps) == 1 {
		return wrapCommand(p.Steps[0].Run), nil
	}

	// Build the pipe chain
	var commands []string
	for _, step := range p.Steps {
		commands = append(commands, wrapCommand(step.Run))
	}

	return strings.Join(commands, " | "), nil
}


// wrapCommand wraps multiline commands in parentheses
func wrapCommand(cmd string) string {
	// If command contains newlines, wrap it in parentheses
	if strings.Contains(cmd, "\n") {
		return fmt.Sprintf("(%s)", cmd)
	}
	return cmd
}

// BuildSmartCommand builds a command with unified monitor
func BuildSmartCommand(step Step) (string, []Monitor) {
	// Always use UnifiedMonitor for all steps
	return step.Run, []Monitor{NewUnifiedMonitor()}
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

// BuildEnhancedPipeline builds a pipeline with unified monitoring
func BuildEnhancedPipeline(p *Pipeline) ([]StepExecution, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	
	var executions []StepExecution
	
	for _, step := range p.Steps {
		exec := StepExecution{
			Step:     step,
			Command:  step.Run,
			Monitors: []Monitor{NewUnifiedMonitor()},
		}
		
		executions = append(executions, exec)
	}
	
	return executions, nil
}

// StepExecution contains execution details for a step
type StepExecution struct {
	Step     Step
	Command  string
	Monitors []Monitor
}
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

// BuildCommandWithLogging creates a shell script with logging
func BuildCommandWithLogging(p *Pipeline, logDir string) (string, error) {
	if len(p.Steps) == 0 {
		return "", fmt.Errorf("empty pipeline")
	}

	// Build the script
	var script strings.Builder
	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -o pipefail\n")
	script.WriteString(fmt.Sprintf("LOGDIR=\"%s\"\n", logDir))
	script.WriteString("mkdir -p \"$LOGDIR\"\n")
	script.WriteString("\n")
	script.WriteString(fmt.Sprintf("# Pipeline: %s\n", p.Name))
	if p.Description != "" {
		script.WriteString(fmt.Sprintf("# %s\n", p.Description))
	}
	script.WriteString("\n")

	if len(p.Steps) == 1 {
		// Single step case
		step := p.Steps[0]
		script.WriteString(fmt.Sprintf("# Step: %s\n", step.Name))
		script.WriteString(fmt.Sprintf("(%s) 2>\"$LOGDIR/%s.err\" | tee \"$LOGDIR/%s.out\"\n",
			step.Run, step.Name, step.Name))
	} else {
		// Multi-step pipeline
		script.WriteString("# Execute pipeline\n")
		var commands []string
		for i, step := range p.Steps {
			if i == 0 {
				// First step only logs stderr
				cmd := fmt.Sprintf("((%s) 2>\"$LOGDIR/%s.err\")",
					step.Run, step.Name)
				commands = append(commands, cmd)
			} else if i == len(p.Steps)-1 {
				// Last step logs both stdout and stderr
				cmd := fmt.Sprintf("((%s) 2>\"$LOGDIR/%s.err\" | tee \"$LOGDIR/%s.out\")",
					step.Run, step.Name, step.Name)
				commands = append(commands, cmd)
			} else {
				// Middle steps only log stderr
				cmd := fmt.Sprintf("((%s) 2>\"$LOGDIR/%s.err\")",
					step.Run, step.Name)
				commands = append(commands, cmd)
			}
		}
		script.WriteString(strings.Join(commands, " | \\\n"))
		script.WriteString("\n")
	}

	script.WriteString("\n")
	script.WriteString("# Save exit code\n")
	script.WriteString("EXIT_CODE=$?\n")
	script.WriteString("echo \"Pipeline exit code: $EXIT_CODE\" > \"$LOGDIR/pipeline.status\"\n")
	script.WriteString("exit $EXIT_CODE\n")

	return script.String(), nil
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
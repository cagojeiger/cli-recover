package strategy

import (
	"fmt"
	"strings"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
)

// buildShellCommand converts a pipeline to a shell command string
func buildShellCommand(pipeline *entity.Pipeline) (string, error) {
	if len(pipeline.Steps) == 0 {
		return "", fmt.Errorf("empty pipeline")
	}

	// Check if pipeline is linear
	if !isSimpleLinear(pipeline) {
		return "", fmt.Errorf("non-linear pipeline cannot be converted to shell command")
	}

	// For single step, just return the command
	if len(pipeline.Steps) == 1 {
		return wrapCommand(pipeline.Steps[0].Run), nil
	}

	// Build the pipe chain
	var commands []string
	for _, step := range pipeline.Steps {
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

// buildShellCommandWithLogging converts a pipeline to a shell script with logging
func buildShellCommandWithLogging(pipeline *entity.Pipeline, logDir string) (string, error) {
	if len(pipeline.Steps) == 0 {
		return "", fmt.Errorf("empty pipeline")
	}

	// Build the script
	var script strings.Builder
	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -o pipefail\n")
	script.WriteString(fmt.Sprintf("LOGDIR=\"%s\"\n", logDir))
	script.WriteString("mkdir -p \"$LOGDIR\"\n")
	script.WriteString("\n")
	script.WriteString(fmt.Sprintf("# Pipeline: %s\n", pipeline.Name))

	if len(pipeline.Steps) == 1 {
		// Single step case
		step := pipeline.Steps[0]
		script.WriteString(fmt.Sprintf("(%s) 2>\"$LOGDIR/%s.err\" | tee \"$LOGDIR/%s.out\"\n",
			step.Run, step.Name, step.Name))
	} else {
		// Multi-step pipeline
		var commands []string
		for _, step := range pipeline.Steps {
			cmd := fmt.Sprintf("((%s) 2>\"$LOGDIR/%s.err\" | tee \"$LOGDIR/%s.out\")",
				step.Run, step.Name, step.Name)
			commands = append(commands, cmd)
		}
		script.WriteString(strings.Join(commands, " | \\\n"))
		script.WriteString("\n")
	}

	script.WriteString("\n")
	script.WriteString("# Save exit code\n")
	script.WriteString("EXIT_CODE=$?\n")
	script.WriteString("echo \"Pipeline exit code: $EXIT_CODE\" > \"$LOGDIR/pipeline.status\"\n")
	script.WriteString("exit $EXIT_CODE")

	return script.String(), nil
}
package kubernetes

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// OSCommandExecutor executes commands using os/exec
type OSCommandExecutor struct{}

// NewOSCommandExecutor creates a new OS command executor
func NewOSCommandExecutor() *OSCommandExecutor {
	return &OSCommandExecutor{}
}

// Execute runs a command and returns the output
func (e *OSCommandExecutor) Execute(ctx context.Context, command []string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("no command provided")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// Stream runs a command and streams the output
func (e *OSCommandExecutor) Stream(ctx context.Context, command []string) (<-chan string, <-chan error) {
	outputCh := make(chan string, 100)
	errorCh := make(chan error, 1)

	go func() {
		defer close(outputCh)
		defer close(errorCh)

		if len(command) == 0 {
			errorCh <- fmt.Errorf("no command provided")
			return
		}

		cmd := exec.CommandContext(ctx, command[0], command[1:]...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errorCh <- fmt.Errorf("failed to create stdout pipe: %w", err)
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			errorCh <- fmt.Errorf("failed to create stderr pipe: %w", err)
			return
		}

		if err := cmd.Start(); err != nil {
			errorCh <- fmt.Errorf("failed to start command: %w", err)
			return
		}

		// Read stderr in a separate goroutine
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				// For now, we'll include stderr in the output
				outputCh <- fmt.Sprintf("STDERR: %s", scanner.Text())
			}
		}()

		// Read stdout
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				cmd.Process.Kill()
				errorCh <- ctx.Err()
				return
			case outputCh <- scanner.Text():
			}
		}

		if err := scanner.Err(); err != nil {
			errorCh <- fmt.Errorf("error reading output: %w", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			errorCh <- fmt.Errorf("command failed: %w", err)
			return
		}
	}()

	return outputCh, errorCh
}

// StreamBinary runs a command and streams binary output safely
func (e *OSCommandExecutor) StreamBinary(ctx context.Context, command []string) (stdout io.ReadCloser, stderr io.ReadCloser, wait func() error, err error) {
	if len(command) == 0 {
		return nil, nil, nil, fmt.Errorf("no command provided")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	// Get pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		stdoutPipe.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		stdoutPipe.Close()
		stderrPipe.Close()
		return nil, nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Wait function to be called when done reading
	waitFunc := func() error {
		return cmd.Wait()
	}

	return stdoutPipe, stderrPipe, waitFunc, nil
}

// EstimateSize estimates the size of a directory in bytes
func EstimateSize(ctx context.Context, executor CommandExecutor, namespace, pod, path string) (int64, error) {
	// Use du command to estimate size
	command := BuildKubectlCommand("exec", "-n", namespace, pod, "--", "du", "-sb", path)
	output, err := executor.Execute(ctx, command)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate size: %w", err)
	}

	// Parse du output: "12345\t/path/to/dir"
	parts := strings.Fields(output)
	if len(parts) < 1 {
		return 0, fmt.Errorf("unexpected du output: %s", output)
	}

	var size int64
	if _, err := fmt.Sscanf(parts[0], "%d", &size); err != nil {
		return 0, fmt.Errorf("failed to parse size: %w", err)
	}

	return size, nil
}

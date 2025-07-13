package pipeline

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ExecutePipelineEnhanced executes an entire pipeline with proper data flow
func (e *Executor) ExecutePipelineEnhanced(p *Pipeline) error {
	e.log("Executing pipeline: %s\n", p.Name)
	if p.Description != "" {
		e.log("Description: %s\n", p.Description)
	}
	
	// For linear pipelines, build and execute as a single command
	if p.IsLinear() {
		return e.executeLinearPipelineEnhanced(p)
	}
	
	// For non-linear pipelines, we need tee-based branching
	return fmt.Errorf("non-linear pipelines not yet supported in enhanced mode")
}

// executeLinearPipelineEnhanced executes a linear pipeline with monitoring
func (e *Executor) executeLinearPipelineEnhanced(p *Pipeline) error {
	// Build the full pipeline command
	shellCmd, err := BuildCommand(p)
	if err != nil {
		return fmt.Errorf("failed to build pipeline command: %w", err)
	}
	
	// Collect all monitors from all steps
	var allMonitors []Monitor
	var outputFile string
	var logFiles []string
	
	for _, step := range p.Steps {
		_, monitors := BuildSmartCommand(step)
		allMonitors = append(allMonitors, monitors...)
		
		// Track output file from last step
		if IsFileOutput(step.Output) {
			outputFile = ExtractFilename(step.Output)
		}
		
		// Track log files
		if step.Log != "" {
			logFiles = append(logFiles, step.Log)
		}
	}
	
	// Create the command
	cmd := exec.Command("bash", "-c", shellCmd)
	
	// Set up pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Create writers
	var writers []io.Writer
	
	// Add output file if specified
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writers = append(writers, f)
	}
	
	// Add monitors as writers
	for _, monitor := range allMonitors {
		if w, ok := monitor.(io.Writer); ok {
			writers = append(writers, w)
		} else if lm, ok := monitor.(*LineMonitor); ok {
			writers = append(writers, NewLineMonitorWriter(lm))
		} else {
			writers = append(writers, NewMonitorWriter(monitor))
		}
	}
	
	// Create TeeWriter if needed
	var tee *TeeWriter
	if len(writers) > 0 {
		tee = NewTeeWriter(writers...)
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pipeline: %w", err)
	}
	
	// Process output
	done := make(chan error)
	go func() {
		if tee != nil {
			io.Copy(io.MultiWriter(os.Stdout, tee), stdout)
		} else {
			io.Copy(os.Stdout, stdout)
		}
		io.Copy(os.Stderr, stderr)
		done <- nil
	}()
	
	// Wait for completion
	cmdErr := cmd.Wait()
	<-done
	
	// Close TeeWriter
	if tee != nil {
		tee.Close()
	}
	
	// Finish monitors
	for _, monitor := range allMonitors {
		monitor.Finish()
	}
	
	// Save checksums
	for _, monitor := range allMonitors {
		if cfw, ok := monitor.(*ChecksumFileWriter); ok {
			if err := cfw.SaveToFile(); err != nil {
				e.log("Warning: failed to save checksum: %v\n", err)
			}
		}
	}
	
	// Report monitor results
	if len(allMonitors) > 0 {
		e.log("\nMonitor reports:\n")
		for _, monitor := range allMonitors {
			e.log("  %s\n", monitor.Report())
		}
	}
	
	if cmdErr != nil {
		return fmt.Errorf("pipeline failed: %w", cmdErr)
	}
	
	e.log("\nPipeline completed successfully\n")
	return nil
}

// ExecuteStepWithDataFlow executes a single step with proper input/output handling
func (e *Executor) ExecuteStepWithDataFlow(step Step, inputData string) (string, error) {
	// Build command with input handling
	var cmd string
	if step.Input != "" && inputData != "" {
		// Create a temporary file with input data
		tmpFile, err := os.CreateTemp("", "pipe-input-*.txt")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		
		if _, err := tmpFile.WriteString(inputData); err != nil {
			return "", fmt.Errorf("failed to write input data: %w", err)
		}
		tmpFile.Close()
		
		// Modify command to read from temp file
		cmd = fmt.Sprintf("cat %s | %s", tmpFile.Name(), step.Run)
	} else {
		cmd = step.Run
	}
	
	// Execute command
	execCmd := exec.Command("bash", "-c", cmd)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	
	result := strings.TrimSpace(string(output))
	
	// Handle file output
	if IsFileOutput(step.Output) {
		filename := ExtractFilename(step.Output)
		if err := os.WriteFile(filename, []byte(result), 0644); err != nil {
			return "", fmt.Errorf("failed to write output file: %w", err)
		}
	}
	
	return result, nil
}
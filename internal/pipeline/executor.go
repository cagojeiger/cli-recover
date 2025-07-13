package pipeline

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Executor handles pipeline execution
type Executor struct {
	logDir    string
	logWriter io.Writer
}

// Option is a functional option for Executor
type Option func(*Executor)

// WithLogDir sets the log directory
func WithLogDir(dir string) Option {
	return func(e *Executor) {
		e.logDir = dir
	}
}

// WithLogWriter sets the log writer for console output
func WithLogWriter(w io.Writer) Option {
	return func(e *Executor) {
		e.logWriter = w
	}
}

// NewExecutor creates a new Executor
func NewExecutor(opts ...Option) *Executor {
	e := &Executor{
		logWriter: os.Stdout,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Execute runs the pipeline using Unix pipes
func (e *Executor) Execute(p *Pipeline) error {
	// Validate pipeline
	if err := p.Validate(); err != nil {
		return fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Log pipeline execution start
	e.log("Executing pipeline: %s\n", p.Name)
	if p.Description != "" {
		e.log("Description: %s\n", p.Description)
	}

	// Execute based on logging requirement
	if e.logDir != "" {
		return e.executeWithLogging(p)
	}
	return e.executeSimple(p)
}

// executeSimple runs the pipeline without file logging
func (e *Executor) executeSimple(p *Pipeline) error {
	// Build shell command
	shellCmd, err := BuildCommand(p)
	if err != nil {
		return fmt.Errorf("failed to build shell command: %w", err)
	}

	e.log("\nCommand: %s\n\n", shellCmd)

	// Execute using bash
	cmd := exec.Command("bash", "-c", shellCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	if err != nil {
		e.log("\nPipeline failed after %v: %v\n", duration, err)
		return fmt.Errorf("command failed: %w", err)
	}

	e.log("\nPipeline completed successfully in %v\n", duration)
	return nil
}

// executeWithLogging runs the pipeline with file logging
func (e *Executor) executeWithLogging(p *Pipeline) error {
	// Create log directory with timestamp
	timestamp := time.Now().Format("20060102_150405")
	runDir := filepath.Join(e.logDir, fmt.Sprintf("%s_%s", p.Name, timestamp))
	
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	e.log("Logging to: %s\n", runDir)

	// Build shell script with logging
	script, err := BuildCommandWithLogging(p, runDir)
	if err != nil {
		return fmt.Errorf("failed to build logging script: %w", err)
	}

	// Write script to file for debugging
	scriptPath := filepath.Join(runDir, "pipeline.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	// Execute the script
	cmd := exec.Command("bash", "-c", script)
	
	// Capture output for console
	var stdout, stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	// Write console output to log files
	if stdout.Len() > 0 {
		os.WriteFile(filepath.Join(runDir, "console.out"), stdout.Bytes(), 0644)
	}
	if stderr.Len() > 0 {
		os.WriteFile(filepath.Join(runDir, "console.err"), stderr.Bytes(), 0644)
	}

	// Write execution summary
	summary := fmt.Sprintf("Pipeline: %s\nDuration: %v\nStatus: ", p.Name, duration)
	if err != nil {
		summary += fmt.Sprintf("Failed\nError: %v\n", err)
		e.log("\nPipeline failed after %v: %v\n", duration, err)
	} else {
		summary += "Success\n"
		e.log("\nPipeline completed successfully in %v\n", duration)
	}
	os.WriteFile(filepath.Join(runDir, "summary.txt"), []byte(summary), 0644)

	return err
}

// log writes to the log writer
func (e *Executor) log(format string, args ...interface{}) {
	if e.logWriter != nil {
		fmt.Fprintf(e.logWriter, format, args...)
	}
}

// CaptureOutput runs a pipeline and captures its output
func (e *Executor) CaptureOutput(p *Pipeline) (string, error) {
	// Build shell command
	shellCmd, err := BuildCommand(p)
	if err != nil {
		return "", fmt.Errorf("failed to build shell command: %w", err)
	}

	// Execute using bash
	cmd := exec.Command("bash", "-c", shellCmd)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ExecuteEnhanced executes a single step with monitoring and logging
func (e *Executor) ExecuteEnhanced(step Step) error {
	e.log("Executing step: %s\n", step.Name)
	
	// Build command and monitors
	cmd, monitors := BuildSmartCommand(step)
	
	// Create the command
	execCmd := exec.Command("bash", "-c", cmd)
	
	// Set up pipes for stdout and stderr
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Create writers for various outputs
	var writers []io.Writer
	
	// Handle file output
	var outputFile *os.File
	if IsFileOutput(step.Output) {
		filename := ExtractFilename(step.Output)
		outputFile, err = os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()
		writers = append(writers, outputFile)
	}
	
	// Handle log file
	var logFile *os.File
	if step.Log != "" {
		logFile, err = os.Create(step.Log)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		defer logFile.Close()
		writers = append(writers, logFile)
	}
	
	// Add monitors that implement io.Writer
	for _, monitor := range monitors {
		if w, ok := monitor.(io.Writer); ok {
			writers = append(writers, w)
		}
	}
	
	// Create TeeWriter for independent writing
	var tee *TeeWriter
	if len(writers) > 0 {
		tee = NewTeeWriter(writers...)
	}
	
	// Start the command
	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
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
	
	// Wait for command to complete
	cmdErr := execCmd.Wait()
	<-done
	
	// Close TeeWriter
	if tee != nil {
		tee.Close()
	}
	
	// Finish monitors
	for _, monitor := range monitors {
		monitor.Finish()
	}
	
	// Save checksums
	for _, monitor := range monitors {
		if cfw, ok := monitor.(*ChecksumFileWriter); ok {
			if err := cfw.SaveToFile(); err != nil {
				e.log("Warning: failed to save checksum: %v\n", err)
			}
		}
	}
	
	// Report monitor results
	if len(monitors) > 0 {
		e.log("\nMonitor reports:\n")
		for _, monitor := range monitors {
			e.log("  %s\n", monitor.Report())
		}
	}
	
	if cmdErr != nil {
		return cmdErr
	}
	
	return nil
}

// ExecutePipeline executes an entire pipeline with enhanced features
func (e *Executor) ExecutePipeline(p *Pipeline) error {
	e.log("Executing pipeline: %s\n", p.Name)
	if p.Description != "" {
		e.log("Description: %s\n", p.Description)
	}
	
	// For now, execute steps sequentially
	// TODO: Implement smart tee-based branching for non-linear pipelines
	if !p.IsLinear() {
		return fmt.Errorf("non-linear pipelines not yet supported in enhanced mode")
	}
	
	// Execute each step
	for i, step := range p.Steps {
		e.log("\n[Step %d/%d] %s\n", i+1, len(p.Steps), step.Name)
		
		if err := e.ExecuteEnhanced(step); err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}
	}
	
	e.log("\nPipeline completed successfully\n")
	return nil
}
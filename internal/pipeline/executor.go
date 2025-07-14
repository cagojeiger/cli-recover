package pipeline

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/cagojeiger/cli-pipe/internal/config"
)

// Executor handles pipeline execution
type Executor struct {
	config    *config.Config
	logWriter io.Writer
}

// NewExecutor creates a new Executor with config
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{
		config:    cfg,
		logWriter: os.Stdout,
	}
}

// Execute runs the pipeline with monitoring and logging
func (e *Executor) Execute(p *Pipeline) error {
	// Validate pipeline
	if err := p.Validate(); err != nil {
		return fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Ensure config is loaded
	if e.config == nil {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		e.config = cfg
	}

	// Create log directory
	timestamp := time.Now().Format("20060102_150405")
	logDir := filepath.Join(e.config.Logs.Directory, fmt.Sprintf("%s_%s", p.Name, timestamp))
	
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	e.log("Executing pipeline: %s\n", p.Name)
	if p.Description != "" {
		e.log("Description: %s\n", p.Description)
	}
	e.log("Logging to: %s\n", logDir)

	// For linear pipelines, execute as single command
	if p.IsLinear() {
		return e.executeLinearPipeline(p, logDir)
	}
	
	// Non-linear pipelines not yet supported
	return fmt.Errorf("non-linear pipelines not yet supported")
}

// executeLinearPipeline executes a linear pipeline with unified monitoring
func (e *Executor) executeLinearPipeline(p *Pipeline, logDir string) error {
	// Build shell command
	shellCmd, err := BuildCommand(p)
	if err != nil {
		return fmt.Errorf("failed to build pipeline command: %w", err)
	}
	
	e.log("\nCommand: %s\n\n", shellCmd)
	
	// Create unified monitor for entire pipeline
	monitor := NewUnifiedMonitor()
	
	// Track output file from last step
	var outputFile string
	for _, step := range p.Steps {
		if IsFileOutput(step.Output) {
			outputFile = ExtractFilename(step.Output)
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
	
	// Add console output
	writers = append(writers, os.Stdout)
	
	// Add monitor
	writers = append(writers, monitor)
	
	// Add log file
	logFile, err := os.Create(filepath.Join(logDir, "pipeline.log"))
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()
	writers = append(writers, logFile)
	
	// Add output file if specified
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writers = append(writers, f)
	}
	
	// Create multi-writer
	multiWriter := io.MultiWriter(writers...)
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pipeline: %w", err)
	}
	
	// Process output
	done := make(chan error, 2)
	
	go func() {
		io.Copy(multiWriter, stdout)
		done <- nil
	}()
	
	go func() {
		stderrFile, _ := os.Create(filepath.Join(logDir, "stderr.log"))
		defer stderrFile.Close()
		io.Copy(io.MultiWriter(os.Stderr, stderrFile), stderr)
		done <- nil
	}()
	
	// Wait for command completion
	cmdErr := cmd.Wait()
	
	// Wait for goroutines
	<-done
	<-done
	
	// Finish monitoring
	monitor.Finish()
	
	// Write summary
	duration := monitor.GetDuration()
	status := "Success"
	if cmdErr != nil {
		status = fmt.Sprintf("Failed: %v", cmdErr)
	}
	
	summary := fmt.Sprintf("Pipeline: %s\nDuration: %v\nBytes: %s\nLines: %d\nStatus: %s\n",
		p.Name,
		duration,
		formatBytes(monitor.GetBytes()),
		monitor.GetLines(),
		status)
	
	os.WriteFile(filepath.Join(logDir, "summary.txt"), []byte(summary), 0644)
	
	// Report results
	e.log("\n%s\n", strings.Repeat("=", 50))
	e.log("Pipeline completed\n")
	e.log("• Duration: %v\n", duration)
	e.log("• Bytes processed: %s\n", formatBytes(monitor.GetBytes()))
	e.log("• Lines processed: %d\n", monitor.GetLines())
	e.log("• Status: %s\n", status)
	e.log("• Logs: %s\n", logDir)
	
	return cmdErr
}

// log writes to the log writer
func (e *Executor) log(format string, args ...interface{}) {
	if e.logWriter != nil {
		fmt.Fprintf(e.logWriter, format, args...)
	}
}

// formatBytes formats bytes in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
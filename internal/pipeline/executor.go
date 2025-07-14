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
	"github.com/cagojeiger/cli-pipe/internal/logger"
)

// Executor handles pipeline execution
type Executor struct {
	config    *config.Config
	logWriter io.Writer  // For backward compatibility with console output
	logger    logger.Logger
}

// NewExecutor creates a new Executor with config
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{
		config:    cfg,
		logWriter: os.Stdout,
		logger:    logger.Default(),
	}
}

// Execute runs the pipeline with monitoring and logging
func (e *Executor) Execute(p *Pipeline) error {
	// Ensure logger exists
	if e.logger == nil {
		e.logger = logger.Default()
	}
	
	log := e.logger.With("pipeline", p.Name)
	log.Info("starting pipeline execution")
	
	// Validate pipeline
	if err := p.Validate(); err != nil {
		log.Error("pipeline validation failed", "error", err)
		return fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Ensure config is loaded
	if e.config == nil {
		log.Debug("loading config")
		cfg, err := config.Load()
		if err != nil {
			log.Error("failed to load config", "error", err)
			return fmt.Errorf("failed to load config: %w", err)
		}
		e.config = cfg
	}

	// Create log directory
	timestamp := time.Now().Format("20060102_150405")
	logDir := filepath.Join(e.config.Logs.Directory, fmt.Sprintf("%s_%s", p.Name, timestamp))
	
	log.Debug("creating log directory", "path", logDir)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Error("failed to create log directory", "path", logDir, "error", err)
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Clean old logs if retention is configured
	if e.config.Logs.RetentionDays > 0 {
		cleaner := logger.NewLogCleaner(log)
		go func() {
			log.Debug("starting background log cleanup", "retention_days", e.config.Logs.RetentionDays)
			if err := cleaner.CleanOldLogs(e.config.Logs.Directory, e.config.Logs.RetentionDays); err != nil {
				log.Error("failed to clean old logs", "error", err)
			}
		}()
	}

	// Log detailed pipeline information
	log.Info("executing pipeline",
		"description", p.Description,
		"log_directory", logDir,
		"steps", len(p.Steps))
	
	// Log each step information
	for i, step := range p.Steps {
		log.Debug("pipeline step configured",
			"index", i,
			"name", step.Name,
			"command", step.Run,
			"input", step.Input,
			"output", step.Output)
	}
	
	// Keep console output for user-facing messages
	e.log("Executing pipeline: %s\n", p.Name)
	if p.Description != "" {
		e.log("Description: %s\n", p.Description)
	}
	e.log("Logging to: %s\n", logDir)

	// For linear pipelines, execute as single command
	if p.IsLinear() {
		return e.executeLinearPipeline(p, logDir, log)
	}
	
	// Non-linear pipelines not yet supported
	log.Error("non-linear pipelines not yet supported")
	return fmt.Errorf("non-linear pipelines not yet supported")
}

// executeLinearPipeline executes a linear pipeline with unified monitoring
func (e *Executor) executeLinearPipeline(p *Pipeline, logDir string, log logger.Logger) error {
	log.Debug("building shell command")
	// Build shell command
	shellCmd, err := BuildCommand(p)
	if err != nil {
		log.Error("failed to build pipeline command", "error", err)
		return fmt.Errorf("failed to build pipeline command: %w", err)
	}
	log.Debug("shell command built", "command", shellCmd)
	
	// Display command structure
	e.log("\nPipeline structure:\n")
	for i, step := range p.Steps {
		e.log("  [%d] %s\n", i+1, step.Name)
		if step.Run != "" {
			e.log("      Command: %s\n", step.Run)
		}
		if step.Input != "" && step.Input != "stdin" {
			e.log("      Input: %s\n", step.Input)
		}
		if step.Output != "" && step.Output != "stdout" {
			e.log("      Output: %s\n", step.Output)
		}
	}
	e.log("\nFull command: %s\n\n", shellCmd)
	
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
	log.Debug("creating command")
	cmd := exec.Command("bash", "-c", shellCmd)
	
	// Set up pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("failed to create stdout pipe", "error", err)
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Error("failed to create stderr pipe", "error", err)
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Create writers
	var writers []io.Writer
	
	// Add console output
	writers = append(writers, os.Stdout)
	
	// Add monitor
	writers = append(writers, monitor)
	
	// Add log file
	logFilePath := filepath.Join(logDir, "pipeline.log")
	log.Debug("creating log file", "path", logFilePath)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Error("failed to create log file", "path", logFilePath, "error", err)
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()
	writers = append(writers, logFile)
	
	// Add output file if specified
	if outputFile != "" {
		log.Debug("creating output file", "path", outputFile)
		f, err := os.Create(outputFile)
		if err != nil {
			log.Error("failed to create output file", "path", outputFile, "error", err)
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writers = append(writers, f)
	}
	
	// Create multi-writer
	multiWriter := io.MultiWriter(writers...)
	
	// Start the command
	log.Info("starting pipeline command")
	if err := cmd.Start(); err != nil {
		log.Error("failed to start pipeline", "error", err)
		return fmt.Errorf("failed to start pipeline: %w", err)
	}
	
	// Process output
	done := make(chan error, 2)
	
	go func() {
		log.Debug("processing stdout")
		n, err := io.Copy(multiWriter, stdout)
		log.Debug("stdout processing complete", "bytes", n, "error", err)
		done <- err
	}()
	
	go func() {
		log.Debug("processing stderr")
		stderrPath := filepath.Join(logDir, "stderr.log")
		stderrFile, err := os.Create(stderrPath)
		if err != nil {
			log.Error("failed to create stderr log", "path", stderrPath, "error", err)
			done <- err
			return
		}
		defer stderrFile.Close()
		n, err := io.Copy(io.MultiWriter(os.Stderr, stderrFile), stderr)
		log.Debug("stderr processing complete", "bytes", n, "error", err)
		done <- err
	}()
	
	// Wait for command completion
	log.Debug("waiting for command completion")
	cmdErr := cmd.Wait()
	
	// Wait for goroutines
	err1 := <-done
	err2 := <-done
	
	// Check for goroutine errors
	if err1 != nil {
		log.Error("error processing output", "error", err1)
	}
	if err2 != nil {
		log.Error("error processing output", "error", err2)
	}
	
	// Finish monitoring
	monitor.Finish()
	
	if cmdErr != nil {
		log.Error("pipeline command failed", "error", cmdErr)
	} else {
		log.Info("pipeline command completed successfully")
	}
	
	// Write summary
	duration := monitor.GetDuration()
	status := "Success"
	if cmdErr != nil {
		status = fmt.Sprintf("Failed: %v", cmdErr)
	}
	
	bytesProcessed := monitor.GetBytes()
	linesProcessed := monitor.GetLines()
	
	log.Info("pipeline execution completed",
		"duration", duration,
		"bytes_processed", bytesProcessed,
		"lines_processed", linesProcessed,
		"status", status)
	
	summary := fmt.Sprintf("Pipeline: %s\nDuration: %v\nBytes: %s\nLines: %d\nStatus: %s\n",
		p.Name,
		duration,
		formatBytes(bytesProcessed),
		linesProcessed,
		status)
	
	summaryPath := filepath.Join(logDir, "summary.txt")
	log.Debug("writing summary", "path", summaryPath)
	os.WriteFile(summaryPath, []byte(summary), 0644)
	
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
	// Ensure logger exists
	if e.logger == nil {
		e.logger = logger.Default()
	}
	
	log := e.logger.With("pipeline", p.Name, "mode", "capture")
	log.Debug("capturing pipeline output")
	
	// Build shell command
	shellCmd, err := BuildCommand(p)
	if err != nil {
		log.Error("failed to build shell command", "error", err)
		return "", fmt.Errorf("failed to build shell command: %w", err)
	}

	// Execute using bash
	log.Debug("executing command for capture", "command", shellCmd)
	cmd := exec.Command("bash", "-c", shellCmd)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Error("command failed", "error", err, "output", string(output))
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	log.Debug("command completed successfully", "output_length", len(output))
	return strings.TrimSpace(string(output)), nil
}
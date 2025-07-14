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

// OutputLimiter limits output to screen while preserving full output in logs
type OutputLimiter struct {
	maxLines    int
	currentLine int
	prefix      string
	truncated   bool
}

// Write implements io.Writer for output limiting
func (ol *OutputLimiter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	
	for i, line := range lines {
		// Skip last empty line from split
		if i == len(lines)-1 && line == "" {
			continue
		}
		
		ol.currentLine++
		
		if ol.currentLine <= ol.maxLines {
			fmt.Printf("%s%s\n", ol.prefix, line)
		} else if !ol.truncated {
			truncatedCount := len(lines) - i + 1
			fmt.Printf("... (%d more lines in log file)\n", truncatedCount)
			ol.truncated = true
			break
		}
	}
	
	return len(p), nil
}

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

// Execute runs the pipeline with logging
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

	// Simple pipeline information (1단계: 단순화)
	log.Info("executing pipeline",
		"description", p.Description,
		"steps", len(p.Steps))
	
	// Simple console output
	e.log("Executing pipeline: %s\n", p.Name)
	if p.Description != "" {
		e.log("Description: %s\n", p.Description)
	}

	// For linear pipelines, execute as single command
	if p.IsLinear() {
		return e.executeLinearPipeline(p, log)
	}
	
	// Non-linear pipelines not yet supported
	log.Error("non-linear pipelines not yet supported")
	return fmt.Errorf("non-linear pipelines not yet supported")
}

// executeLinearPipeline executes a linear pipeline with tee logging
func (e *Executor) executeLinearPipeline(p *Pipeline, log logger.Logger) error {
	log.Debug("building shell command")
	// Create log directory for this execution
	timestamp := time.Now().Format("20060102_150405")
	logDir := filepath.Join(e.config.Logs.Directory, fmt.Sprintf("%s_%s", p.Name, timestamp))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Error("failed to create log directory", "error", err, "path", logDir)
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Build shell command with log directory
	shellCmd, err := BuildCommand(p, logDir)
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
	e.log("\nFull command: %s\n", shellCmd)
	e.log("Log directory: %s\n\n", logDir)
	
	// Create the command (1단계: 단순화)
	log.Debug("creating command")
	cmd := exec.Command("bash", "-c", shellCmd)
	
	// Create output limiters for screen display
	stdoutLimiter := &OutputLimiter{maxLines: 50, prefix: ""}
	stderrLimiter := &OutputLimiter{maxLines: 20, prefix: "[STDERR] "}
	
	// Set up output handling
	cmd.Stdout = stdoutLimiter
	cmd.Stderr = stderrLimiter
	
	// Execute the command with progress tracking
	log.Info("starting pipeline command")
	e.log("Starting execution...\n")
	startTime := time.Now()
	
	// Start a goroutine to show progress (only time elapsed, no complex monitoring)
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				e.log("\r⏱️  Running: %v", elapsed.Round(time.Second))
			}
		}
	}()
	
	cmdErr := cmd.Run()
	done <- true
	
	duration := time.Since(startTime)
	e.log("\n") // New line after progress
	
	// Show output summary
	totalStdoutLines := stdoutLimiter.currentLine
	if stdoutLimiter.truncated {
		e.log("[Output truncated: showing first %d of %d+ lines]\n", stdoutLimiter.maxLines, totalStdoutLines)
	}
	
	// Simple result logging (1단계: 단순 결과 로깅)
	if cmdErr != nil {
		log.Error("pipeline command failed", "error", cmdErr)
	} else {
		log.Info("pipeline command completed successfully")
	}
	
	status := "Success"
	if cmdErr != nil {
		status = fmt.Sprintf("Failed: %v", cmdErr)
	}
	
	log.Info("pipeline execution completed",
		"duration", duration,
		"status", status)
	
	// Create summary file
	if err := e.writeSummary(logDir, p, duration, status); err != nil {
		log.Error("failed to write summary", "error", err)
	}
	
	// User feedback
	e.log("\n%s\n", strings.Repeat("=", 50))
	e.log("Pipeline completed\n")
	e.log("• Duration: %v\n", duration)
	e.log("• Status: %s\n", status)
	e.log("• Full output: %s\n", filepath.Join(logDir, "pipeline.out"))
	e.log("• Summary: %s\n", filepath.Join(logDir, "summary.txt"))
	
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

// writeSummary creates a summary file for the pipeline execution
func (e *Executor) writeSummary(logDir string, p *Pipeline, duration time.Duration, status string) error {
	summaryPath := filepath.Join(logDir, "summary.txt")
	content := fmt.Sprintf(`Pipeline: %s
Description: %s
Executed: %s
Duration: %v
Status: %s
Log Directory: %s
`,
		p.Name,
		p.Description,
		time.Now().Format("2006-01-02 15:04:05"),
		duration,
		status,
		logDir)
	
	return os.WriteFile(summaryPath, []byte(content), 0644)
}

// CaptureOutput runs a pipeline and captures its output
func (e *Executor) CaptureOutput(p *Pipeline) (string, error) {
	// Ensure logger exists
	if e.logger == nil {
		e.logger = logger.Default()
	}
	
	log := e.logger.With("pipeline", p.Name, "mode", "capture")
	log.Debug("capturing pipeline output")
	
	// Build shell command (for capture, we don't need persistent logging)
	shellCmd, err := BuildCommand(p, "/tmp")
	if err != nil {
		log.Error("failed to build shell command", "error", err)
		return "", fmt.Errorf("failed to build shell command: %w", err)
	}
	
	// Remove tee for capture mode to get clean output
	if strings.Contains(shellCmd, " | tee ") {
		parts := strings.Split(shellCmd, " | tee ")
		shellCmd = parts[0]
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


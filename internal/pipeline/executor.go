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

// executeLinearPipeline executes a linear pipeline with tee logging
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
	
	// Create the command (1단계: 단순화)
	log.Debug("creating command")
	cmd := exec.Command("bash", "-c", shellCmd)
	
	// Simple direct output (고루틴 없이 직접 출력)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Execute the command (1단계: 단순 실행)
	log.Info("starting pipeline command")
	startTime := time.Now()
	
	cmdErr := cmd.Run()
	
	duration := time.Since(startTime)
	
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
	
	// Simple user feedback (1단계: 단순 사용자 피드백)
	e.log("\n%s\n", strings.Repeat("=", 50))
	e.log("Pipeline completed\n")
	e.log("• Duration: %v\n", duration)
	e.log("• Status: %s\n", status)
	e.log("• Debug log: /tmp/cli-pipe-debug.log\n")
	
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


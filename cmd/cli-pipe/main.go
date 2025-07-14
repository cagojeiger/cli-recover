package main

import (
	"fmt"
	"os"
	
	"github.com/cagojeiger/cli-pipe/internal/config"
	"github.com/cagojeiger/cli-pipe/internal/logger"
	"github.com/cagojeiger/cli-pipe/internal/pipeline"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	// Initialize logger early for all commands
	initializeLogger()
	
	// Check if we have at least a command
	if len(args) < 2 {
		printUsage()
		return 1
	}
	
	// Handle special cases before flag parsing
	command := args[1]
	switch command {
	case "version", "--version", "-v":
		printVersion()
		return 0
	case "help", "--help", "-h":
		printUsage()
		return 0
	case "init":
		return initConfigCmd()
	}
	
	// For 'run' command
	if command == "run" {
		// Get pipeline file
		if len(args) < 3 {
			logger.Error("missing pipeline file")
			fmt.Println("Usage: cli-pipe run <pipeline.yaml>")
			return 1
		}
		
		return runPipelineCmd(args[2])
	} else {
		logger.Error("unknown command", "command", command)
		printUsage()
		return 1
	}
}

func printVersion() {
	fmt.Printf("cli-pipe %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built:  %s\n", date)
}

func printUsage() {
	fmt.Println("cli-pipe - Unix command pipeline orchestrator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cli-pipe <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init       Initialize cli-pipe configuration")
	fmt.Println("  run        Run a pipeline from a YAML file")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cli-pipe init")
	fmt.Println("  cli-pipe run pipeline.yaml")
}

func runPipelineCmd(filename string) int {
	log := logger.With("command", "run", "file", filename)
	log.Info("parsing pipeline file")
	
	// Parse pipeline file
	p, err := pipeline.ParseFile(filename)
	if err != nil {
		log.Error("failed to parse pipeline file", "error", err)
		return 1
	}
	
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		return 1
	}
	
	// Create executor with config
	executor := pipeline.NewExecutor(cfg)
	
	// Execute pipeline
	log.Info("executing pipeline", "name", p.Name)
	if err := executor.Execute(p); err != nil {
		log.Error("pipeline execution failed", "error", err)
		return 1
	}
	
	log.Info("pipeline execution completed successfully")
	return 0
}

func initConfigCmd() int {
	log := logger.With("command", "init")
	log.Info("initializing cli-pipe configuration")
	
	// Create default config
	cfg := config.DefaultConfig()
	
	// Save config
	if err := cfg.Save(); err != nil {
		log.Error("failed to save config", "error", err)
		return 1
	}
	
	// Ensure log directory exists
	if err := cfg.EnsureLogDir(); err != nil {
		log.Error("failed to create log directory", "error", err)
		return 1
	}
	
	log.Info("cli-pipe configuration initialized",
		"config_dir", config.ConfigDir(),
		"config_file", config.ConfigPath(),
		"log_dir", cfg.Logs.Directory)
	
	// Still print user-friendly output
	fmt.Printf("Initialized cli-pipe configuration at %s\n", config.ConfigDir())
	fmt.Printf("Configuration file: %s\n", config.ConfigPath())
	fmt.Printf("Log directory: %s\n", cfg.Logs.Directory)
	
	return 0
}

// initializeLogger sets up the logger based on configuration
func initializeLogger() {
	// Try to load config to get logger settings
	cfg, err := config.Load()
	if err != nil {
		// If config doesn't exist, use defaults
		return
	}
	
	// If logger config exists, initialize with it
	if cfg.Logger != nil {
		if log, err := logger.New(cfg.Logger); err == nil {
			logger.SetDefault(log)
		}
	}
}
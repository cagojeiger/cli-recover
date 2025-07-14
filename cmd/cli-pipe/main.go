package main

import (
	"fmt"
	"os"
	
	"github.com/cagojeiger/cli-pipe/internal/config"
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
			fmt.Fprintf(os.Stderr, "Error: missing pipeline file\n")
			fmt.Println("Usage: cli-pipe run <pipeline.yaml>")
			return 1
		}
		
		return runPipelineCmd(args[2])
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
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
	// Parse pipeline file
	p, err := pipeline.ParseFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pipeline file: %v\n", err)
		return 1
	}
	
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}
	
	// Create executor with config
	executor := pipeline.NewExecutor(cfg)
	
	// Execute pipeline
	if err := executor.Execute(p); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pipeline: %v\n", err)
		return 1
	}
	
	return 0
}

func initConfigCmd() int {
	// Create default config
	cfg := config.DefaultConfig()
	
	// Save config
	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return 1
	}
	
	// Ensure log directory exists
	if err := cfg.EnsureLogDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log directory: %v\n", err)
		return 1
	}
	
	fmt.Printf("Initialized cli-pipe configuration at %s\n", config.ConfigDir())
	fmt.Printf("Configuration file: %s\n", config.ConfigPath())
	fmt.Printf("Log directory: %s\n", cfg.Logs.Directory)
	
	return 0
}
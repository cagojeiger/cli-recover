package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/cagojeiger/cli-recover/internal/application/usecase"
	"github.com/cagojeiger/cli-recover/internal/domain/service"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/persistence"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Command-line options
var (
	strategyFlag = flag.String("strategy", "auto", "Execution strategy: auto, shell-pipe, go-stream")
	logDirFlag   = flag.String("log-dir", "", "Directory for logging (shell strategy only)")
)

func main() {
	// Check if we have at least a command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	// Handle special cases before flag parsing
	command := os.Args[1]
	switch command {
	case "version", "--version", "-v":
		printVersion()
		return
	case "help", "--help", "-h":
		printUsage()
		return
	}
	
	// For 'run' command, parse flags
	if command == "run" {
		// Create new flag set for run command
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		strategy := runCmd.String("strategy", "auto", "Execution strategy: auto, shell-pipe, go-stream")
		logDir := runCmd.String("log-dir", "", "Directory for logging (shell strategy only)")
		
		// Parse from os.Args[2:] to skip program name and "run"
		runCmd.Parse(os.Args[2:])
		
		// Get remaining args after flags
		args := runCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: missing pipeline file\n")
			fmt.Println("Usage: cli-pipe run [options] <pipeline.yaml>")
			os.Exit(1)
		}
		
		// Set global flags for runPipeline
		*strategyFlag = *strategy
		*logDirFlag = *logDir
		
		runPipeline(args[0])
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
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
	fmt.Println("  run        Run a pipeline from a YAML file")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Options for 'run' command:")
	fmt.Println("  --strategy string   Execution strategy: auto, shell-pipe, go-stream (default \"auto\")")
	fmt.Println("  --log-dir string    Directory for logging (shell strategy only)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cli-pipe run pipeline.yaml")
	fmt.Println("  cli-pipe run --strategy shell-pipe --log-dir ./logs pipeline.yaml")
}

func runPipeline(filename string) {
	// Create dependencies
	parser := persistence.NewYAMLParser()
	streamManager := service.NewStreamManager()
	stepExecutor := usecase.NewExecuteStep(streamManager)
	pipelineExecutor := usecase.NewExecutePipeline(stepExecutor, streamManager)
	
	// Set log output to stdout
	pipelineExecutor.SetLogWriter(os.Stdout)
	
	// Parse pipeline file
	config, err := parser.ParseFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pipeline file: %v\n", err)
		os.Exit(1)
	}
	
	// Convert to domain model
	pipeline, err := parser.ToPipeline(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pipeline: %v\n", err)
		os.Exit(1)
	}
	
	// Prepare execution options
	options := usecase.ExecuteOptions{
		UseStrategy: true, // Always use strategy pattern
	}
	
	// Handle strategy flag
	switch *strategyFlag {
	case "shell-pipe":
		options.ForceStrategy = "shell-pipe"
	case "go-stream":
		options.ForceStrategy = "go-stream"
	case "auto":
		// Let the system determine the best strategy
	default:
		fmt.Fprintf(os.Stderr, "Invalid strategy: %s\n", *strategyFlag)
		os.Exit(1)
	}
	
	// Handle log directory
	if *logDirFlag != "" {
		// Expand ~ to home directory
		if (*logDirFlag)[0] == '~' {
			home, err := os.UserHomeDir()
			if err == nil {
				*logDirFlag = filepath.Join(home, (*logDirFlag)[1:])
			}
		}
		options.LogDir = *logDirFlag
	}
	
	// Execute pipeline with options
	if err := pipelineExecutor.ExecuteWithOptions(pipeline, options); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pipeline: %v\n", err)
		os.Exit(1)
	}
}
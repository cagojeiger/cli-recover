package main

import (
	"fmt"
	"os"
	
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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version", "--version", "-v":
		printVersion()
	case "help", "--help", "-h":
		printUsage()
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: missing pipeline file\n")
			fmt.Println("Usage: cli-pipe run <pipeline.yaml>")
			os.Exit(1)
		}
		runPipeline(os.Args[2])
	default:
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
	fmt.Println("Use \"cli-pipe <command> --help\" for more information about a command.")
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
	
	// Execute pipeline
	if err := pipelineExecutor.Execute(pipeline); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pipeline: %v\n", err)
		os.Exit(1)
	}
}
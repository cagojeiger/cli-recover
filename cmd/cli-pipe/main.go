package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/cagojeiger/cli-pipe/internal/pipeline"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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
		logDir := runCmd.String("log-dir", "", "Directory for logging")
		
		// Parse from os.Args[2:] to skip program name and "run"
		if err := runCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}
		
		// Get remaining args after flags
		args := runCmd.Args()
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Error: missing pipeline file\n")
			fmt.Println("Usage: cli-pipe run [options] <pipeline.yaml>")
			os.Exit(1)
		}
		
		runPipeline(args[0], *logDir)
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
	fmt.Println("  --log-dir string    Directory for logging")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cli-pipe run pipeline.yaml")
	fmt.Println("  cli-pipe run --log-dir ./logs pipeline.yaml")
}

func runPipeline(filename string, logDir string) {
	// Parse pipeline file
	p, err := pipeline.ParseFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pipeline file: %v\n", err)
		os.Exit(1)
	}
	
	// Create executor with options
	opts := []pipeline.Option{
		pipeline.WithLogWriter(os.Stdout),
	}
	
	// Handle log directory
	if logDir != "" {
		// Expand ~ to home directory
		if logDir[0] == '~' {
			home, err := os.UserHomeDir()
			if err == nil {
				logDir = filepath.Join(home, logDir[1:])
			}
		}
		opts = append(opts, pipeline.WithLogDir(logDir))
	}
	
	executor := pipeline.NewExecutor(opts...)
	
	// Execute pipeline
	if err := executor.Execute(p); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pipeline: %v\n", err)
		os.Exit(1)
	}
}
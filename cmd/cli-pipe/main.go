package main

import (
	"fmt"
	"os"
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
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Use \"cli-pipe <command> --help\" for more information about a command.")
}
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
	"github.com/cagojeiger/cli-recover/internal/runner"
	"github.com/cagojeiger/cli-recover/internal/tui"
)

// version will be set by ldflags during build
var version = "dev"

func main() {
	var rootCmd = &cobra.Command{
		Use:     "cli-recover",
		Short:   "Kubernetes integrated backup and restore tool",
		Long:    `CLI-Recover provides backup and restore capabilities for Kubernetes environments including pod filesystems, databases, and object storage.`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize logger from flags
			logLevel, _ := cmd.Flags().GetString("log-level")
			logFile, _ := cmd.Flags().GetString("log-file")
			logFormat, _ := cmd.Flags().GetString("log-format")
			
			cfg := logger.DefaultConfig()
			if logLevel != "" {
				cfg.Level = logLevel
			}
			if logFile != "" {
				cfg.Output = "both"
				cfg.FilePath = logFile
			}
			if logFormat == "json" {
				cfg.JSONFormat = true
			}
			
			// Set logger level based on debug flag
			debug, _ := cmd.Flags().GetBool("debug")
			if debug && cfg.Level == "info" {
				cfg.Level = "debug"
			}
			
			// Initialize global logger
			if err := logger.InitializeFromConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Get debug flag for TUI mode
			debug, _ := cmd.Flags().GetBool("debug")
			
			// If no args, start TUI
			runner := runner.NewRunner()
			tui.SetVersion(version)
			tui.SetDebug(debug)
			model := tui.InitialModel(runner)
			
			if debug {
				fmt.Printf("Debug: Starting TUI mode\n")
			}
			
			p := tea.NewProgram(model, tea.WithAltScreen())
			
			// Set program reference for message passing
			model.SetProgram(p)
			
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Customize version template
	rootCmd.SetVersionTemplate("cli-recover version {{.Version}}\n")
	
	// Add global flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-file", "", "Log file path (logs to console if not specified)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (text, json)")

	// Add new provider-based backup command (recommended)
	rootCmd.AddCommand(newBackupCommand())
	
	// Add new provider-based restore command
	rootCmd.AddCommand(newRestoreCommand())
	
	// Add list command
	rootCmd.AddCommand(newListCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
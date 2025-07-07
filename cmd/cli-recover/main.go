package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"

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
	
	// Add global debug flag
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")

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
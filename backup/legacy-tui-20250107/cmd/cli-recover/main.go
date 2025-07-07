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

	// Add backup command with subcommands
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup resources from Kubernetes",
		Long: `Backup various types of resources from Kubernetes pods.
		
Available backup types:
  filesystem - Backup files and directories from pod filesystem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}
	
	// Add subcommands for different backup types
	backupCmd.AddCommand(newFilesystemBackupCmd())
	
	rootCmd.AddCommand(backupCmd)
	
	// Legacy backup command for backward compatibility
	var legacyBackupCmd = &cobra.Command{
		Use:    "backup-legacy [pod] [path]",
		Hidden: true, // Hide from help
		Short:  "Legacy backup command (deprecated)",
		Args:   cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("This command is deprecated. Please use 'cli-recover backup filesystem' instead.\n")
			return runFilesystemBackup(cmd, args)
		},
	}
	
	// Add flags to legacy command
	legacyBackupCmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
	legacyBackupCmd.Flags().StringP("compression", "c", "gzip", "Compression type (gzip, bzip2, xz, none)")
	legacyBackupCmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude patterns (can be used multiple times)")
	legacyBackupCmd.Flags().BoolP("exclude-vcs", "", false, "Exclude version control systems (.git, .svn, etc.)")
	legacyBackupCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	legacyBackupCmd.Flags().BoolP("totals", "t", false, "Show transfer totals")
	legacyBackupCmd.Flags().BoolP("preserve-perms", "p", false, "Preserve file permissions")
	legacyBackupCmd.Flags().StringP("container", "", "", "Container name (for multi-container pods)")
	legacyBackupCmd.Flags().StringP("output", "o", "", "Output file path (auto-generated if not specified)")
	legacyBackupCmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
	
	rootCmd.AddCommand(legacyBackupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}


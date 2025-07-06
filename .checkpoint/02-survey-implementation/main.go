package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/cagojeiger/cli-restore/internal/backup"
	"github.com/cagojeiger/cli-restore/internal/tui"
)

// version will be set by ldflags during build
var version = "dev"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cli-restore",
		Short: "Kubernetes Pod backup utility",
		Long: `CLI-Restore is a tool for backing up files and directories from Kubernetes pods.
It creates tar archives with optional splitting for large files.`,
		Version: version,
	}

	// Customize version template to show only version string
	rootCmd.SetVersionTemplate("cli-restore version {{.Version}}\n")

	// Add TUI command
	var tuiCmd = &cobra.Command{
		Use:   "tui",
		Short: "Interactive backup configuration",
		Long:  `Launch interactive TUI to configure and execute backup operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("CLI Restore - Interactive Mode")
			fmt.Println(strings.Repeat("=", 30))
			fmt.Println()

			options, err := tui.RunInteractiveBackup()
			if err != nil {
				return fmt.Errorf("interactive backup failed: %w", err)
			}

			// Execute backup
			backupOpts := &backup.Options{
				Pod:       options.Pod,
				Namespace: options.Namespace,
				Path:      options.Path,
				SplitSize: options.SplitSize,
				Output:    "./backup",
			}

			if err := backup.ValidateOptions(backupOpts); err != nil {
				return fmt.Errorf("invalid backup options: %w", err)
			}

			return backup.Execute(backupOpts)
		},
	}

	// Add backup command (placeholder for now)
	var backupCmd = &cobra.Command{
		Use:   "backup <pod> <path>",
		Short: "Backup pod files directly",
		Long:  `Backup files from a Kubernetes pod directly using command line arguments.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pod := args[0]
			path := args[1]
			namespace, _ := cmd.Flags().GetString("namespace")
			splitSize, _ := cmd.Flags().GetString("split-size")
			output, _ := cmd.Flags().GetString("output")

			backupOpts := &backup.Options{
				Pod:       pod,
				Namespace: namespace,
				Path:      path,
				SplitSize: splitSize,
				Output:    output,
			}

			if err := backup.ValidateOptions(backupOpts); err != nil {
				return fmt.Errorf("invalid backup options: %w", err)
			}

			return backup.Execute(backupOpts)
		},
	}

	// Add flags to backup command
	backupCmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
	backupCmd.Flags().StringP("split-size", "s", "1G", "Split size for large files")
	backupCmd.Flags().StringP("output", "o", "./backup", "Output directory")

	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(backupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
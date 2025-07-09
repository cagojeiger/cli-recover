package main

import (
	"github.com/spf13/cobra"
)

// newBackupCommand creates the new backup command structure
func newBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
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
	cmd.AddCommand(newProviderBackupCmd("filesystem"))

	return cmd
}

// newProviderBackupCmd creates a backup command for a specific provider
func newProviderBackupCmd(providerName string) *cobra.Command {
	var cmd *cobra.Command

	switch providerName {
	case "filesystem":
		cmd = &cobra.Command{
			Use:   "filesystem [pod] [path]",
			Short: "Backup pod filesystem",
			Long:  `Backup files and directories from a pod's filesystem using tar`,
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return executeBackup(providerName, cmd, args)
			},
		}

		// Add filesystem-specific flags
		cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
		cmd.Flags().StringP("compression", "c", "none", "Compression type (none=.tar, gzip=.tar.gz)")
		cmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude patterns (can be used multiple times)")
		cmd.Flags().BoolP("exclude-vcs", "", false, "Exclude version control systems (.git, .svn, etc.)")
		cmd.Flags().BoolP("verbose", "v", false, "Verbose output")
		cmd.Flags().BoolP("totals", "T", false, "Show transfer totals")
		cmd.Flags().BoolP("preserve-perms", "p", false, "Preserve file permissions")
		cmd.Flags().StringP("container", "", "", "Container name (for multi-container pods)")
		cmd.Flags().StringP("output", "o", "", "Output file path (auto-generated if not specified)")
		cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
		cmd.Flags().String("log-dir", "", "Directory to store logs (default: ~/.cli-recover/logs)")
	}

	return cmd
}

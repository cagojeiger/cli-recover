package main

import (
	"github.com/cagojeiger/cli-recover/internal/domain/flags"
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

		// Add filesystem-specific flags using registry
		cmd.Flags().StringP(flags.LongNames.Namespace, flags.Registry.Namespace, "default", "Kubernetes namespace")
		cmd.Flags().StringP(flags.LongNames.Compression, flags.Registry.Compression, "none", "Compression type (none=.tar, gzip=.tar.gz)")
		cmd.Flags().StringSliceP(flags.LongNames.Exclude, flags.Registry.Exclude, []string{}, "Exclude patterns (can be used multiple times)")
		cmd.Flags().BoolP("exclude-vcs", "", false, "Exclude version control systems (.git, .svn, etc.)")
		cmd.Flags().BoolP(flags.LongNames.Verbose, flags.Registry.Verbose, false, "Verbose output")
		cmd.Flags().BoolP(flags.LongNames.Totals, flags.Registry.Totals, false, "Show transfer totals")
		cmd.Flags().BoolP(flags.LongNames.PreservePerms, flags.Registry.PreservePerms, false, "Preserve file permissions")
		cmd.Flags().StringP(flags.LongNames.Container, flags.Registry.Container, "", "Container name (for multi-container pods)")
		cmd.Flags().StringP(flags.LongNames.Output, flags.Registry.Output, "", "Output file path (auto-generated if not specified)")
		cmd.Flags().BoolP(flags.LongNames.DryRun, flags.Registry.DryRun, false, "Show what would be executed without running")
		cmd.Flags().String("log-dir", "", "Directory to store logs (default: ~/.cli-recover/logs)")
	}

	return cmd
}

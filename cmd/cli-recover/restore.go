package main

import (
	"github.com/cagojeiger/cli-recover/internal/domain/flags"
	"github.com/spf13/cobra"
)

// newRestoreCommand creates the new restore command structure
func newRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore resources to Kubernetes",
		Long: `Restore various types of backups to Kubernetes pods.

Available restore types:
  filesystem - Restore files and directories to pod filesystem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Add subcommands for different restore types
	cmd.AddCommand(newProviderRestoreCmd("filesystem"))

	return cmd
}

// newProviderRestoreCmd creates a restore command for a specific provider
func newProviderRestoreCmd(providerName string) *cobra.Command {
	var cmd *cobra.Command

	switch providerName {
	case "filesystem":
		cmd = &cobra.Command{
			Use:   "filesystem [pod] [backup-file]",
			Short: "Restore pod filesystem from backup",
			Long:  `Restore files and directories to a pod's filesystem from a tar backup`,
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return executeRestore(providerName, cmd, args)
			},
		}

		// Add filesystem-specific flags using registry
		cmd.Flags().StringP(flags.LongNames.Namespace, flags.Registry.Namespace, "default", "Kubernetes namespace")
		cmd.Flags().StringP(flags.LongNames.TargetPath, flags.Registry.TargetPath, "/", "Target restore path in the pod")
		cmd.Flags().BoolP(flags.LongNames.Force, flags.Registry.Force, false, "Force overwrite existing files")
		cmd.Flags().BoolP(flags.LongNames.PreservePerms, flags.Registry.PreservePerms, false, "Preserve file permissions")
		cmd.Flags().StringSliceP(flags.LongNames.SkipPaths, flags.Registry.SkipPaths, []string{}, "Paths to skip during restore")
		cmd.Flags().StringP(flags.LongNames.Container, flags.Registry.Container, "", "Container name (for multi-container pods)")
		cmd.Flags().BoolP(flags.LongNames.Verbose, flags.Registry.Verbose, false, "Verbose output")
		cmd.Flags().BoolP(flags.LongNames.DryRun, flags.Registry.DryRun, false, "Show what would be executed without running")
	}

	return cmd
}

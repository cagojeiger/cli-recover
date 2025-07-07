package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/cagojeiger/cli-recover/internal/application/adapters"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// newRestoreCommand creates the new restore command structure
func newRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore resources to Kubernetes",
		Long: `Restore various types of backups to Kubernetes pods.

Available restore types:
  filesystem - Restore files and directories to pod filesystem
  minio      - Restore MinIO buckets (coming soon)
  mongodb    - Restore MongoDB databases (coming soon)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}
	
	// Add subcommands for different restore types
	cmd.AddCommand(newProviderRestoreCmd("filesystem"))
	// TODO: Add when providers are ready
	// cmd.AddCommand(newProviderRestoreCmd("minio"))
	// cmd.AddCommand(newProviderRestoreCmd("mongodb"))
	
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
				adapter := adapters.NewRestoreAdapter(restore.GlobalRegistry)
				return adapter.ExecuteRestore(providerName, cmd, args)
			},
		}
		
		// Add filesystem-specific flags
		cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
		cmd.Flags().StringP("target-path", "t", "/", "Target restore path in the pod")
		cmd.Flags().BoolP("overwrite", "o", false, "Overwrite existing files")
		cmd.Flags().BoolP("preserve-perms", "p", false, "Preserve file permissions")
		cmd.Flags().StringSliceP("skip-paths", "s", []string{}, "Paths to skip during restore")
		cmd.Flags().StringP("container", "c", "", "Container name (for multi-container pods)")
		cmd.Flags().BoolP("verbose", "v", false, "Verbose output")
		cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
		
	case "minio":
		cmd = &cobra.Command{
			Use:   "minio [bucket] [backup-dir]",
			Short: "Restore MinIO bucket from backup",
			Long:  `Restore MinIO bucket contents from backup`,
			Args:  cobra.MinimumNArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("MinIO restore provider not yet implemented")
			},
		}
		
		// Add MinIO-specific flags
		cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
		cmd.Flags().StringP("service", "s", "", "MinIO service name")
		cmd.Flags().StringP("access-key", "", "", "MinIO access key")
		cmd.Flags().StringP("secret-key", "", "", "MinIO secret key")
		cmd.Flags().BoolP("overwrite", "o", false, "Overwrite existing objects")
		cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
		
	case "mongodb":
		cmd = &cobra.Command{
			Use:   "mongodb [database] [backup-file]",
			Short: "Restore MongoDB database from backup",
			Long:  `Restore MongoDB database using mongorestore`,
			Args:  cobra.MinimumNArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("MongoDB restore provider not yet implemented")
			},
		}
		
		// Add MongoDB-specific flags
		cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
		cmd.Flags().StringP("pod", "p", "", "MongoDB pod name")
		cmd.Flags().StringP("uri", "u", "", "MongoDB connection URI")
		cmd.Flags().BoolP("drop", "d", false, "Drop collections before restore")
		cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
	}
	
	return cmd
}
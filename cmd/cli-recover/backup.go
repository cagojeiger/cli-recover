package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/cagojeiger/cli-recover/internal/application/adapters"
)

// newBackupCommand creates the new backup command structure
func newBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup resources from Kubernetes",
		Long: `Backup various types of resources from Kubernetes pods.

Available backup types:
  filesystem - Backup files and directories from pod filesystem
  minio      - Backup MinIO buckets (coming soon)
  mongodb    - Backup MongoDB databases (coming soon)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}
	
	// Add subcommands for different backup types
	cmd.AddCommand(newProviderBackupCmd("filesystem"))
	// TODO: Add when providers are ready
	// cmd.AddCommand(newProviderBackupCmd("minio"))
	// cmd.AddCommand(newProviderBackupCmd("mongodb"))
	
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
				adapter := adapters.NewBackupAdapter()
				return adapter.ExecuteBackup(providerName, cmd, args)
			},
		}
		
		// Add filesystem-specific flags
		cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
		cmd.Flags().StringP("compression", "c", "gzip", "Compression type (gzip, bzip2, xz, none)")
		cmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude patterns (can be used multiple times)")
		cmd.Flags().BoolP("exclude-vcs", "", false, "Exclude version control systems (.git, .svn, etc.)")
		cmd.Flags().BoolP("verbose", "v", false, "Verbose output")
		cmd.Flags().BoolP("totals", "t", false, "Show transfer totals")
		cmd.Flags().BoolP("preserve-perms", "p", false, "Preserve file permissions")
		cmd.Flags().StringP("container", "", "", "Container name (for multi-container pods)")
		cmd.Flags().StringP("output", "o", "", "Output file path (auto-generated if not specified)")
		cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
		
	case "minio":
		cmd = &cobra.Command{
			Use:   "minio [bucket]",
			Short: "Backup MinIO bucket",
			Long:  `Backup MinIO bucket contents`,
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("MinIO backup provider not yet implemented")
			},
		}
		
		// Add MinIO-specific flags
		cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
		cmd.Flags().StringP("service", "s", "", "MinIO service name")
		cmd.Flags().StringP("access-key", "", "", "MinIO access key")
		cmd.Flags().StringP("secret-key", "", "", "MinIO secret key")
		cmd.Flags().StringP("output", "o", "", "Output directory")
		cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
		
	case "mongodb":
		cmd = &cobra.Command{
			Use:   "mongodb [database]",
			Short: "Backup MongoDB database",
			Long:  `Backup MongoDB database using mongodump`,
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("MongoDB backup provider not yet implemented")
			},
		}
		
		// Add MongoDB-specific flags
		cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
		cmd.Flags().StringP("pod", "p", "", "MongoDB pod name")
		cmd.Flags().StringP("uri", "u", "", "MongoDB connection URI")
		cmd.Flags().StringSliceP("collections", "c", []string{}, "Specific collections to backup")
		cmd.Flags().StringP("output", "o", "", "Output file path")
		cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
	}
	
	return cmd
}
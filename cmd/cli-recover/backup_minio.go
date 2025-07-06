package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/runner"
)

// newMinioBackupCmd creates the MinIO backup command
func newMinioBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minio [pod] [bucket/path]",
		Short: "Backup MinIO object storage",
		Long:  `Backup objects from MinIO object storage`,
		Args:  cobra.ExactArgs(2),
		RunE:  runMinioBackup,
	}
	
	// Add MinIO-specific flags
	cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringP("format", "f", "tar", "Output format (tar, zip)")
	cmd.Flags().BoolP("recursive", "r", true, "Recursive backup")
	cmd.Flags().StringP("output", "o", "", "Output file path")
	cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
	cmd.Flags().StringP("endpoint", "", "", "MinIO endpoint URL (if not default)")
	cmd.Flags().StringP("access-key", "", "", "MinIO access key (if not from env)")
	cmd.Flags().StringP("secret-key", "", "", "MinIO secret key (if not from env)")
	cmd.Flags().StringP("container", "", "", "Container name (for multi-container pods)")
	
	return cmd
}

func runMinioBackup(cmd *cobra.Command, args []string) error {
	pod := args[0]
	source := args[1]
	
	namespace, _ := cmd.Flags().GetString("namespace")
	format, _ := cmd.Flags().GetString("format")
	recursive, _ := cmd.Flags().GetBool("recursive")
	outputFile, _ := cmd.Flags().GetString("output")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	debug, _ := cmd.Flags().GetBool("debug")
	
	endpoint, _ := cmd.Flags().GetString("endpoint")
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")
	container, _ := cmd.Flags().GetString("container")
	
	if debug {
		fmt.Printf("Debug: MinIO backup\n")
		fmt.Printf("  pod: %s\n", pod)
		fmt.Printf("  source: %s\n", source)
		fmt.Printf("  namespace: %s\n", namespace)
		fmt.Printf("  format: %s\n", format)
		fmt.Printf("  recursive: %v\n", recursive)
		fmt.Printf("  output: %s\n", outputFile)
		fmt.Printf("  dry-run: %v\n", dryRun)
		fmt.Printf("  endpoint: %s\n", endpoint)
		fmt.Printf("  container: %s\n", container)
	}
	
	runner := runner.NewRunner()
	
	// Verify pod exists
	if debug {
		fmt.Printf("Debug: Verifying pod exists in namespace %s\n", namespace)
	}
	pods, err := kubernetes.GetPods(runner, namespace)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}
	
	found := false
	for _, p := range pods {
		if p.Name == pod {
			found = true
			if debug {
				fmt.Printf("Debug: Found pod %s (status: %s, ready: %s)\n", p.Name, p.Status, p.Ready)
			}
			break
		}
	}
	
	if !found {
		return fmt.Errorf("pod %s not found in namespace %s", pod, namespace)
	}
	
	// Generate output filename if not provided
	if outputFile == "" {
		// Extract bucket name from source
		bucketName := strings.Split(source, "/")[0]
		if format == "zip" {
			outputFile = fmt.Sprintf("minio-backup-%s-%s-%s.zip", namespace, pod, bucketName)
		} else {
			outputFile = fmt.Sprintf("minio-backup-%s-%s-%s.tar.gz", namespace, pod, bucketName)
		}
	}
	
	// Build mc mirror command
	command := buildMinioBackupCommand(pod, namespace, source, container, recursive, endpoint, accessKey, secretKey)
	
	if dryRun {
		fmt.Printf("Dry run - would execute:\n")
		fmt.Printf("1. Setup MinIO alias: %s\n", command.setupCmd)
		fmt.Printf("2. Mirror data: %s\n", command.mirrorCmd)
		fmt.Printf("3. Create archive: %s\n", command.archiveCmd)
		fmt.Printf("Output would be saved to: %s\n", outputFile)
		return nil
	}
	
	fmt.Printf("Starting MinIO backup...\n")
	
	// Execute backup
	return executeMinioBackup(runner, command, outputFile, pod, namespace, source, format, debug)
}

// minioCommands holds the commands for MinIO backup
type minioCommands struct {
	setupCmd   string
	mirrorCmd  string
	archiveCmd string
}

// buildMinioBackupCommand creates the mc commands for MinIO backup
func buildMinioBackupCommand(pod, namespace, source, container string, recursive bool, endpoint, accessKey, secretKey string) minioCommands {
	var cmds minioCommands
	
	// Build kubectl exec prefix
	kubectlPrefix := fmt.Sprintf("kubectl exec -n %s %s", namespace, pod)
	if container != "" {
		kubectlPrefix += fmt.Sprintf(" -c %s", container)
	}
	kubectlPrefix += " --"
	
	// Setup mc alias
	minioEndpoint := endpoint
	if minioEndpoint == "" {
		minioEndpoint = "http://localhost:9000"
	}
	
	cmds.setupCmd = fmt.Sprintf("%s mc alias set backup %s", kubectlPrefix, minioEndpoint)
	if accessKey != "" && secretKey != "" {
		cmds.setupCmd += fmt.Sprintf(" %s %s", accessKey, secretKey)
	}
	
	// Build mirror command
	cmds.mirrorCmd = fmt.Sprintf("%s mc mirror", kubectlPrefix)
	if !recursive {
		cmds.mirrorCmd += " --no-recursion"
	}
	cmds.mirrorCmd += fmt.Sprintf(" backup/%s /tmp/minio-backup/", source)
	
	// Build archive command
	cmds.archiveCmd = fmt.Sprintf("%s tar -czf - -C /tmp/minio-backup .", kubectlPrefix)
	
	return cmds
}

// executeMinioBackup performs the MinIO backup operation
func executeMinioBackup(runner runner.Runner, cmds minioCommands, outputFile, pod, namespace, source, format string, debug bool) error {
	if debug {
		fmt.Printf("Debug: Starting MinIO backup execution\n")
	}
	
	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputFile, err)
	}
	defer outFile.Close()
	
	// Step 1: Setup mc alias
	if debug {
		fmt.Printf("Debug: Setting up MinIO client alias\n")
	}
	parts := strings.Fields(cmds.setupCmd)
	if _, err := runner.Run(parts[0], parts[1:]...); err != nil {
		return fmt.Errorf("failed to setup mc alias: %w", err)
	}
	
	// Step 2: Create temp directory
	mkdirCmd := fmt.Sprintf("kubectl exec -n %s %s -- mkdir -p /tmp/minio-backup", namespace, pod)
	parts = strings.Fields(mkdirCmd)
	if _, err := runner.Run(parts[0], parts[1:]...); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Step 3: Mirror data
	if debug {
		fmt.Printf("Debug: Mirroring MinIO data to temp directory\n")
	}
	fmt.Printf("Downloading MinIO objects from %s...\n", source)
	parts = strings.Fields(cmds.mirrorCmd)
	if _, err := runner.Run(parts[0], parts[1:]...); err != nil {
		return fmt.Errorf("failed to mirror MinIO data: %w", err)
	}
	
	// Step 4: Create archive
	if debug {
		fmt.Printf("Debug: Creating archive\n")
	}
	fmt.Printf("Creating backup archive...\n")
	parts = strings.Fields(cmds.archiveCmd)
	output, err := runner.Run(parts[0], parts[1:]...)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	
	// Write output to file
	_, err = outFile.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}
	
	// Step 5: Cleanup temp directory
	cleanupCmd := fmt.Sprintf("kubectl exec -n %s %s -- rm -rf /tmp/minio-backup", namespace, pod)
	parts = strings.Fields(cleanupCmd)
	if _, err := runner.Run(parts[0], parts[1:]...); err != nil {
		// Non-fatal error
		if debug {
			fmt.Printf("Debug: Warning - failed to cleanup temp directory: %v\n", err)
		}
	}
	
	// Get file info for success message
	fileInfo, err := outFile.Stat()
	if err != nil {
		fmt.Printf("MinIO backup completed successfully: %s\n", outputFile)
	} else {
		fmt.Printf("MinIO backup completed successfully: %s (%d bytes)\n", outputFile, fileInfo.Size())
	}
	
	if debug {
		fmt.Printf("Debug: MinIO backup execution completed\n")
	}
	
	return nil
}
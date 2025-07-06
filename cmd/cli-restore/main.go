package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cagojeiger/cli-restore/internal/kubernetes"
	"github.com/cagojeiger/cli-restore/internal/runner"
	"github.com/cagojeiger/cli-restore/internal/tui"
)

// version will be set by ldflags during build
var version = "dev"

func main() {
	var rootCmd = &cobra.Command{
		Use:     "cli-restore",
		Short:   "Kubernetes integrated backup and restore tool",
		Long:    `CLI-Restore provides backup and restore capabilities for Kubernetes environments including pod filesystems, databases, and object storage.`,
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
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Customize version template
	rootCmd.SetVersionTemplate("cli-restore version {{.Version}}\n")
	
	// Add global debug flag
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")

	// Add backup command for CLI mode
	var backupCmd = &cobra.Command{
		Use:   "backup [pod] [path]",
		Short: "Backup pod files directly",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pod := args[0]
			path := args[1]
			
			// Get all flags
			namespace, _ := cmd.Flags().GetString("namespace")
			compression, _ := cmd.Flags().GetString("compression")
			excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")
			excludeVCS, _ := cmd.Flags().GetBool("exclude-vcs")
			verbose, _ := cmd.Flags().GetBool("verbose")
			showTotals, _ := cmd.Flags().GetBool("totals")
			preservePerms, _ := cmd.Flags().GetBool("preserve-perms")
			container, _ := cmd.Flags().GetString("container")
			outputFile, _ := cmd.Flags().GetString("output")
			debug, _ := cmd.Flags().GetBool("debug")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			
			if debug {
				fmt.Printf("Debug: CLI flags parsed\n")
				fmt.Printf("  pod: %s\n", pod)
				fmt.Printf("  path: %s\n", path)
				fmt.Printf("  namespace: %s\n", namespace)
				fmt.Printf("  compression: %s\n", compression)
				fmt.Printf("  exclude-patterns: %v\n", excludePatterns)
				fmt.Printf("  exclude-vcs: %v\n", excludeVCS)
				fmt.Printf("  verbose: %v\n", verbose)
				fmt.Printf("  container: %s\n", container)
				fmt.Printf("  output: %s\n", outputFile)
				fmt.Printf("  dry-run: %v\n", dryRun)
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
			
			// Build backup options from flags
			options := kubernetes.BackupOptions{
				CompressionType: compression,
				ExcludePatterns: excludePatterns,
				ExcludeVCS:      excludeVCS,
				Verbose:         verbose,
				ShowTotals:      showTotals,
				PreservePerms:   preservePerms,
				Container:       container,
				OutputFile:      outputFile,
			}
			
			if debug {
				fmt.Printf("Debug: Built backup options: %+v\n", options)
			}
			
			// Generate command
			command := kubernetes.GenerateBackupCommand(pod, namespace, path, options)
			
			if dryRun {
				fmt.Printf("Dry run - would execute: %s\n", command)
				if outputFile != "" {
					fmt.Printf("Output would be saved to: %s\n", outputFile)
				} else {
					fmt.Printf("Output would be saved to: backup-%s-%s-%s.tar.gz\n", namespace, pod, 
						generatePathSuffix(path))
				}
				return nil
			}
			
			fmt.Printf("Executing: %s\n", command)
			
			// Execute actual backup
			return executeBackup(runner, command, outputFile, pod, namespace, path, debug)
		},
	}
	
	// Add comprehensive flags
	backupCmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
	backupCmd.Flags().StringP("compression", "c", "gzip", "Compression type (gzip, bzip2, xz, none)")
	backupCmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude patterns (can be used multiple times)")
	backupCmd.Flags().BoolP("exclude-vcs", "", false, "Exclude version control systems (.git, .svn, etc.)")
	backupCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	backupCmd.Flags().BoolP("totals", "t", false, "Show transfer totals")
	backupCmd.Flags().BoolP("preserve-perms", "p", false, "Preserve file permissions")
	backupCmd.Flags().StringP("container", "", "", "Container name (for multi-container pods)")
	backupCmd.Flags().StringP("output", "o", "", "Output file path (auto-generated if not specified)")
	backupCmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
	
	rootCmd.AddCommand(backupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generatePathSuffix creates a safe filename suffix from a path
func generatePathSuffix(path string) string {
	if path == "/" {
		return "root"
	}
	// Remove leading slash and replace slashes with dashes
	suffix := strings.TrimPrefix(path, "/")
	suffix = strings.ReplaceAll(suffix, "/", "-")
	suffix = strings.ReplaceAll(suffix, " ", "-")
	suffix = strings.ReplaceAll(suffix, ".", "-")
	return suffix
}

// executeBackup performs the actual backup operation
func executeBackup(runner runner.Runner, command, outputFile, pod, namespace, path string, debug bool) error {
	if debug {
		fmt.Printf("Debug: Starting backup execution\n")
	}
	
	// Generate output filename if not provided
	if outputFile == "" {
		extension := getFileExtension(getCompressionFromCommand(command))
		outputFile = fmt.Sprintf("backup-%s-%s-%s%s", namespace, pod, generatePathSuffix(path), extension)
	}
	
	if debug {
		fmt.Printf("Debug: Output file: %s\n", outputFile)
	}
	
	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputFile, err)
	}
	defer outFile.Close()
	
	if debug {
		fmt.Printf("Debug: Created output file, executing kubectl command\n")
	}
	
	// Execute kubectl command and stream to file
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	
	// Execute the command and get output
	output, err := runner.Run(parts[0], parts[1:]...)
	if err != nil {
		return fmt.Errorf("backup command failed: %w", err)
	}
	
	// Write output to file
	_, err = outFile.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}
	
	// Get file info for success message
	fileInfo, err := outFile.Stat()
	if err != nil {
		fmt.Printf("Backup completed successfully: %s\n", outputFile)
	} else {
		fmt.Printf("Backup completed successfully: %s (%d bytes)\n", outputFile, fileInfo.Size())
	}
	
	if debug {
		fmt.Printf("Debug: Backup execution completed\n")
	}
	
	return nil
}

// getCompressionFromCommand extracts compression type from kubectl tar command
func getCompressionFromCommand(command string) string {
	if strings.Contains(command, "-z") {
		return "gzip"
	}
	if strings.Contains(command, "-j") {
		return "bzip2"
	}
	if strings.Contains(command, "-J") {
		return "xz"
	}
	return "none"
}

// getFileExtension returns file extension based on compression type
func getFileExtension(compression string) string {
	switch compression {
	case "gzip":
		return ".tar.gz"
	case "bzip2":
		return ".tar.bz2"
	case "xz":
		return ".tar.xz"
	case "none":
		return ".tar"
	default:
		return ".tar.gz"
	}
}
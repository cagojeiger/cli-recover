package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Options represents backup execution options
type Options struct {
	Pod       string
	Namespace string
	Path      string
	SplitSize string
	Output    string
}

// Execute performs the actual backup operation
func Execute(opts *Options) error {
	// Create output directory
	timestamp := time.Now().Format("20060102-150405")
	outputDir := filepath.Join(opts.Output, fmt.Sprintf("backup-%s-%s", opts.Pod, timestamp))
	
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Starting backup...\n")
	fmt.Printf("Pod: %s\n", opts.Pod)
	fmt.Printf("Namespace: %s\n", opts.Namespace)
	fmt.Printf("Path: %s\n", opts.Path)
	fmt.Printf("Output: %s\n", outputDir)
	fmt.Printf("Split Size: %s\n", opts.SplitSize)
	fmt.Println()

	// Execute backup command
	return executeBackupCommand(opts, outputDir)
}

// executeBackupCommand runs the actual kubectl exec + tar + split pipeline
func executeBackupCommand(opts *Options, outputDir string) error {
	// Construct the backup filename
	backupBase := filepath.Join(outputDir, fmt.Sprintf("%s-backup.tar.gz", opts.Pod))
	
	// Build the command pipeline
	// kubectl exec pod -- tar -czf - /path | split -b 1G - backup.tar.gz.
	
	// Step 1: kubectl exec with tar
	kubectlCmd := exec.Command("kubectl", "exec", "-n", opts.Namespace, opts.Pod, "--", 
		"tar", "-czf", "-", opts.Path)
	
	// Step 2: split command
	splitCmd := exec.Command("split", "-b", opts.SplitSize, "-", backupBase+".")
	
	// Connect the pipeline
	pipe, err := kubectlCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	splitCmd.Stdin = pipe
	splitCmd.Stderr = os.Stderr
	
	// Start split command first
	fmt.Println("Starting split process...")
	if err := splitCmd.Start(); err != nil {
		return fmt.Errorf("failed to start split command: %w", err)
	}
	
	// Start kubectl command
	fmt.Println("Starting kubectl exec + tar...")
	if err := kubectlCmd.Start(); err != nil {
		return fmt.Errorf("failed to start kubectl command: %w", err)
	}
	
	// Wait for kubectl to finish
	if err := kubectlCmd.Wait(); err != nil {
		return fmt.Errorf("kubectl command failed: %w", err)
	}
	
	// Close the pipe
	pipe.Close()
	
	// Wait for split to finish
	if err := splitCmd.Wait(); err != nil {
		return fmt.Errorf("split command failed: %w", err)
	}
	
	fmt.Println("Backup completed successfully!")
	
	// List created files
	return listBackupFiles(outputDir)
}

// listBackupFiles shows the created backup files
func listBackupFiles(outputDir string) error {
	files, err := filepath.Glob(filepath.Join(outputDir, "*"))
	if err != nil {
		return fmt.Errorf("failed to list backup files: %w", err)
	}
	
	if len(files) == 0 {
		fmt.Println("Warning: No backup files found!")
		return nil
	}
	
	fmt.Printf("\nCreated backup files:\n")
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		
		size := formatFileSize(info.Size())
		fmt.Printf("  %s (%s)\n", filepath.Base(file), size)
	}
	
	fmt.Printf("\nBackup location: %s\n", outputDir)
	return nil
}

// formatFileSize converts bytes to human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ValidateOptions checks if backup options are valid
func ValidateOptions(opts *Options) error {
	if opts.Pod == "" {
		return fmt.Errorf("pod name is required")
	}
	
	if opts.Path == "" {
		return fmt.Errorf("path is required")
	}
	
	if opts.Namespace == "" {
		opts.Namespace = "default"
	}
	
	if opts.SplitSize == "" {
		opts.SplitSize = "1G"
	}
	
	if opts.Output == "" {
		opts.Output = "./backup"
	}
	
	// Validate split size format
	if !isValidSplitSize(opts.SplitSize) {
		return fmt.Errorf("invalid split size format: %s (use format like 1G, 500M, etc.)", opts.SplitSize)
	}
	
	return nil
}

// isValidSplitSize checks if split size has valid format
func isValidSplitSize(size string) bool {
	if len(size) < 2 {
		return false
	}
	
	// Check if it ends with valid unit
	unit := strings.ToUpper(size[len(size)-1:])
	if unit != "B" && unit != "K" && unit != "M" && unit != "G" && unit != "T" {
		return false
	}
	
	// Check if the rest is a number
	number := size[:len(size)-1]
	if number == "" {
		return false
	}
	
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}
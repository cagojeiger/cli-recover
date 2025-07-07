package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/runner"
)

// newFilesystemBackupCmd creates the filesystem backup command
func newFilesystemBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filesystem [pod] [path]",
		Short: "Backup pod filesystem",
		Long:  `Backup files and directories from a pod's filesystem`,
		Args:  cobra.ExactArgs(2),
		RunE:  runFilesystemBackup,
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
	
	return cmd
}

func runFilesystemBackup(cmd *cobra.Command, args []string) error {
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
		fmt.Printf("Debug: Filesystem backup\n")
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

// estimateBackupSize estimates the total size of the backup by measuring directory size
func estimateBackupSize(runner runner.Runner, pod, namespace, path string, debug bool) int64 {
	if debug {
		fmt.Printf("Debug: Estimating backup size for %s\n", path)
	}
	
	// Use 'du -sb' to get size in bytes
	sizeCmd := fmt.Sprintf("kubectl exec -n %s %s -- du -sb %s", namespace, pod, path)
	
	if debug {
		fmt.Printf("Debug: Size estimation command: %s\n", sizeCmd)
	}
	
	// Set a reasonable timeout for size estimation (30 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "sh", "-c", sizeCmd)
	output, err := cmd.Output()
	if err != nil {
		if debug {
			fmt.Printf("Debug: Failed to estimate size: %v\n", err)
		}
		return 0 // Return 0 if estimation fails
	}
	
	// Parse du output: "12345678\t/path"
	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		if debug {
			fmt.Printf("Debug: No size output received\n")
		}
		return 0
	}
	
	sizeStr := parts[0]
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		if debug {
			fmt.Printf("Debug: Failed to parse size '%s': %v\n", sizeStr, err)
		}
		return 0
	}
	
	if debug {
		fmt.Printf("Debug: Estimated size: %d bytes (%s)\n", size, humanizeBytes(size))
	}
	
	return size
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
	
	// Progress notification
	fmt.Fprintf(os.Stderr, "[START] Backing up %s from pod %s/%s\n", path, namespace, pod)
	fmt.Fprintf(os.Stderr, "[INFO] Output file: %s\n", outputFile)
	
	// Try to estimate size (best effort, may fail)
	fmt.Fprintf(os.Stderr, "[INFO] Estimating backup size...\n")
	estimatedSize := estimateBackupSize(runner, pod, namespace, path, debug)
	if estimatedSize > 0 {
		fmt.Fprintf(os.Stderr, "[INFO] Estimated size: %s\n", humanizeBytes(estimatedSize))
	} else {
		fmt.Fprintf(os.Stderr, "[INFO] Size estimation failed, progress percentage will not be available\n")
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
	
	// For progress tracking with verbose mode
	if strings.Contains(command, "--verbose") {
		return executeBackupWithProgress(runner, parts, outFile, outputFile, estimatedSize, debug)
	}
	
	// For non-verbose mode, show progress bar with ETA
	return executeBackupWithProgressBar(runner, parts, outFile, outputFile, estimatedSize, debug)
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

// executeBackupWithProgress performs backup with real-time progress output
func executeBackupWithProgress(runner runner.Runner, parts []string, outFile *os.File, outputFile string, estimatedSize int64, debug bool) error {
	if debug {
		fmt.Printf("Debug: Executing backup with progress tracking\n")
	}
	
	// Create command directly for streaming output
	cmd := exec.Command(parts[0], parts[1:]...)
	
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start backup command: %w", err)
	}
	
	// Track progress
	var fileCount int
	var written int64
	var mu sync.Mutex
	startTime := time.Now()
	
	// Process stderr (tar verbose output) in a goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		lastProgressUpdate := time.Now()
		
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" && !strings.Contains(line, "Defaulted container") && !strings.Contains(line, "tar: Removing leading") {
				mu.Lock()
				fileCount++
				
				// Update progress every 100ms to avoid spam
				if time.Since(lastProgressUpdate) > 100*time.Millisecond {
					elapsed := time.Since(startTime)
					filesPerSecond := float64(fileCount) / elapsed.Seconds()
					
					// Calculate progress and ETA if we have estimated size
					progressMsg := fmt.Sprintf("[PROGRESS] %d files processed (%.1f files/sec)", 
						fileCount, filesPerSecond)
					
					if estimatedSize > 0 && written > 0 {
						progress := float64(written) / float64(estimatedSize) * 100
						if progress > 100 {
							progress = 100
						}
						
						// Calculate ETA based on current speed
						bytesPerSecond := float64(written) / elapsed.Seconds()
						if bytesPerSecond > 0 {
							remainingBytes := estimatedSize - written
							etaSeconds := float64(remainingBytes) / bytesPerSecond
							eta := time.Duration(etaSeconds) * time.Second
							
							progressMsg += fmt.Sprintf(" - %.1f%% complete, ETA: %s", 
								progress, eta.Round(time.Second))
						}
					}
					
					fmt.Fprintf(os.Stderr, "%s\n", progressMsg)
					lastProgressUpdate = time.Now()
				}
				mu.Unlock()
			}
		}
		
		// Final file count
		mu.Lock()
		fmt.Fprintf(os.Stderr, "[INFO] Total files: %d\n", fileCount)
		mu.Unlock()
	}()
	
	// Copy stdout to file with progress tracking
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			nw, werr := outFile.Write(buf[:n])
			if werr != nil {
				return fmt.Errorf("failed to write backup data: %w", werr)
			}
			mu.Lock()
			written += int64(nw)
			mu.Unlock()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read backup data: %w", err)
		}
	}
	
	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("backup command failed: %w", err)
	}
	
	// Final report
	mu.Lock()
	totalTime := time.Since(startTime)
	throughput := float64(written) / totalTime.Seconds()
	fmt.Fprintf(os.Stderr, "[DONE] Backup completed: %s (%s, %d files in %s, %s/s)\n", 
		outputFile, humanizeBytes(written), fileCount, totalTime.Round(time.Second), humanizeBytes(int64(throughput)))
	fmt.Printf("Backup completed successfully: %s (%s, %d files processed in %s)\n", 
		outputFile, humanizeBytes(written), fileCount, totalTime.Round(time.Second))
	mu.Unlock()
	
	return nil
}

// humanizeBytes converts bytes to human readable format
func humanizeBytes(bytes int64) string {
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

// executeBackupWithProgressBar performs backup with progress bar for non-verbose mode
func executeBackupWithProgressBar(runner runner.Runner, parts []string, outFile *os.File, outputFile string, estimatedSize int64, debug bool) error {
	if debug {
		fmt.Printf("Debug: Executing backup with progress bar\n")
	}
	
	// Create command directly for streaming output
	cmd := exec.Command(parts[0], parts[1:]...)
	
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start backup command: %w", err)
	}
	
	// Discard stderr for non-verbose mode
	go io.Copy(io.Discard, stderr)
	
	// Track progress
	var written int64
	var mu sync.Mutex
	startTime := time.Now()
	done := make(chan bool)
	
	// Progress bar updater
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				mu.Lock()
				elapsed := time.Since(startTime)
				throughput := float64(written) / elapsed.Seconds()
				
				// Build progress message
				progressMsg := fmt.Sprintf("[PROGRESS] %s written (%s/s)", 
					humanizeBytes(written), humanizeBytes(int64(throughput)))
				
				// Add progress percentage and ETA if we have estimated size
				if estimatedSize > 0 && written > 0 {
					progress := float64(written) / float64(estimatedSize) * 100
					if progress > 100 {
						progress = 100
					}
					
					// Calculate ETA
					if throughput > 0 {
						remainingBytes := estimatedSize - written
						etaSeconds := float64(remainingBytes) / throughput
						eta := time.Duration(etaSeconds) * time.Second
						
						progressMsg += fmt.Sprintf(" - %.1f%% complete, ETA: %s", 
							progress, eta.Round(time.Second))
					}
				}
				
				// Clear line and show progress
				fmt.Fprintf(os.Stderr, "\r%s", progressMsg)
				mu.Unlock()
			}
		}
	}()
	
	// Copy stdout to file with progress tracking
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			nw, werr := outFile.Write(buf[:n])
			if werr != nil {
				close(done)
				return fmt.Errorf("failed to write backup data: %w", werr)
			}
			mu.Lock()
			written += int64(nw)
			mu.Unlock()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			close(done)
			return fmt.Errorf("failed to read backup data: %w", err)
		}
	}
	
	// Stop progress bar
	close(done)
	
	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("backup command failed: %w", err)
	}
	
	// Final report
	totalTime := time.Since(startTime)
	throughput := float64(written) / totalTime.Seconds()
	
	// Clear progress line and show final result
	fmt.Fprintf(os.Stderr, "\r\033[K") // Clear line
	fmt.Fprintf(os.Stderr, "[DONE] Backup completed: %s (%s in %s, %s/s)\n", 
		outputFile, humanizeBytes(written), totalTime.Round(time.Second), humanizeBytes(int64(throughput)))
	fmt.Printf("Backup completed successfully: %s (%s)\n", outputFile, humanizeBytes(written))
	
	return nil
}
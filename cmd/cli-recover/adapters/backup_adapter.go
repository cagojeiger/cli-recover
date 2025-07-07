package adapters

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/providers"
)

// BackupAdapter adapts CLI commands to Provider interface
type BackupAdapter struct {
	registry *backup.Registry
}

// NewBackupAdapter creates a new backup adapter
func NewBackupAdapter() *BackupAdapter {
	// Initialize kubernetes clients
	executor := kubernetes.NewOSCommandExecutor()
	kubeClient := kubernetes.NewKubectlClient(executor)
	
	// Register all providers
	if err := providers.RegisterProviders(kubeClient, executor); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register providers: %v\n", err)
	}
	
	return &BackupAdapter{
		registry: backup.GlobalRegistry,
	}
}

// ExecuteBackup executes a backup using the specified provider
func (a *BackupAdapter) ExecuteBackup(providerName string, cmd *cobra.Command, args []string) error {
	// Create provider instance
	provider, err := a.registry.Create(providerName)
	if err != nil {
		return fmt.Errorf("failed to create provider %s: %w", providerName, err)
	}

	// Build options from command flags and args
	opts, err := a.buildOptions(providerName, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to build options: %w", err)
	}

	// Get debug flag
	debug, _ := cmd.Flags().GetBool("debug")
	
	// Validate options
	if err := provider.ValidateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Check dry-run
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Printf("Dry run - would execute %s backup\n", providerName)
		fmt.Printf("Options: %+v\n", opts)
		return nil
	}

	// Estimate size if possible
	fmt.Fprintf(os.Stderr, "[INFO] Estimating backup size...\n")
	estimatedSize, err := provider.EstimateSize(opts)
	if err != nil {
		if debug {
			fmt.Printf("Debug: Size estimation failed: %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "[INFO] Size estimation failed, progress percentage will not be available\n")
		estimatedSize = 0
	} else {
		fmt.Fprintf(os.Stderr, "[INFO] Estimated size: %s\n", humanizeBytes(estimatedSize))
	}

	// Start progress monitoring
	progressDone := make(chan bool)
	go a.monitorProgress(provider, estimatedSize, progressDone, opts.Extra["verbose"].(bool))

	// Execute backup with context
	ctx := context.Background()
	startTime := time.Now()
	
	fmt.Fprintf(os.Stderr, "[START] Starting %s backup\n", providerName)
	fmt.Fprintf(os.Stderr, "[INFO] Output file: %s\n", opts.OutputFile)
	
	// Execute backup - provider handles file writing
	err = provider.Execute(ctx, opts)
	
	// Stop progress monitoring
	close(progressDone)
	
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Final report
	elapsed := time.Since(startTime)
	
	// Get file size if file exists
	var size int64
	if fileInfo, err := os.Stat(opts.OutputFile); err == nil {
		size = fileInfo.Size()
	}
	
	fmt.Fprintf(os.Stderr, "\r\033[K") // Clear progress line
	fmt.Fprintf(os.Stderr, "[DONE] Backup completed: %s (%s in %s)\n", 
		opts.OutputFile, humanizeBytes(size), elapsed.Round(time.Second))
	fmt.Printf("Backup completed successfully: %s (%s)\n", opts.OutputFile, humanizeBytes(size))
	
	return nil
}

// buildOptions builds backup options from command flags
func (a *BackupAdapter) buildOptions(providerName string, cmd *cobra.Command, args []string) (backup.Options, error) {
	opts := backup.Options{
		Extra: make(map[string]interface{}),
	}

	// Common options
	opts.Namespace, _ = cmd.Flags().GetString("namespace")
	
	// Provider-specific options
	switch providerName {
	case "filesystem":
		if len(args) < 2 {
			return opts, fmt.Errorf("filesystem backup requires [pod] [path] arguments")
		}
		opts.PodName = args[0]
		opts.SourcePath = args[1]
		
		// Compression
		compression, _ := cmd.Flags().GetString("compression")
		opts.Compress = compression != "none"
		opts.Extra["compression"] = compression
		
		// Exclude patterns
		opts.Exclude, _ = cmd.Flags().GetStringSlice("exclude")
		excludeVCS, _ := cmd.Flags().GetBool("exclude-vcs")
		if excludeVCS {
			opts.Exclude = append(opts.Exclude, ".git", ".svn", ".hg", ".bzr")
		}
		
		// Other options
		opts.Container, _ = cmd.Flags().GetString("container")
		opts.OutputFile, _ = cmd.Flags().GetString("output")
		
		// Generate output filename if not provided
		if opts.OutputFile == "" {
			ext := getFileExtension(compression)
			opts.OutputFile = fmt.Sprintf("backup-%s-%s-%s%s", 
				opts.Namespace, opts.PodName, 
				sanitizePath(opts.SourcePath), ext)
		}
		
		// Store extra flags
		opts.Extra["verbose"], _ = cmd.Flags().GetBool("verbose")
		opts.Extra["totals"], _ = cmd.Flags().GetBool("totals")
		opts.Extra["preserve-perms"], _ = cmd.Flags().GetBool("preserve-perms")
		
	// TODO: Add MinIO options
	// case "minio":
	
	// TODO: Add MongoDB options  
	// case "mongodb":
		
	case "test":
		// Test provider for unit tests
		opts.PodName = "test-pod"
		opts.SourcePath = "/test"
		opts.OutputFile = "test-output.tar"
		opts.Extra["verbose"] = false
		
	default:
		return opts, fmt.Errorf("unknown provider: %s", providerName)
	}
	
	return opts, nil
}

// monitorProgress monitors and displays backup progress
func (a *BackupAdapter) monitorProgress(provider backup.Provider, estimatedSize int64, done <-chan bool, verbose bool) {
	progressCh := provider.StreamProgress()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	var lastProgress backup.Progress
	startTime := time.Now()
	
	for {
		select {
		case <-done:
			return
		case progress, ok := <-progressCh:
			if ok {
				lastProgress = progress
				if verbose {
					// In verbose mode, show each file
					fmt.Fprintf(os.Stderr, "[PROGRESS] %s\n", progress.Message)
				}
			}
		case <-ticker.C:
			// Update progress bar
			if !verbose && lastProgress.Total > 0 {
				elapsed := time.Since(startTime)
				throughput := float64(lastProgress.Current) / elapsed.Seconds()
				
				progressMsg := fmt.Sprintf("[PROGRESS] %s/%s (%s/s)",
					humanizeBytes(lastProgress.Current),
					humanizeBytes(lastProgress.Total),
					humanizeBytes(int64(throughput)))
				
				// Add percentage and ETA
				if estimatedSize > 0 || lastProgress.Total > 0 {
					total := estimatedSize
					if lastProgress.Total > 0 {
						total = lastProgress.Total
					}
					
					percentage := float64(lastProgress.Current) / float64(total) * 100
					if percentage > 100 {
						percentage = 100
					}
					
					// Calculate ETA
					if throughput > 0 {
						remaining := total - lastProgress.Current
						etaSeconds := float64(remaining) / throughput
						eta := time.Duration(etaSeconds) * time.Second
						
						progressMsg += fmt.Sprintf(" - %.1f%% complete, ETA: %s",
							percentage, eta.Round(time.Second))
					}
				}
				
				// Clear line and show progress
				fmt.Fprintf(os.Stderr, "\r%s", progressMsg)
			}
		}
	}
}

// Helper functions

func sanitizePath(path string) string {
	if path == "/" {
		return "root"
	}
	// Remove leading slash and replace special chars
	s := strings.TrimPrefix(path, "/")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "-")
	return s
}

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
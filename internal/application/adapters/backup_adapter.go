package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/log"
	"github.com/cagojeiger/cli-recover/internal/domain/log/storage"
	"github.com/cagojeiger/cli-recover/internal/domain/logger"
	"github.com/cagojeiger/cli-recover/internal/domain/metadata"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	infLogger "github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/providers"
)

// BackupAdapter adapts CLI commands to Provider interface
type BackupAdapter struct {
	registry *backup.Registry
	logger   logger.Logger
}

// NewBackupAdapter creates a new backup adapter
func NewBackupAdapter() *BackupAdapter {
	// Initialize kubernetes clients
	executor := kubernetes.NewOSCommandExecutor()
	kubeClient := kubernetes.NewKubectlClient(executor)
	
	// Get logger
	log := infLogger.GetGlobalLogger()
	
	// Register all providers
	if err := providers.RegisterProviders(kubeClient, executor); err != nil {
		log.Error("Failed to register providers", logger.F("error", err))
	}
	
	return &BackupAdapter{
		registry: backup.GlobalRegistry,
		logger:   log,
	}
}

// ExecuteBackup executes a backup using the specified provider
func (a *BackupAdapter) ExecuteBackup(providerName string, cmd *cobra.Command, args []string) error {
	// Initialize log system
	logDir := getLogDirFromCmd(cmd)
	logRepo, err := storage.NewFileRepository(logDir)
	if err != nil {
		a.logger.Warn("Failed to initialize log repository", logger.F("error", err))
		// Continue without logging to file
	}
	
	var logService *log.Service
	var logWriter *log.Writer
	var logEntry *log.Log
	
	if logRepo != nil {
		logService = log.NewService(logRepo, filepath.Join(logDir, "files"))
		
		// Build metadata for log
		metadata := make(map[string]string)
		if ns, _ := cmd.Flags().GetString("namespace"); ns != "" {
			metadata["namespace"] = ns
		}
		if len(args) >= 2 {
			metadata["pod"] = args[0]
			metadata["path"] = args[1]
		}
		
		// Start log
		logEntry, logWriter, err = logService.StartLog(log.TypeBackup, providerName, metadata)
		if err != nil {
			a.logger.Warn("Failed to start log", logger.F("error", err))
		} else {
			defer func() {
				if logWriter != nil {
					logWriter.Close()
				}
			}()
			
			// Create tee writer to write to both console and log file
			if logWriter != nil {
				// Override logger to also write to log file
				a.logger.Info("Log file created", logger.F("log_id", logEntry.ID), logger.F("file", logEntry.FilePath))
			}
		}
	}

	// Create provider instance
	provider, err := a.registry.Create(providerName)
	if err != nil {
		if logWriter != nil {
			logWriter.WriteLine("ERROR: Failed to create provider %s: %v", providerName, err)
		}
		if logEntry != nil && logService != nil {
			logService.FailLog(logEntry.ID, fmt.Sprintf("Failed to create provider: %v", err))
		}
		return fmt.Errorf("failed to create provider %s: %w", providerName, err)
	}

	// Build options from command flags and args
	opts, err := a.buildOptions(providerName, cmd, args)
	if err != nil {
		if logWriter != nil {
			logWriter.WriteLine("ERROR: Failed to build options: %v", err)
		}
		if logEntry != nil && logService != nil {
			logService.FailLog(logEntry.ID, fmt.Sprintf("Failed to build options: %v", err))
		}
		return fmt.Errorf("failed to build options: %w", err)
	}

	// Log backup details
	if logWriter != nil {
		logWriter.WriteLine("Backup configuration:")
		logWriter.WriteLine("  Provider: %s", providerName)
		logWriter.WriteLine("  Output: %s", opts.OutputFile)
		if opts.Namespace != "" {
			logWriter.WriteLine("  Namespace: %s", opts.Namespace)
		}
		if opts.PodName != "" {
			logWriter.WriteLine("  Pod: %s", opts.PodName)
		}
		if opts.SourcePath != "" {
			logWriter.WriteLine("  Path: %s", opts.SourcePath)
		}
		logWriter.WriteLine("")
	}

	// Get debug flag
	debug, _ := cmd.Flags().GetBool("debug")
	
	// Validate options
	if err := provider.ValidateOptions(opts); err != nil {
		if logWriter != nil {
			logWriter.WriteLine("ERROR: Invalid options: %v", err)
		}
		if logEntry != nil && logService != nil {
			logService.FailLog(logEntry.ID, fmt.Sprintf("Invalid options: %v", err))
		}
		return fmt.Errorf("invalid options: %w", err)
	}

	// Check dry-run
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		a.logger.Info("Dry run - would execute backup", 
			logger.F("provider", providerName),
			logger.F("options", opts))
		if logWriter != nil {
			logWriter.WriteLine("DRY RUN - Would execute backup with above configuration")
		}
		if logEntry != nil && logService != nil {
			logService.CompleteLog(logEntry.ID)
		}
		return nil
	}

	// Estimate size if possible
	a.logger.Info("Estimating backup size...")
	if logWriter != nil {
		logWriter.WriteLine("Estimating backup size...")
	}
	
	estimatedSize, err := provider.EstimateSize(opts)
	if err != nil {
		if debug {
			a.logger.Debug("Size estimation failed", logger.F("error", err))
		}
		a.logger.Info("Size estimation failed, progress percentage will not be available")
		if logWriter != nil {
			logWriter.WriteLine("Size estimation failed: %v", err)
		}
		estimatedSize = 0
	} else {
		a.logger.Info("Estimated size", logger.F("size", humanizeBytes(estimatedSize)))
		if logWriter != nil {
			logWriter.WriteLine("Estimated size: %s", humanizeBytes(estimatedSize))
		}
	}

	// Start progress monitoring
	progressDone := make(chan bool)
	go a.monitorProgress(provider, estimatedSize, progressDone, opts.Extra["verbose"].(bool))

	// Execute backup with context
	ctx := context.Background()
	startTime := time.Now()
	
	a.logger.Info("Starting backup", 
		logger.F("provider", providerName),
		logger.F("output_file", opts.OutputFile))
	
	if logWriter != nil {
		logWriter.WriteLine("Starting backup at %s", startTime.Format(time.RFC3339))
		logWriter.WriteLine("")
	}
	
	// Execute backup - provider handles file writing
	err = provider.Execute(ctx, opts)
	
	// Stop progress monitoring
	close(progressDone)
	
	if err != nil {
		if logWriter != nil {
			logWriter.WriteLine("ERROR: Backup failed: %v", err)
		}
		if logEntry != nil && logService != nil {
			logService.FailLog(logEntry.ID, fmt.Sprintf("Backup failed: %v", err))
		}
		return fmt.Errorf("backup failed: %w", err)
	}

	// Final report
	elapsed := time.Since(startTime)
	
	// Get file size if file exists
	var size int64
	if fileInfo, err := os.Stat(opts.OutputFile); err == nil {
		size = fileInfo.Size()
	}
	
	// Save metadata
	if err := a.saveMetadata(providerName, opts, size, startTime, time.Now()); err != nil {
		if debug {
			a.logger.Debug("Failed to save metadata", logger.F("error", err))
		}
		// Don't fail the backup operation if metadata save fails
	}
	
	a.logger.Info("Backup completed successfully",
		logger.F("output_file", opts.OutputFile),
		logger.F("size", humanizeBytes(size)),
		logger.F("duration", elapsed.Round(time.Second).String()))
	
	if logWriter != nil {
		logWriter.WriteLine("")
		logWriter.WriteLine("Backup completed successfully")
		logWriter.WriteLine("  Output file: %s", opts.OutputFile)
		logWriter.WriteLine("  Size: %s", humanizeBytes(size))
		logWriter.WriteLine("  Duration: %s", elapsed.Round(time.Second))
	}
	
	if logEntry != nil && logService != nil {
		logEntry.SetMetadata("output_file", opts.OutputFile)
		logEntry.SetMetadata("size", fmt.Sprintf("%d", size))
		logService.CompleteLog(logEntry.ID)
	}
	
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
		
		// Compression - only support none (.tar) and gzip (.tar.gz)
		compression, _ := cmd.Flags().GetString("compression")
		if compression != "none" && compression != "gzip" {
			return opts, fmt.Errorf("unsupported compression type '%s', only 'none' (.tar) and 'gzip' (.tar.gz) are supported", compression)
		}
		opts.Compress = compression == "gzip"
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
	case "none":
		return ".tar"
	default:
		return ".tar" // Default to uncompressed
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

// saveMetadata saves backup metadata to the metadata store
func (a *BackupAdapter) saveMetadata(providerName string, opts backup.Options, size int64, startTime, endTime time.Time) error {
	// Calculate checksum of the backup file
	checksum, err := calculateFileChecksum(opts.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	// Create metadata
	backupMetadata := &restore.Metadata{
		Type:        providerName,
		Namespace:   opts.Namespace,
		PodName:     opts.PodName,
		SourcePath:  opts.SourcePath,
		BackupFile:  opts.OutputFile,
		Size:        size,
		Checksum:    checksum,
		CreatedAt:   startTime,
		CompletedAt: endTime,
		Status:      "completed",
		ProviderInfo: map[string]interface{}{
			"container": opts.Container,
			"exclude":   opts.Exclude,
			"compress":  opts.Compress,
		},
	}
	
	// Add compression info
	if compression, ok := opts.Extra["compression"].(string); ok {
		backupMetadata.Compression = compression
	}
	
	// Save to metadata store
	metadataStore := metadata.DefaultStore
	if metadataStore != nil {
		return metadataStore.Save(backupMetadata)
	}
	
	return nil
}

// calculateFileChecksum calculates SHA256 checksum of a file
func calculateFileChecksum(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// getLogDirFromCmd returns the log directory from command flags or default
func getLogDirFromCmd(cmd *cobra.Command) string {
	logDir, _ := cmd.Flags().GetString("log-dir")
	if logDir == "" {
		homeDir, _ := os.UserHomeDir()
		logDir = filepath.Join(homeDir, ".cli-recover", "logs")
	}
	return logDir
}
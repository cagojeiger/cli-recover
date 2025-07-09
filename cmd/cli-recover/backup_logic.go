package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	domainLog "github.com/cagojeiger/cli-recover/internal/domain/log"
	"github.com/cagojeiger/cli-recover/internal/domain/logger"
	"github.com/cagojeiger/cli-recover/internal/domain/metadata"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/cagojeiger/cli-recover/internal/infrastructure"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/log/storage"
	infLogger "github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/progress"
)

// executeBackup contains the integrated backup logic from the adapter
func executeBackup(providerName string, cmd *cobra.Command, args []string) error {
	// Initialize kubernetes clients
	executor := kubernetes.NewOSCommandExecutor()
	kubeClient := kubernetes.NewKubectlClient(executor)

	// Get logger
	log := infLogger.GetGlobalLogger()

	// No need to register providers - we'll use factory directly

	// Initialize log system
	logDir := getLogDirFromCmd(cmd)
	logRepo, err := storage.NewFileRepository(logDir)
	if err != nil {
		log.Warn("Failed to initialize log repository", logger.F("error", err))
		// Continue without logging to file
	}

	var logService *domainLog.Service
	var logWriter *domainLog.Writer
	var logEntry *domainLog.Log

	if logRepo != nil {
		logService = domainLog.NewService(logRepo, filepath.Join(logDir, "files"))

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
		logEntry, logWriter, err = logService.StartLog(domainLog.TypeBackup, providerName, metadata)
		if err != nil {
			log.Warn("Failed to start log", logger.F("error", err))
		} else {
			defer func() {
				if logWriter != nil {
					logWriter.Close()
				}
			}()

			// Create tee writer to write to both console and log file
			if logWriter != nil {
				// Override logger to also write to log file
				log.Info("Log file created", logger.F("log_id", logEntry.ID), logger.F("file", logEntry.FilePath))
			}
		}
	}

	// Create provider instance using factory
	provider, err := infrastructure.CreateBackupProvider(providerName, kubeClient, executor)
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
	opts, err := buildBackupOptions(providerName, cmd, args)
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
		log.Info("Dry run - would execute backup",
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
	log.Info("Estimating backup size...")
	if logWriter != nil {
		logWriter.WriteLine("Estimating backup size...")
	}

	estimatedSize, err := provider.EstimateSize(opts)
	if err != nil {
		if debug {
			log.Debug("Size estimation failed", logger.F("error", err))
		}
		log.Info("Size estimation failed, progress percentage will not be available")
		if logWriter != nil {
			logWriter.WriteLine("Size estimation failed: %v", err)
		}
		estimatedSize = 0
	} else {
		log.Info("Estimated size", logger.F("size", humanizeBytes(estimatedSize)))
		if logWriter != nil {
			logWriter.WriteLine("Estimated size: %s", humanizeBytes(estimatedSize))
		}
	}

	// Start progress monitoring
	progressDone := make(chan bool)
	go monitorBackupProgress(provider, estimatedSize, progressDone, opts.Extra["verbose"].(bool))

	// Execute backup with context
	ctx := context.Background()
	startTime := time.Now()

	log.Info("Starting backup",
		logger.F("provider", providerName),
		logger.F("output_file", opts.OutputFile))

	if logWriter != nil {
		logWriter.WriteLine("Starting backup at %s", startTime.Format(time.RFC3339))
		logWriter.WriteLine("")
	}

	// Execute backup with result - provider handles file writing and checksum
	result, err := provider.ExecuteWithResult(ctx, opts)

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

	// Use size from result
	size := result.Size

	// Save metadata with checksum from result
	if err := saveBackupMetadataWithChecksum(providerName, opts, size, startTime, time.Now(), result.Checksum); err != nil {
		if debug {
			log.Debug("Failed to save metadata", logger.F("error", err))
		}
		// Don't fail the backup operation if metadata save fails
	}

	log.Info("Backup completed successfully",
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

// buildBackupOptions builds backup options from command flags
func buildBackupOptions(providerName string, cmd *cobra.Command, args []string) (backup.Options, error) {
	opts := backup.Options{
		Extra: make(map[string]interface{}),
	}

	// Common options
	opts.Namespace, _ = cmd.Flags().GetString("namespace")

	// Provider-specific options
	switch providerName {
	case "filesystem":
		if len(args) < 2 {
			return opts, &CLIError{
				Message: "Missing required arguments",
				Reason:  "Filesystem backup requires both pod name and path",
				Fix:     "Usage: cli-recover backup filesystem [pod] [path]",
			}
		}
		opts.PodName = args[0]
		opts.SourcePath = args[1]

		// Compression - only support none (.tar) and gzip (.tar.gz)
		compression, _ := cmd.Flags().GetString("compression")
		if compression != "none" && compression != "gzip" {
			return opts, NewInvalidFlagError("--compression", compression, "'none' or 'gzip'")
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

// monitorBackupProgress monitors and displays backup progress
func monitorBackupProgress(provider backup.Provider, estimatedSize int64, done <-chan bool, verbose bool) {
	progressCh := provider.StreamProgress()

	// Create the appropriate progress reporter
	reporter := progress.NewAutoReporter(os.Stderr)

	// Start the operation
	reporter.Start("Backup", estimatedSize)

	// Track last update time for throttling
	lastUpdate := time.Time{}
	updateInterval := 100 * time.Millisecond // Throttle updates

	for {
		select {
		case <-done:
			// Operation completed
			reporter.Complete()
			// Clear any remaining progress output for terminal
			if _, ok := reporter.(*progress.TerminalReporter); ok {
				fmt.Fprint(os.Stderr, "") // Terminal reporter already adds newline
			}
			return

		case progress, ok := <-progressCh:
			if !ok {
				continue
			}

			// In verbose mode, still show individual file messages
			if verbose && progress.Message != "" && !strings.Contains(progress.Message, "Written") && !strings.Contains(progress.Message, "Backing up:") {
				fmt.Fprintf(os.Stderr, "[VERBOSE] %s\n", progress.Message)
			}

			// Update progress reporter (it handles throttling internally)
			now := time.Now()
			if now.Sub(lastUpdate) >= updateInterval || progress.Current == progress.Total {
				// Use estimated size if progress doesn't include total
				total := progress.Total
				if total == 0 && estimatedSize > 0 {
					total = estimatedSize
				}

				reporter.Update(progress.Current, total)
				lastUpdate = now
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

// saveBackupMetadataWithChecksum saves backup metadata to the metadata store with pre-calculated checksum
func saveBackupMetadataWithChecksum(providerName string, opts backup.Options, size int64, startTime, endTime time.Time, checksum string) error {

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

// getLogDirFromCmd returns the log directory from command flags or default
func getLogDirFromCmd(cmd *cobra.Command) string {
	logDir, _ := cmd.Flags().GetString("log-dir")
	if logDir == "" {
		homeDir, _ := os.UserHomeDir()
		logDir = filepath.Join(homeDir, ".cli-recover", "logs")
	}
	return logDir
}

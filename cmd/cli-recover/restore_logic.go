package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cagojeiger/cli-recover/internal/domain/logger"
	"github.com/cagojeiger/cli-recover/internal/domain/metadata"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/cagojeiger/cli-recover/internal/infrastructure"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	infLogger "github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/progress"
)

// executeRestore contains the integrated restore logic from the adapter
func executeRestore(providerName string, cmd *cobra.Command, args []string) error {
	log := infLogger.GetGlobalLogger()

	// Initialize kubernetes clients
	executor := kubernetes.NewOSCommandExecutor()
	kubeClient := kubernetes.NewKubectlClient(executor)

	// Build options from command flags and args
	opts, err := buildRestoreOptions(providerName, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to build options: %w", err)
	}

	// Create provider instance using factory
	provider, err := infrastructure.CreateRestoreProvider(providerName, kubeClient, executor)
	if err != nil {
		return fmt.Errorf("failed to create provider %s: %w", providerName, err)
	}

	// Get debug flag
	debug, _ := cmd.Flags().GetBool("debug")

	// Validate options
	if err := provider.ValidateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Load metadata if available
	var restoreMetadata *restore.Metadata
	metadataStore := metadata.DefaultStore
	if metadataStore != nil {
		restoreMetadata, _ = metadataStore.GetByFile(opts.BackupFile)
	}

	// Validate backup
	if err := provider.ValidateBackup(opts.BackupFile, restoreMetadata); err != nil {
		return fmt.Errorf("backup validation failed: %w", err)
	}

	// Check dry-run
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		log.Info("Dry run - would execute restore",
			logger.F("provider", providerName),
			logger.F("options", opts))
		return nil
	}

	// Estimate size if possible
	log.Info("Analyzing backup file...")
	estimatedSize, err := provider.EstimateSize(opts.BackupFile)
	if err != nil {
		if debug {
			log.Debug("Size estimation failed", logger.F("error", err))
		}
		log.Info("Size estimation failed, progress percentage will not be available")
		estimatedSize = 0
	} else {
		log.Info("Estimated size", logger.F("size", humanizeBytes(estimatedSize)))
	}

	// Start progress monitoring
	progressDone := make(chan bool)
	go monitorRestoreProgress(provider, estimatedSize, progressDone, opts.Extra["verbose"].(bool))

	// Execute restore with context
	ctx := context.Background()
	startTime := time.Now()

	log.Info("Starting restore",
		logger.F("provider", providerName),
		logger.F("pod", opts.PodName),
		logger.F("target_path", opts.TargetPath))

	// Execute restore
	result, err := provider.Execute(ctx, opts)

	// Stop progress monitoring
	close(progressDone)

	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// Final report
	elapsed := time.Since(startTime)

	// Log completion
	log.Info("Restore completed successfully",
		logger.F("restored_path", result.RestoredPath),
		logger.F("file_count", result.FileCount),
		logger.F("bytes_written", humanizeBytes(result.BytesWritten)),
		logger.F("duration", elapsed.Round(time.Second).String()))

	// Log warnings if any
	for _, warning := range result.Warnings {
		log.Warn("Restore warning", logger.F("warning", warning))
	}

	return nil
}

// buildRestoreOptions builds restore options from command flags
func buildRestoreOptions(providerName string, cmd *cobra.Command, args []string) (restore.Options, error) {
	opts := restore.Options{
		Extra: make(map[string]interface{}),
	}

	// Common options
	opts.Namespace, _ = cmd.Flags().GetString("namespace")

	// Provider-specific options
	switch providerName {
	case "filesystem":
		if len(args) < 2 {
			return opts, fmt.Errorf("filesystem restore requires [pod] [backup-file] arguments")
		}
		opts.PodName = args[0]
		opts.BackupFile = args[1]

		// Target path
		opts.TargetPath, _ = cmd.Flags().GetString("target-path")
		if opts.TargetPath == "" {
			opts.TargetPath = "/"
		}

		// Restore options
		opts.Overwrite, _ = cmd.Flags().GetBool("force")
		opts.PreservePerms, _ = cmd.Flags().GetBool("preserve-perms")
		opts.SkipPaths, _ = cmd.Flags().GetStringSlice("skip-paths")
		opts.Container, _ = cmd.Flags().GetString("container")

		// Store extra flags
		opts.Extra["verbose"], _ = cmd.Flags().GetBool("verbose")

	case "test":
		// Test provider for unit tests
		opts.PodName = "test-pod"
		opts.BackupFile = "test-backup.tar"
		opts.TargetPath = "/test"
		opts.Extra["verbose"] = false

	default:
		return opts, fmt.Errorf("unknown provider: %s", providerName)
	}

	return opts, nil
}

// monitorRestoreProgress monitors and displays restore progress
func monitorRestoreProgress(provider restore.Provider, estimatedSize int64, done <-chan bool, verbose bool) {
	progressCh := provider.StreamProgress()

	// Create the appropriate progress reporter
	reporter := progress.NewAutoReporter(os.Stderr)

	// Start the operation
	reporter.Start("Restore", estimatedSize)

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
			if verbose && progress.Message != "" {
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

// Helper function to get absolute path
func getAbsolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	// Try to make it absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

// Helper function to sanitize paths
func sanitizeTargetPath(path string) string {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// Clean the path
	return filepath.Clean(path)
}

package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
)

// RestoreProvider implements restore.Provider for filesystem restores
type RestoreProvider struct {
	kubeClient kubernetes.KubeClient
	executor   kubernetes.CommandExecutor
	progressCh chan restore.Progress
}

// Ensure RestoreProvider implements restore.Provider interface
var _ restore.Provider = (*RestoreProvider)(nil)

// NewRestoreProvider creates a new filesystem restore provider
func NewRestoreProvider(kubeClient kubernetes.KubeClient, executor kubernetes.CommandExecutor) *RestoreProvider {
	return &RestoreProvider{
		kubeClient: kubeClient,
		executor:   executor,
		progressCh: make(chan restore.Progress, 100),
	}
}

// Name returns the provider name
func (p *RestoreProvider) Name() string {
	return "filesystem"
}

// Description returns the provider description
func (p *RestoreProvider) Description() string {
	return "Restore filesystem to Kubernetes pods"
}

// ValidateOptions validates the restore options
func (p *RestoreProvider) ValidateOptions(opts restore.Options) error {
	if opts.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if opts.PodName == "" {
		return fmt.Errorf("pod name is required")
	}
	if opts.BackupFile == "" {
		return fmt.Errorf("backup file is required")
	}
	if opts.TargetPath == "" {
		return fmt.Errorf("target path is required")
	}

	// Validate target path for security
	if err := validatePath(opts.TargetPath); err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	// Note: File existence check is done in Execute() to allow for remote files
	// or files that will be created later

	return nil
}

// ValidateBackup validates that the backup file is compatible
func (p *RestoreProvider) ValidateBackup(backupFile string, metadata *restore.Metadata) error {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(backupFile))
	validExts := []string{".tar", ".tar.gz", ".tgz", ".tar.bz2", ".tar.xz"}

	valid := false
	for _, validExt := range validExts {
		if strings.HasSuffix(backupFile, validExt) {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("unsupported backup file format: %s", ext)
	}

	// If metadata is provided, validate provider type
	if metadata != nil && metadata.Type != "filesystem" {
		return fmt.Errorf("backup was created by %s provider, not filesystem", metadata.Type)
	}

	return nil
}

// Execute performs the restore operation
func (p *RestoreProvider) Execute(ctx context.Context, opts restore.Options) (*restore.RestoreResult, error) {
	startTime := time.Now()

	// Apply timeout if context doesn't have one
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
	}

	// Check if backup file exists
	fileInfo, err := os.Stat(opts.BackupFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("backup file not found: %s", opts.BackupFile)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access backup file: %w", err)
	}

	// Validate it's a regular file
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("backup file is not a regular file: %s", opts.BackupFile)
	}

	// Verify pod exists
	pods, err := p.kubeClient.GetPods(ctx, opts.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	found := false
	for _, pod := range pods {
		if pod.Name == opts.PodName {
			found = true
			break
		}
	}

	if !found {
		return nil, restore.NewRestoreError("POD_NOT_FOUND",
			fmt.Sprintf("pod %s not found in namespace %s", opts.PodName, opts.Namespace))
	}

	// Send initial progress
	p.progressCh <- restore.Progress{
		Current: 0,
		Total:   fileInfo.Size(),
		Message: "Starting restore operation",
	}

	// Use new StreamRestore function
	progressCh, err := kubernetes.StreamRestore(
		ctx,
		opts.BackupFile,
		opts.Namespace,
		opts.PodName,
		opts.Container,
		opts.TargetPath,
		opts.PreservePerms,
		opts.Overwrite,
		opts.SkipPaths,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start restore: %w", err)
	}

	// Monitor progress
	fileCount := 0
	bytesProcessed := int64(0)
	lastUpdate := time.Now()

	for progress := range progressCh {
		if progress.Error != nil {
			return nil, fmt.Errorf("restore error: %w", progress.Error)
		}

		fileCount = progress.FileCount

		// Estimate bytes processed based on file count
		if fileInfo.Size() > 0 && fileCount > 0 {
			// Simple estimation: assume linear progress
			bytesProcessed = int64(float64(fileCount) * float64(fileInfo.Size()) / 1000.0)
			if bytesProcessed > fileInfo.Size() {
				bytesProcessed = fileInfo.Size()
			}
		}

		// Send progress updates (throttled)
		if time.Since(lastUpdate) >= 100*time.Millisecond {
			p.progressCh <- restore.Progress{
				Current: bytesProcessed,
				Total:   fileInfo.Size(),
				Message: fmt.Sprintf("Restoring: %s", progress.FileName),
			}
			lastUpdate = time.Now()
		}
	}

	// Send completion progress
	p.progressCh <- restore.Progress{
		Current: fileInfo.Size(),
		Total:   fileInfo.Size(),
		Message: "Restore completed successfully",
	}

	duration := time.Since(startTime)

	return &restore.RestoreResult{
		Success:      true,
		RestoredPath: opts.TargetPath,
		FileCount:    fileCount,
		BytesWritten: fileInfo.Size(),
		Duration:     duration,
		Warnings:     []string{},
	}, nil
}

// StreamProgress returns the progress channel
func (p *RestoreProvider) StreamProgress() <-chan restore.Progress {
	return p.progressCh
}

// EstimateSize estimates the size of data to be restored
func (p *RestoreProvider) EstimateSize(backupFile string) (int64, error) {
	// Use stat to get file size
	info, err := os.Stat(backupFile)
	if err != nil {
		return 0, fmt.Errorf("failed to stat backup file: %w", err)
	}

	return info.Size(), nil
}

// ValidatePath validates the target path for security
func validatePath(path string) error {
	// Ensure path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("target path must be absolute: %s", path)
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain '..': %s", path)
	}

	// Warn about dangerous paths
	dangerousPaths := []string{"/", "/etc", "/var", "/usr", "/bin", "/sbin", "/lib", "/lib64"}
	for _, dangerous := range dangerousPaths {
		if path == dangerous || strings.HasPrefix(path, dangerous+"/") {
			// Just a warning, not an error - user might know what they're doing
			fmt.Fprintf(os.Stderr, "⚠️  Warning: Restoring to system directory: %s\n", path)
			break
		}
	}

	return nil
}

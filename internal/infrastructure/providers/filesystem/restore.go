package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	
	// Check if backup file exists
	if _, err := os.Stat(opts.BackupFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup file not found: %s", opts.BackupFile)
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
	
	// Build tar extract command
	tarCmd := p.buildTarCommand(opts)
	
	// Send initial progress
	p.progressCh <- restore.Progress{
		Current: 0,
		Total:   100,
		Message: "Starting restore operation",
	}
	
	// Execute tar command
	outputCh, errorCh := p.executor.Stream(ctx, tarCmd)
	
	// Monitor progress
	fileCount := 0
	go p.monitorProgress(outputCh, &fileCount)
	
	// Wait for completion or error
	select {
	case err := <-errorCh:
		if err != nil {
			return nil, fmt.Errorf("restore failed: %w", err)
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	
	// Send completion progress
	p.progressCh <- restore.Progress{
		Current: 100,
		Total:   100,
		Message: "Restore completed successfully",
	}
	
	duration := time.Since(startTime)
	
	return &restore.RestoreResult{
		Success:      true,
		RestoredPath: opts.TargetPath,
		FileCount:    fileCount,
		Duration:     duration,
		Warnings:     []string{},
	}, nil
}

// buildTarCommand builds the kubectl exec tar command for restore
func (p *RestoreProvider) buildTarCommand(opts restore.Options) []string {
	// Read backup file and pipe to kubectl exec tar
	catCmd := fmt.Sprintf("cat %s", opts.BackupFile)
	
	// Build kubectl exec command
	kubectlArgs := []string{"exec", "-i", "-n", opts.Namespace, opts.PodName}
	
	// Add container if specified
	if opts.Container != "" {
		kubectlArgs = append(kubectlArgs, "-c", opts.Container)
	}
	
	kubectlArgs = append(kubectlArgs, "--", "tar")
	
	// Detect compression from file extension
	compression := detectCompression(opts.BackupFile)
	switch compression {
	case "gzip":
		kubectlArgs = append(kubectlArgs, "-xzf")
	case "bzip2":
		kubectlArgs = append(kubectlArgs, "-xjf")
	case "xz":
		kubectlArgs = append(kubectlArgs, "-xJf")
	default:
		kubectlArgs = append(kubectlArgs, "-xf")
	}
	
	// Add stdin flag
	kubectlArgs = append(kubectlArgs, "-")
	
	// Add target directory
	kubectlArgs = append(kubectlArgs, "-C", opts.TargetPath)
	
	// Add overwrite flag if needed
	if !opts.Overwrite {
		kubectlArgs = append(kubectlArgs, "--keep-old-files")
	}
	
	// Add preserve permissions flag
	if opts.PreservePerms {
		kubectlArgs = append(kubectlArgs, "-p")
	}
	
	// Add skip paths
	for _, skip := range opts.SkipPaths {
		kubectlArgs = append(kubectlArgs, "--exclude="+skip)
	}
	
	// Combine cat and kubectl with pipe
	fullCmd := fmt.Sprintf("%s | kubectl %s", catCmd, strings.Join(kubectlArgs, " "))
	
	return []string{"sh", "-c", fullCmd}
}

// detectCompression detects compression type from file extension
func detectCompression(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return "gzip"
	case strings.HasSuffix(lower, ".tar.bz2"):
		return "bzip2"
	case strings.HasSuffix(lower, ".tar.xz"):
		return "xz"
	default:
		return "none"
	}
}

// monitorProgress monitors tar output and updates progress
func (p *RestoreProvider) monitorProgress(outputCh <-chan string, fileCount *int) {
	extractPattern := regexp.MustCompile(`^(x |extracting:?\s*)(.+)$`)
	
	for line := range outputCh {
		if matches := extractPattern.FindStringSubmatch(line); matches != nil {
			*fileCount++
			p.progressCh <- restore.Progress{
				Current: int64(*fileCount),
				Total:   0, // Unknown total
				Message: fmt.Sprintf("Restoring: %s", matches[2]),
			}
		}
	}
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
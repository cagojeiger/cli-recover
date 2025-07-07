package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
)

// Provider implements backup.Provider for filesystem backups
type Provider struct {
	kubeClient kubernetes.KubeClient
	executor   kubernetes.CommandExecutor
	progressCh chan backup.Progress
}

// Ensure Provider implements backup.Provider interface
var _ backup.Provider = (*Provider)(nil)

// NewProvider creates a new filesystem backup provider
func NewProvider(kubeClient kubernetes.KubeClient, executor kubernetes.CommandExecutor) *Provider {
	return &Provider{
		kubeClient: kubeClient,
		executor:   executor,
		progressCh: make(chan backup.Progress, 100),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "filesystem"
}

// Description returns the provider description
func (p *Provider) Description() string {
	return "Backup filesystem from Kubernetes pods"
}

// ValidateOptions validates the backup options
func (p *Provider) ValidateOptions(opts backup.Options) error {
	if opts.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if opts.PodName == "" {
		return fmt.Errorf("pod name is required")
	}
	if opts.SourcePath == "" {
		return fmt.Errorf("source path is required")
	}
	if opts.OutputFile == "" {
		return fmt.Errorf("output file is required")
	}
	return nil
}

// EstimateSize estimates the size of the source directory
func (p *Provider) EstimateSize(opts backup.Options) (int64, error) {
	ctx := context.Background()
	return kubernetes.EstimateSize(ctx, p.executor, opts.Namespace, opts.PodName, opts.SourcePath)
}

// Execute performs the backup
func (p *Provider) Execute(ctx context.Context, opts backup.Options) error {
	if err := p.ValidateOptions(opts); err != nil {
		return err
	}

	// Get output directory
	outputDir := filepath.Dir(opts.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(opts.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Build tar command (without shell redirection)
	tarCmd := p.buildTarCommand(opts)
	
	// Execute tar command and capture stdout to file
	stdout, err := p.executor.Execute(ctx, tarCmd)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	
	// Write stdout to file
	if _, err := outputFile.WriteString(stdout); err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}
	
	// Simulate progress for successful backup
	p.progressCh <- backup.Progress{
		Current: 1,
		Total:   1,
		Message: "Backup data written successfully",
	}
	
	// Send completion progress
	p.progressCh <- backup.Progress{
		Current: 100,
		Total:   100,
		Message: "Backup completed successfully",
	}
	
	return nil
}

// buildTarCommand builds the kubectl exec tar command
func (p *Provider) buildTarCommand(opts backup.Options) []string {
	args := []string{"exec", "-n", opts.Namespace, opts.PodName}
	
	// Add container if specified
	if container, ok := opts.Extra["container"].(string); ok && container != "" {
		args = append(args, "-c", container)
	}
	
	args = append(args, "--", "tar")
	
	// Add compression flag if requested
	if opts.Compress {
		args = append(args, "-czf")
	} else {
		args = append(args, "-cf")
	}
	
	// Output to stdout
	args = append(args, "-")
	
	// Add exclude patterns
	for _, exclude := range opts.Exclude {
		args = append(args, "--exclude="+exclude)
	}
	
	// Add source path
	args = append(args, "-C", "/")
	args = append(args, strings.TrimPrefix(opts.SourcePath, "/"))
	
	// Return kubectl command without shell redirection
	return kubernetes.BuildKubectlCommand(args...)
}

// monitorProgress monitors tar output and updates progress
func (p *Provider) monitorProgress(outputCh <-chan string, opts backup.Options) {
	fileCount := 0
	tarPattern := regexp.MustCompile(`^tar: (.+)$`)
	
	for line := range outputCh {
		if matches := tarPattern.FindStringSubmatch(line); matches != nil {
			fileCount++
			p.progressCh <- backup.Progress{
				Current: int64(fileCount),
				Total:   0, // Unknown total
				Message: fmt.Sprintf("Backing up: %s", matches[1]),
			}
		}
	}
}


// StreamProgress returns the progress channel
func (p *Provider) StreamProgress() <-chan backup.Progress {
	return p.progressCh
}
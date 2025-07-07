package filesystem

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

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

// EstimateSizeWithContext estimates the size with a given context
func (p *Provider) EstimateSizeWithContext(ctx context.Context, opts backup.Options) (int64, error) {
	return kubernetes.EstimateSize(ctx, p.executor, opts.Namespace, opts.PodName, opts.SourcePath)
}

// Execute performs the backup using binary-safe streaming approach
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

	// Build tar command with verbose enabled for progress monitoring
	tarCmd := p.buildTarCommand(opts)
	
	// Execute streaming tar command with binary safety
	stdout, stderr, wait, err := p.executor.StreamBinary(ctx, tarCmd)
	if err != nil {
		return fmt.Errorf("failed to start backup command: %w", err)
	}
	defer stdout.Close()
	defer stderr.Close()

	// Use WaitGroup for proper synchronization
	var wg sync.WaitGroup
	var copyErr error
	
	// Stream stdout (binary tar data) directly to file
	wg.Add(1)
	go func() {
		defer wg.Done()
		written, err := io.Copy(outputFile, stdout)
		if err != nil {
			copyErr = fmt.Errorf("failed to write backup data: %w", err)
		} else {
			// Send intermediate progress based on data written
			p.progressCh <- backup.Progress{
				Current: written,
				Total:   0, // Unknown total
				Message: fmt.Sprintf("Written %s", humanizeBytes(written)),
			}
		}
	}()
	
	// Monitor stderr for verbose progress information
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.monitorStderr(stderr, opts)
	}()
	
	// Wait for all goroutines to complete
	wg.Wait()
	
	// Wait for command completion
	if waitErr := wait(); waitErr != nil {
		return fmt.Errorf("backup command failed: %w", waitErr)
	}
	
	// Check for copy errors
	if copyErr != nil {
		return copyErr
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
	
	// Add container if specified and not empty
	if opts.Extra != nil {
		if container, ok := opts.Extra["container"].(string); ok && strings.TrimSpace(container) != "" {
			args = append(args, "-c", container)
		}
	}
	
	args = append(args, "--", "tar")
	
	// Always enable verbose for progress monitoring, compression as requested
	if opts.Compress {
		args = append(args, "-czvf")
	} else {
		args = append(args, "-cvf")
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

// monitorStderr monitors stderr for tar verbose output and updates progress
func (p *Provider) monitorStderr(stderr io.Reader, opts backup.Options) {
	fileCount := 0
	// Pattern for tar verbose output (file paths)
	verbosePattern := regexp.MustCompile(`^(.+)$`)
	// Pattern for tar error messages
	tarErrorPattern := regexp.MustCompile(`^tar: (.+)$`)
	
	// Create a buffered reader to read line by line
	scanner := bufio.NewScanner(stderr)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		// Check for tar error/warning messages first
		if matches := tarErrorPattern.FindStringSubmatch(line); matches != nil {
			p.progressCh <- backup.Progress{
				Current: int64(fileCount),
				Total:   0,
				Message: fmt.Sprintf("Tar: %s", matches[1]),
			}
		} else if verbosePattern.MatchString(line) {
			// Assume it's a file path from verbose output
			fileCount++
			p.progressCh <- backup.Progress{
				Current: int64(fileCount),
				Total:   0,
				Message: fmt.Sprintf("Backing up: %s", line),
			}
		}
	}
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

// streamWithProgress streams output to file while monitoring progress (legacy function)
func (p *Provider) streamWithProgress(outputCh <-chan string, outputFile *os.File, opts backup.Options) error {
	fileCount := 0
	// Pattern for tar verbose output (when -v flag is used)
	verbosePattern := regexp.MustCompile(`^([^/\s].*)/$|^([^/\s].*[^/])$`)
	// Pattern for tar stderr messages
	tarErrorPattern := regexp.MustCompile(`^tar: (.+)$`)
	
	for line := range outputCh {
		// Write line to output file
		if _, err := outputFile.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}
		
		// Check for progress indicators
		if verbosePattern.MatchString(line) {
			fileCount++
			p.progressCh <- backup.Progress{
				Current: int64(fileCount),
				Total:   0, // Unknown total
				Message: fmt.Sprintf("Backing up: %s", line),
			}
		} else if matches := tarErrorPattern.FindStringSubmatch(line); matches != nil {
			// Handle tar error/warning messages
			p.progressCh <- backup.Progress{
				Current: int64(fileCount),
				Total:   0,
				Message: fmt.Sprintf("Tar: %s", matches[1]),
			}
		}
	}
	
	return nil
}

// monitorProgress monitors tar output and updates progress (legacy function)
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

// Close properly cleans up resources
func (p *Provider) Close() error {
	if p.progressCh != nil {
		close(p.progressCh)
		p.progressCh = nil
	}
	return nil
}

// GetProgressChannel returns a new progress channel for this operation
func (p *Provider) GetProgressChannel() <-chan backup.Progress {
	// Create a new channel if the current one is closed
	if p.progressCh == nil {
		p.progressCh = make(chan backup.Progress, 100)
	}
	return p.progressCh
}
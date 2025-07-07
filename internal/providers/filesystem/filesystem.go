package filesystem

import (
	"context"
	"fmt"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
)

// Provider implements backup.Provider for filesystem backups
type Provider struct {
	kubeClient kubernetes.KubeClient
	executor   kubernetes.CommandExecutor
	progressCh chan backup.Progress
}

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
	// TODO: Implement backup execution
	return nil
}

// StreamProgress returns the progress channel
func (p *Provider) StreamProgress() <-chan backup.Progress {
	return p.progressCh
}
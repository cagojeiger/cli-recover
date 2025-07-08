package operation

import (
	"context"
	"fmt"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// BackupProviderAdapter adapts a backup.Provider to the unified Provider interface
type BackupProviderAdapter struct {
	provider backup.Provider
}

// NewBackupAdapter creates a new backup provider adapter
func NewBackupAdapter(provider backup.Provider) Provider {
	return &BackupProviderAdapter{provider: provider}
}

func (a *BackupProviderAdapter) Name() string {
	return a.provider.Name()
}

func (a *BackupProviderAdapter) Description() string {
	return a.provider.Description()
}

func (a *BackupProviderAdapter) Type() ProviderType {
	return TypeBackup
}

func (a *BackupProviderAdapter) ValidateOptions(opts Options) error {
	backupOpts := backup.Options{
		Namespace:  opts.Namespace,
		PodName:    opts.PodName,
		Container:  opts.Container,
		SourcePath: opts.SourcePath,
		OutputFile: opts.OutputFile,
		Compress:   opts.Compress,
		Exclude:    opts.Exclude,
		Extra:      opts.Extra,
	}
	return a.provider.ValidateOptions(backupOpts)
}

func (a *BackupProviderAdapter) Execute(ctx context.Context, opts Options) (*Result, error) {
	backupOpts := backup.Options{
		Namespace:  opts.Namespace,
		PodName:    opts.PodName,
		Container:  opts.Container,
		SourcePath: opts.SourcePath,
		OutputFile: opts.OutputFile,
		Compress:   opts.Compress,
		Exclude:    opts.Exclude,
		Extra:      opts.Extra,
	}

	err := a.provider.Execute(ctx, backupOpts)
	if err != nil {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Backup failed: %v", err),
			Error:   err,
		}, err
	}

	return &Result{
		Success: true,
		Message: "Backup completed successfully",
	}, nil
}

func (a *BackupProviderAdapter) EstimateSize(opts Options) (int64, error) {
	backupOpts := backup.Options{
		Namespace:  opts.Namespace,
		PodName:    opts.PodName,
		Container:  opts.Container,
		SourcePath: opts.SourcePath,
		Extra:      opts.Extra,
	}
	return a.provider.EstimateSize(backupOpts)
}

func (a *BackupProviderAdapter) StreamProgress() <-chan Progress {
	backupProgress := a.provider.StreamProgress()
	progress := make(chan Progress)

	go func() {
		defer close(progress)
		for bp := range backupProgress {
			progress <- Progress{
				Current: bp.Current,
				Total:   bp.Total,
				Message: bp.Message,
			}
		}
	}()

	return progress
}

// RestoreProviderAdapter adapts a restore.Provider to the unified Provider interface
type RestoreProviderAdapter struct {
	provider restore.Provider
}

// NewRestoreAdapter creates a new restore provider adapter
func NewRestoreAdapter(provider restore.Provider) Provider {
	return &RestoreProviderAdapter{provider: provider}
}

func (a *RestoreProviderAdapter) Name() string {
	return a.provider.Name()
}

func (a *RestoreProviderAdapter) Description() string {
	return a.provider.Description()
}

func (a *RestoreProviderAdapter) Type() ProviderType {
	return TypeRestore
}

func (a *RestoreProviderAdapter) ValidateOptions(opts Options) error {
	restoreOpts := restore.Options{
		Namespace:     opts.Namespace,
		PodName:       opts.PodName,
		Container:     opts.Container,
		BackupFile:    opts.BackupFile,
		TargetPath:    opts.TargetPath,
		Overwrite:     opts.Overwrite,
		PreservePerms: opts.PreservePerms,
		SkipPaths:     opts.SkipPaths,
		Extra:         opts.Extra,
	}
	return a.provider.ValidateOptions(restoreOpts)
}

func (a *RestoreProviderAdapter) Execute(ctx context.Context, opts Options) (*Result, error) {
	restoreOpts := restore.Options{
		Namespace:     opts.Namespace,
		PodName:       opts.PodName,
		Container:     opts.Container,
		BackupFile:    opts.BackupFile,
		TargetPath:    opts.TargetPath,
		Overwrite:     opts.Overwrite,
		PreservePerms: opts.PreservePerms,
		SkipPaths:     opts.SkipPaths,
		Extra:         opts.Extra,
	}

	result, err := a.provider.Execute(ctx, restoreOpts)
	if err != nil {
		return &Result{
			Success: false,
			Message: fmt.Sprintf("Restore failed: %v", err),
			Error:   err,
		}, err
	}

	return &Result{
		Success:      true,
		Message:      "Restore completed successfully",
		RestoredPath: result.RestoredPath,
		FileCount:    result.FileCount,
		BytesWritten: result.BytesWritten,
		Warnings:     result.Warnings,
	}, nil
}

func (a *RestoreProviderAdapter) EstimateSize(opts Options) (int64, error) {
	return a.provider.EstimateSize(opts.BackupFile)
}

func (a *RestoreProviderAdapter) StreamProgress() <-chan Progress {
	restoreProgress := a.provider.StreamProgress()
	progress := make(chan Progress)

	go func() {
		defer close(progress)
		for rp := range restoreProgress {
			progress <- Progress{
				Current: rp.Current,
				Total:   rp.Total,
				Message: rp.Message,
			}
		}
	}()

	return progress
}

// ConvertBackupMetadata converts from restore.Metadata to operation.Metadata
func ConvertBackupMetadata(m *restore.Metadata) *Metadata {
	if m == nil {
		return nil
	}

	return &Metadata{
		ID:           m.ID,
		Type:         m.Type,
		Provider:     m.Type, // Using Type as Provider for backward compatibility
		Namespace:    m.Namespace,
		PodName:      m.PodName,
		Container:    "", // Not available in old format
		SourcePath:   m.SourcePath,
		BackupFile:   m.BackupFile,
		Compression:  m.Compression,
		Size:         m.Size,
		Checksum:     m.Checksum,
		CreatedAt:    m.CreatedAt,
		CompletedAt:  m.CompletedAt,
		Status:       m.Status,
		ProviderInfo: m.ProviderInfo,
		Extra:        make(map[string]string),
	}
}

// ConvertToRestoreMetadata converts from operation.Metadata to restore.Metadata
func ConvertToRestoreMetadata(m *Metadata) *restore.Metadata {
	if m == nil {
		return nil
	}

	return &restore.Metadata{
		ID:           m.ID,
		Type:         m.Type,
		Namespace:    m.Namespace,
		PodName:      m.PodName,
		SourcePath:   m.SourcePath,
		BackupFile:   m.BackupFile,
		Compression:  m.Compression,
		Size:         m.Size,
		Checksum:     m.Checksum,
		CreatedAt:    m.CreatedAt,
		CompletedAt:  m.CompletedAt,
		Status:       m.Status,
		ProviderInfo: m.ProviderInfo,
	}
}

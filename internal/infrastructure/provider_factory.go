package infrastructure

import (
	"fmt"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/filesystem"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
)

// CreateBackupProvider creates a backup provider by name
func CreateBackupProvider(name string, kubeClient kubernetes.KubeClient, executor kubernetes.CommandExecutor) (backup.Provider, error) {
	switch name {
	case "filesystem":
		return filesystem.NewProvider(kubeClient, executor), nil
	// TODO: Add other providers when implemented
	// case "minio":
	//     return minio.NewProvider(kubeClient, executor), nil
	// case "mongodb":
	//     return mongodb.NewProvider(kubeClient, executor), nil
	default:
		return nil, fmt.Errorf("unknown backup provider: %s", name)
	}
}

// CreateRestoreProvider creates a restore provider by name
func CreateRestoreProvider(name string, kubeClient kubernetes.KubeClient, executor kubernetes.CommandExecutor) (restore.Provider, error) {
	switch name {
	case "filesystem":
		return filesystem.NewRestoreProvider(kubeClient, executor), nil
	// TODO: Add other providers when implemented
	// case "minio":
	//     return minio.NewRestoreProvider(kubeClient, executor), nil
	// case "mongodb":
	//     return mongodb.NewRestoreProvider(kubeClient, executor), nil
	default:
		return nil, fmt.Errorf("unknown restore provider: %s", name)
	}
}

// GetAvailableBackupProviders returns a list of available backup provider names
func GetAvailableBackupProviders() []string {
	return []string{"filesystem"}
	// TODO: Add other providers when implemented
	// return []string{"filesystem", "minio", "mongodb"}
}

// GetAvailableRestoreProviders returns a list of available restore provider names
func GetAvailableRestoreProviders() []string {
	return []string{"filesystem"}
	// TODO: Add other providers when implemented
	// return []string{"filesystem", "minio", "mongodb"}
}

package providers

import (
	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/providers/filesystem"
)

// RegisterProviders registers all available providers to the global registry
func RegisterProviders(kubeClient kubernetes.KubeClient, executor kubernetes.CommandExecutor) error {
	// Register filesystem provider
	err := backup.GlobalRegistry.RegisterFactory("filesystem", func() backup.Provider {
		return filesystem.NewProvider(kubeClient, executor)
	})
	if err != nil {
		return err
	}

	// TODO: Register MinIO provider
	// err = backup.GlobalRegistry.RegisterFactory("minio", func() backup.Provider {
	//     return minio.NewProvider(kubeClient, executor)
	// })

	// TODO: Register MongoDB provider
	// err = backup.GlobalRegistry.RegisterFactory("mongodb", func() backup.Provider {
	//     return mongodb.NewProvider(kubeClient, executor)
	// })

	return nil
}
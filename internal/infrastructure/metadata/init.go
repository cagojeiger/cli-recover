package metadata

import (
	"os"
	"path/filepath"

	"github.com/cagojeiger/cli-recover/internal/domain/metadata"
)

func init() {
	// Initialize default store
	store, err := NewFileStore("")
	if err != nil {
		// Fallback to temporary directory
		tmpDir := filepath.Join(os.TempDir(), "cli-recover-metadata")
		store, _ = NewFileStore(tmpDir)
	}
	metadata.DefaultStore = store
}
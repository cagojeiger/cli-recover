package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackupCommand_Structure(t *testing.T) {
	cmd := newBackupCommand()

	assert.Equal(t, "backup", cmd.Use)
	assert.Equal(t, "Backup resources from Kubernetes", cmd.Short)
	assert.Contains(t, cmd.Long, "Backup various types of resources")

	// Check subcommands exist
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1) // Only filesystem is implemented

	// Check filesystem subcommand
	assert.Equal(t, "filesystem", subcommands[0].Name())
}

func TestBackupProviderCmd_Filesystem_Flags(t *testing.T) {
	cmd := newProviderBackupCmd("filesystem")

	// Test command structure
	assert.Equal(t, "filesystem [pod] [path]", cmd.Use)
	assert.Equal(t, "Backup pod filesystem", cmd.Short)
	assert.Contains(t, cmd.Long, "Backup files and directories")
	assert.NotNil(t, cmd.RunE)

	// Test filesystem-specific flags
	flags := cmd.Flags()

	namespaceFlag := flags.Lookup("namespace")
	assert.NotNil(t, namespaceFlag)
	assert.Equal(t, "default", namespaceFlag.DefValue)

	compressionFlag := flags.Lookup("compression")
	assert.NotNil(t, compressionFlag)
	assert.Equal(t, "none", compressionFlag.DefValue)

	excludeFlag := flags.Lookup("exclude")
	assert.NotNil(t, excludeFlag)
	assert.Equal(t, "stringSlice", excludeFlag.Value.Type())

	verboseFlag := flags.Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)

	outputFlag := flags.Lookup("output")
	assert.NotNil(t, outputFlag)
}

func TestBackupProviderCmd_UnknownProvider(t *testing.T) {
	// Test that unknown provider returns nil
	cmd := newProviderBackupCmd("unknown")
	assert.Nil(t, cmd)
}

func TestBackupCommand_ShowsHelp(t *testing.T) {
	cmd := newBackupCommand()

	// When run without subcommand, it should show help
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err) // Help() returns nil
}

func TestBackupProviderCmd_FilesystemArgs(t *testing.T) {
	cmd := newProviderBackupCmd("filesystem")

	// Test with correct number of args (ExactArgs(2) is used in backup.go)
	err := cmd.Args(cmd, []string{"pod-name", "/path/to/backup"})
	assert.NoError(t, err)

	// Test with too few args
	err = cmd.Args(cmd, []string{"pod-name"})
	assert.Error(t, err)

	// Test with no args
	err = cmd.Args(cmd, []string{})
	assert.Error(t, err)

	// Test with too many args
	err = cmd.Args(cmd, []string{"pod-name", "/path", "extra"})
	assert.Error(t, err) // ExactArgs doesn't allow extra
}

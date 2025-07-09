package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRestoreCommand_Structure(t *testing.T) {
	cmd := newRestoreCommand()

	assert.Equal(t, "restore", cmd.Use)
	assert.Equal(t, "Restore resources to Kubernetes", cmd.Short)
	assert.Contains(t, cmd.Long, "Restore various types of backups")

	// Check subcommands exist
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1) // Only filesystem is implemented

	// Check filesystem subcommand
	assert.Equal(t, "filesystem", subcommands[0].Name())
}

func TestRestoreProviderCmd_Filesystem_Flags(t *testing.T) {
	cmd := newProviderRestoreCmd("filesystem")

	// Test command structure
	assert.Equal(t, "filesystem [pod] [backup-file]", cmd.Use)
	assert.Equal(t, "Restore pod filesystem from backup", cmd.Short)
	assert.Contains(t, cmd.Long, "Restore files and directories")
	assert.NotNil(t, cmd.RunE)

	// Test filesystem-specific flags
	flags := cmd.Flags()

	namespaceFlag := flags.Lookup("namespace")
	assert.NotNil(t, namespaceFlag)
	assert.Equal(t, "default", namespaceFlag.DefValue)

	targetPathFlag := flags.Lookup("target-path")
	assert.NotNil(t, targetPathFlag)
	assert.Equal(t, "/", targetPathFlag.DefValue)

	forceFlag := flags.Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)

	preservePermsFlag := flags.Lookup("preserve-perms")
	assert.NotNil(t, preservePermsFlag)
	assert.Equal(t, "false", preservePermsFlag.DefValue)

	skipPathsFlag := flags.Lookup("skip-paths")
	assert.NotNil(t, skipPathsFlag)
	assert.Equal(t, "stringSlice", skipPathsFlag.Value.Type())

	containerFlag := flags.Lookup("container")
	assert.NotNil(t, containerFlag)

	verboseFlag := flags.Lookup("verbose")
	assert.NotNil(t, verboseFlag)

	dryRunFlag := flags.Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)
}

func TestRestoreProviderCmd_UnknownProvider(t *testing.T) {
	// Test that unknown provider returns nil
	cmd := newProviderRestoreCmd("unknown")
	assert.Nil(t, cmd)
}

func TestRestoreCommand_ShowsHelp(t *testing.T) {
	cmd := newRestoreCommand()

	// When run without subcommand, it should show help
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err) // Help() returns nil
}

func TestRestoreProviderCmd_FilesystemArgs(t *testing.T) {
	cmd := newProviderRestoreCmd("filesystem")

	// Test with correct number of args (ExactArgs(2) is used in restore.go)
	err := cmd.Args(cmd, []string{"pod-name", "/backup.tar"})
	assert.NoError(t, err)

	// Test with too few args
	err = cmd.Args(cmd, []string{"pod-name"})
	assert.Error(t, err)

	// Test with no args
	err = cmd.Args(cmd, []string{})
	assert.Error(t, err)

	// Test with too many args
	err = cmd.Args(cmd, []string{"pod-name", "/backup.tar", "extra"})
	assert.Error(t, err) // ExactArgs doesn't allow extra
}

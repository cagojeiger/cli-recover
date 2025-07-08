package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommand_Structure(t *testing.T) {
	cmd := newListCommand()

	// Test command structure
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List backups and other resources", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Test subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1)
	assert.Equal(t, "backups", subcommands[0].Use)
}

func TestNewListBackupsCommand(t *testing.T) {
	cmd := newListBackupsCommand()

	// Test command structure
	assert.Equal(t, "backups", cmd.Use)
	assert.Equal(t, "List all backups", cmd.Short)
	assert.Contains(t, cmd.Long, "List all backups stored")
	assert.NotNil(t, cmd.RunE)

	// Test flags exist
	flags := cmd.Flags()

	// Check flags
	namespaceFlag := flags.Lookup("namespace")
	assert.NotNil(t, namespaceFlag)
	assert.Equal(t, "", namespaceFlag.DefValue) // Empty means all namespaces

	outputFlag := flags.Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "table", outputFlag.DefValue)

	detailsFlag := flags.Lookup("details")
	assert.NotNil(t, detailsFlag)
}

func TestListCommand_ShowsHelpWhenNoSubcommand(t *testing.T) {
	cmd := newListCommand()

	// Execute without subcommand should show help (no error expected as help is shown)
	err := cmd.RunE(cmd, []string{})

	// The RunE should return cmd.Help() which should not return an error
	assert.NoError(t, err)
}

func TestListBackupsCommand_NoArgs(t *testing.T) {
	cmd := newListBackupsCommand()

	// Test that no args are required (Args should be nil or allow 0 args)
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		assert.NoError(t, err)
	}
}

func TestListCommand_FlagTypes(t *testing.T) {
	cmd := newListBackupsCommand()
	flags := cmd.Flags()

	// Test namespace flag type
	namespaceFlag := flags.Lookup("namespace")
	assert.Equal(t, "string", namespaceFlag.Value.Type())

	// Test output flag type
	outputFlag := flags.Lookup("output")
	assert.Equal(t, "string", outputFlag.Value.Type())

	// Test boolean flags
	detailsFlag := flags.Lookup("details")
	assert.Equal(t, "bool", detailsFlag.Value.Type())
}

func TestListCommand_FlagShortcuts(t *testing.T) {
	cmd := newListBackupsCommand()
	flags := cmd.Flags()

	// Test namespace shortcut
	namespaceFlag := flags.Lookup("namespace")
	assert.Equal(t, "n", namespaceFlag.Shorthand)

	// Test output shortcut
	outputFlag := flags.Lookup("output")
	assert.Equal(t, "o", outputFlag.Shorthand)
}

func TestListCommand_DefaultOutputFormat(t *testing.T) {
	cmd := newListBackupsCommand()
	flags := cmd.Flags()

	outputFlag := flags.Lookup("output")
	assert.Equal(t, "table", outputFlag.DefValue)
}

func TestListCommand_FlagCount(t *testing.T) {
	cmd := newListBackupsCommand()
	flags := cmd.Flags()

	// Check that specific flags exist (NFlag() returns 0 until parsed)
	assert.NotNil(t, flags.Lookup("namespace"))
	assert.NotNil(t, flags.Lookup("output"))
	assert.NotNil(t, flags.Lookup("details"))
}

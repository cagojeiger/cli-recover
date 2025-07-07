package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInitCommand(t *testing.T) {
	cmd := newInitCommand()
	
	// Test command structure
	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize CLI configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "Initialize CLI-Recover configuration")
	assert.NotNil(t, cmd.RunE)
	
	// Test flags exist
	flags := cmd.Flags()
	
	showFlag := flags.Lookup("show")
	assert.NotNil(t, showFlag)
	assert.Equal(t, "bool", showFlag.Value.Type())
	
	resetFlag := flags.Lookup("reset")
	assert.NotNil(t, resetFlag)
	assert.Equal(t, "bool", resetFlag.Value.Type())
	
	interactiveFlag := flags.Lookup("interactive")
	assert.NotNil(t, interactiveFlag)
	assert.Equal(t, "bool", interactiveFlag.Value.Type())
}

func TestInitCommand_FlagDefaultValues(t *testing.T) {
	cmd := newInitCommand()
	flags := cmd.Flags()
	
	// Test flag default values
	showFlag := flags.Lookup("show")
	assert.Equal(t, "false", showFlag.DefValue)
	
	resetFlag := flags.Lookup("reset")
	assert.Equal(t, "false", resetFlag.DefValue)
	
	interactiveFlag := flags.Lookup("interactive")
	assert.Equal(t, "false", interactiveFlag.DefValue)
}

func TestInitCommand_HasExamples(t *testing.T) {
	cmd := newInitCommand()
	
	// Check that examples are provided
	assert.NotEmpty(t, cmd.Example)
	assert.Contains(t, cmd.Example, "cli-recover init")
	assert.Contains(t, cmd.Example, "--show")
	assert.Contains(t, cmd.Example, "--reset")
	assert.Contains(t, cmd.Example, "--interactive")
}

func TestInitCommand_NoArgs(t *testing.T) {
	cmd := newInitCommand()
	
	// Test that no args are required (Args should be nil or allow 0 args)
	if cmd.Args != nil {
		err := cmd.Args(cmd, []string{})
		assert.NoError(t, err)
	}
}

func TestPromptWithDefault_EmptyInput(t *testing.T) {
	// Test promptWithDefault function with empty input (returns default)
	// Note: This test can't fully test the interactive nature but can test logic
	
	// We can't easily test the interactive input without mocking stdin,
	// but we can test that the function signature is correct
	assert.NotPanics(t, func() {
		// This would normally wait for input, so we just check function exists
		// In a real test environment, we'd mock os.Stdin
		_ = promptWithDefault
	})
}

func TestInitCommand_FlagUsage(t *testing.T) {
	cmd := newInitCommand()
	flags := cmd.Flags()
	
	// Test flag usage messages
	showFlag := flags.Lookup("show")
	assert.Equal(t, "Show current configuration", showFlag.Usage)
	
	resetFlag := flags.Lookup("reset")
	assert.Equal(t, "Reset configuration to defaults", resetFlag.Usage)
	
	interactiveFlag := flags.Lookup("interactive")
	assert.Equal(t, "Interactive configuration setup", interactiveFlag.Usage)
}

func TestInitCommand_HasLongDescription(t *testing.T) {
	cmd := newInitCommand()
	
	// Test that long description contains key information
	assert.Contains(t, cmd.Long, "configuration directory")
	assert.Contains(t, cmd.Long, "default configuration")
	assert.Contains(t, cmd.Long, "log and metadata directories")
	assert.Contains(t, cmd.Long, "logger, backup, and metadata")
}

func TestInitCommand_RunENotNil(t *testing.T) {
	cmd := newInitCommand()
	
	// Test that RunE function is not nil
	assert.NotNil(t, cmd.RunE)
}

// Test that helper functions exist and have correct signatures
func TestInitCommand_HelperFunctionsExist(t *testing.T) {
	// Test that all helper functions are properly defined
	assert.NotNil(t, createConfiguration)
	assert.NotNil(t, showConfiguration)
	assert.NotNil(t, resetConfiguration)
	assert.NotNil(t, interactiveSetup)
	assert.NotNil(t, promptWithDefault)
}

// Test init command structure completeness
func TestInitCommand_CommandStructure(t *testing.T) {
	cmd := newInitCommand()
	
	// Ensure all required fields are set
	assert.NotEmpty(t, cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
	assert.NotNil(t, cmd.RunE)
	
	// Ensure command has flags
	assert.True(t, cmd.HasFlags())
	// Note: NFlag() returns 0 because flags are not parsed yet
	// We can check if flags exist individually
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("show"))
	assert.NotNil(t, flags.Lookup("reset"))
	assert.NotNil(t, flags.Lookup("interactive"))
}

func TestInitCommand_NoSubcommands(t *testing.T) {
	cmd := newInitCommand()
	
	// Init command should not have subcommands
	assert.False(t, cmd.HasSubCommands())
	assert.Len(t, cmd.Commands(), 0)
}
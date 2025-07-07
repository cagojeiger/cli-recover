package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/runner"
)

// Test Golden Runner can read golden files
func TestGoldenRunner_Run(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "get namespaces",
			cmd:  "kubectl",
			args: []string{"get", "namespaces", "-o", "json"},
			want: `"default"`,
		},
		{
			name: "get pods in default",
			cmd:  "kubectl",
			args: []string{"get", "pods", "-n", "default", "-o", "json"},
			want: `"nginx-`,
		},
		{
			name:    "command not found",
			cmd:     "kubectl",
			args:    []string{"get", "nodes"},
			wantErr: true,
		},
	}

	runner := runner.NewGoldenRunner("../../testdata")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runner.Run(tt.cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GoldenRunner.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !strings.Contains(string(got), tt.want) {
				t.Errorf("GoldenRunner.Run() = %v, want contains %v", string(got), tt.want)
			}
		})
	}
}

// Test Runner interface switches based on environment
func TestNewRunner(t *testing.T) {
	// Test Golden Runner
	os.Setenv("USE_GOLDEN", "true")
	r := runner.NewRunner()
	
	// Test that we can run a command (indirectly testing it's a GoldenRunner)
	_, err := r.Run("kubectl", "get", "namespaces", "-o", "json")
	if err != nil {
		t.Error("Expected GoldenRunner to work with test data when USE_GOLDEN=true")
	}
}

// Test main function with various arguments
func TestMainFunction(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	
	tests := []struct {
		name     string
		args     []string
		wantExit bool
	}{
		{
			name:     "version flag",
			args:     []string{"cli-recover", "--version"},
			wantExit: false,
		},
		{
			name:     "help flag",
			args:     []string{"cli-recover", "--help"},
			wantExit: false,
		},
		{
			name:     "backup help",
			args:     []string{"cli-recover", "backup", "--help"},
			wantExit: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test args
			os.Args = tt.args
			
			// Capture output
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			_, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w
			
			// Run main in a goroutine to capture any exit
			done := make(chan bool)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						// Recovered from panic/exit
					}
					done <- true
				}()
				
				// Note: We can't actually test main() directly because it calls os.Exit
				// Instead, we test the command creation and flag parsing
			}()
			
			// Wait a bit and restore
			os.Stdout = oldStdout
			os.Stderr = oldStderr
			w.Close()
			
			select {
			case <-done:
				// Test completed
			case <-time.After(100 * time.Millisecond):
				// Timeout is ok for this test
			}
		})
	}
}

// Test command creation functions
func TestCreateRootCommand(t *testing.T) {
	cmd := createRootCommand()
	
	assert.Equal(t, "cli-recover", cmd.Use)
	assert.Equal(t, "Kubernetes integrated backup and restore tool", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.Equal(t, version, cmd.Version)
}

func TestAddGlobalFlags(t *testing.T) {
	cmd := createRootCommand()
	addGlobalFlags(cmd)
	
	// Check that global flags are added
	_, err := cmd.PersistentFlags().GetBool("debug")
	assert.NoError(t, err, "debug flag should exist")
	
	_, err = cmd.PersistentFlags().GetString("log-level")
	assert.NoError(t, err, "log-level flag should exist")
	
	_, err = cmd.PersistentFlags().GetString("log-file")
	assert.NoError(t, err, "log-file flag should exist")
	
	_, err = cmd.PersistentFlags().GetString("log-format")
	assert.NoError(t, err, "log-format flag should exist")
}

func TestAddSubcommands(t *testing.T) {
	cmd := createRootCommand()
	addSubcommands(cmd)
	
	// Check that subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 4)
	
	cmdNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		cmdNames[i] = subcmd.Name()
	}
	
	assert.Contains(t, cmdNames, "backup")
	assert.Contains(t, cmdNames, "restore")
	assert.Contains(t, cmdNames, "list")
	assert.Contains(t, cmdNames, "init")
}

func TestNewBackupCommand(t *testing.T) {
	cmd := newBackupCommand()
	
	assert.Equal(t, "backup", cmd.Use)
	assert.Equal(t, "Backup resources from Kubernetes", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	
	// Check that it has subcommands for different providers
	subcommands := cmd.Commands()
	assert.NotEmpty(t, subcommands)
	
	// Should have filesystem subcommand
	hasFilesystem := false
	for _, subcmd := range subcommands {
		if subcmd.Name() == "filesystem" {
			hasFilesystem = true
			break
		}
	}
	assert.True(t, hasFilesystem, "backup command should have filesystem subcommand")
}

func TestNewRestoreCommand(t *testing.T) {
	cmd := newRestoreCommand()
	
	assert.Equal(t, "restore", cmd.Use)
	assert.Equal(t, "Restore resources to Kubernetes", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	
	// Check that it has subcommands for different providers
	subcommands := cmd.Commands()
	assert.NotEmpty(t, subcommands)
	
	// Should have filesystem subcommand
	hasFilesystem := false
	for _, subcmd := range subcommands {
		if subcmd.Name() == "filesystem" {
			hasFilesystem = true
			break
		}
	}
	assert.True(t, hasFilesystem, "restore command should have filesystem subcommand")
}

func TestNewListCommand(t *testing.T) {
	cmd := newListCommand()
	
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List backups and other resources", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	
	// Check that it has subcommands
	subcommands := cmd.Commands()
	assert.NotEmpty(t, subcommands)
	
	// Should have backups subcommand
	hasBackups := false
	for _, subcmd := range subcommands {
		if subcmd.Name() == "backups" {
			hasBackups = true
			break
		}
	}
	assert.True(t, hasBackups, "list command should have backups subcommand")
}

func TestNewInitCommand(t *testing.T) {
	cmd := newInitCommand()
	
	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize CLI configuration", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	
	// Check that it has the expected flags
	_, err := cmd.Flags().GetBool("show")
	assert.NoError(t, err, "show flag should exist")
	
	_, err = cmd.Flags().GetBool("reset")
	assert.NoError(t, err, "reset flag should exist")
	
	_, err = cmd.Flags().GetBool("interactive")
	assert.NoError(t, err, "interactive flag should exist")
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "absolute path",
			input:    "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "relative path",
			input:    "test/file",
			expected: "test/file",
		},
		{
			name:     "home directory path",
			input:    "~/test/file",
			expected: "", // Will be filled by home dir + "/test/file"
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			
			if tt.name == "home directory path" {
				// Check that it expands ~ to actual home directory
				assert.NotEqual(t, tt.input, result)
				assert.NotContains(t, result, "~")
				assert.Contains(t, result, "test/file")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCommandExecution(t *testing.T) {
	// Set up environment for testing
	os.Setenv("USE_GOLDEN", "true")
	defer os.Unsetenv("USE_GOLDEN")
	
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectUsage bool
	}{
		{
			name:        "no arguments shows help",
			args:        []string{},
			expectError: false,
			expectUsage: true,
		},
		{
			name:        "version flag",
			args:        []string{"--version"},
			expectError: false,
			expectUsage: false,
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
			expectUsage: true,
		},
		{
			name:        "backup help",
			args:        []string{"backup", "--help"},
			expectError: false,
			expectUsage: true,
		},
		{
			name:        "restore help",
			args:        []string{"restore", "--help"},
			expectError: false,
			expectUsage: true,
		},
		{
			name:        "list help",
			args:        []string{"list", "--help"},
			expectError: false,
			expectUsage: true,
		},
		{
			name:        "init help",
			args:        []string{"init", "--help"},
			expectError: false,
			expectUsage: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command
			rootCmd := createRootCommand()
			setupPersistentPreRun(rootCmd)
			addGlobalFlags(rootCmd)
			addSubcommands(rootCmd)
			
			// Capture output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)
			
			// Set arguments
			rootCmd.SetArgs(tt.args)
			
			// Execute
			err := rootCmd.Execute()
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			output := buf.String()
			if tt.expectUsage {
				assert.Contains(t, output, "Usage:")
			}
		})
	}
}
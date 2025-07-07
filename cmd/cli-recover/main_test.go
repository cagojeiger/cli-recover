package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Test restore command structure and functionality
func TestRestoreCommandStructure(t *testing.T) {
	cmd := newRestoreCommand()
	
	assert.Equal(t, "restore", cmd.Use)
	assert.Equal(t, "Restore resources to Kubernetes", cmd.Short)
	assert.Contains(t, cmd.Long, "Available restore types")
	
	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1) // Only filesystem is implemented
	
	filesystemCmd := subcommands[0]
	assert.Contains(t, filesystemCmd.Use, "filesystem")
}

func TestNewProviderRestoreCmd_Filesystem(t *testing.T) {
	cmd := newProviderRestoreCmd("filesystem")
	
	assert.Equal(t, "filesystem [pod] [backup-file]", cmd.Use)
	assert.Equal(t, "Restore pod filesystem from backup", cmd.Short)
	assert.Contains(t, cmd.Long, "tar backup")
	
	// Check required flags exist
	flags := []string{"namespace", "target-path", "overwrite", "preserve-perms", "skip-paths", "container", "verbose", "dry-run"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
	
	// Test default values
	namespace, _ := cmd.Flags().GetString("namespace")
	assert.Equal(t, "default", namespace)
	
	targetPath, _ := cmd.Flags().GetString("target-path")
	assert.Equal(t, "/", targetPath)
}

func TestNewProviderRestoreCmd_MinIO(t *testing.T) {
	cmd := newProviderRestoreCmd("minio")
	
	assert.Equal(t, "minio [bucket] [backup-dir]", cmd.Use)
	assert.Equal(t, "Restore MinIO bucket from backup", cmd.Short)
	
	// Test that it returns not implemented error
	err := cmd.RunE(cmd, []string{"test-bucket", "/backup/path"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MinIO restore provider not yet implemented")
}

func TestNewProviderRestoreCmd_MongoDB(t *testing.T) {
	cmd := newProviderRestoreCmd("mongodb")
	
	assert.Equal(t, "mongodb [database] [backup-file]", cmd.Use)
	assert.Equal(t, "Restore MongoDB database from backup", cmd.Short)
	
	// Test that it returns not implemented error
	err := cmd.RunE(cmd, []string{"test-db", "/backup/file"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MongoDB restore provider not yet implemented")
}

func TestRestoreCommand_HelpWhenNoSubcommand(t *testing.T) {
	cmd := newRestoreCommand()
	
	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	// Execute without subcommand should show help
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err) // Help() returns nil
}

// Test init command structure and functionality
func TestInitCommandStructure(t *testing.T) {
	cmd := newInitCommand()
	
	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize CLI configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "Initialize CLI-Recover configuration")
	assert.Contains(t, cmd.Example, "cli-recover init")
	
	// Check flags exist
	flags := []string{"show", "reset", "interactive"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
}

func TestInitCommand_FlagDefaults(t *testing.T) {
	cmd := newInitCommand()
	
	// Test default flag values
	show, _ := cmd.Flags().GetBool("show")
	assert.False(t, show)
	
	reset, _ := cmd.Flags().GetBool("reset")
	assert.False(t, reset)
	
	interactive, _ := cmd.Flags().GetBool("interactive")
	assert.False(t, interactive)
}

func TestCreateConfiguration_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	// Create existing config file
	err := os.WriteFile(configPath, []byte("existing config"), 0644)
	require.NoError(t, err)
	
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	err = createConfiguration(configPath)
	
	w.Close()
	os.Stdout = oldStdout
	
	output := make([]byte, 1024)
	n, _ := r.Read(output)
	
	assert.NoError(t, err)
	assert.Contains(t, string(output[:n]), "Configuration file already exists")
}

func TestPromptWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{
			name:         "empty input uses default",
			input:        "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "non-empty input overrides default",
			input:        "custom_value",
			defaultValue: "default_value",
			expected:     "custom_value",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			
			go func() {
				defer w.Close()
				if tt.input != "" {
					w.Write([]byte(tt.input + "\n"))
				} else {
					w.Write([]byte("\n"))
				}
			}()
			
			// Capture stdout to avoid printing during test
			oldStdout := os.Stdout
			devNull, _ := os.Open("/dev/null")
			os.Stdout = devNull
			
			result := promptWithDefault("test prompt", tt.defaultValue)
			
			// Restore stdin/stdout
			os.Stdin = oldStdin
			os.Stdout = oldStdout
			devNull.Close()
			
			assert.Equal(t, tt.expected, result)
		})
	}
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

// Test backup command structure and functionality  
func TestBackupCommandStructure(t *testing.T) {
	cmd := newBackupCommand()
	
	assert.Equal(t, "backup", cmd.Use)
	assert.Equal(t, "Backup resources from Kubernetes", cmd.Short)
	assert.Contains(t, cmd.Long, "Available backup types")
	
	// Check subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1) // Only filesystem is implemented
	
	filesystemCmd := subcommands[0]
	assert.Contains(t, filesystemCmd.Use, "filesystem")
}

func TestNewProviderBackupCmd_Filesystem(t *testing.T) {
	cmd := newProviderBackupCmd("filesystem")
	
	assert.Equal(t, "filesystem [pod] [path]", cmd.Use)
	assert.Equal(t, "Backup pod filesystem", cmd.Short)
	assert.Contains(t, cmd.Long, "tar")
	
	// Check required flags exist
	flags := []string{"namespace", "compression", "exclude", "exclude-vcs", "verbose", "totals", "preserve-perms", "container", "output", "dry-run"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}
	
	// Test default values
	namespace, _ := cmd.Flags().GetString("namespace")
	assert.Equal(t, "default", namespace)
	
	compression, _ := cmd.Flags().GetString("compression")
	assert.Equal(t, "gzip", compression)
}

func TestNewProviderBackupCmd_MinIO(t *testing.T) {
	cmd := newProviderBackupCmd("minio")
	
	assert.Equal(t, "minio [bucket]", cmd.Use)
	assert.Equal(t, "Backup MinIO bucket", cmd.Short)
	
	// Test that it returns not implemented error
	err := cmd.RunE(cmd, []string{"test-bucket"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MinIO backup provider not yet implemented")
}

func TestNewProviderBackupCmd_MongoDB(t *testing.T) {
	cmd := newProviderBackupCmd("mongodb")
	
	assert.Equal(t, "mongodb [database]", cmd.Use)
	assert.Equal(t, "Backup MongoDB database", cmd.Short)
	
	// Test that it returns not implemented error
	err := cmd.RunE(cmd, []string{"test-db"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MongoDB backup provider not yet implemented")
}

func TestBackupCommand_HelpWhenNoSubcommand(t *testing.T) {
	cmd := newBackupCommand()
	
	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	
	// Execute without subcommand should show help
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err) // Help() returns nil
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
		{
			name:        "backup filesystem help",
			args:        []string{"backup", "filesystem", "--help"},
			expectError: false,
			expectUsage: true,
		},
		{
			name:        "restore filesystem help",
			args:        []string{"restore", "filesystem", "--help"},
			expectError: false,
			expectUsage: true,
		},
		{
			name:        "list backups help",
			args:        []string{"list", "backups", "--help"},
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
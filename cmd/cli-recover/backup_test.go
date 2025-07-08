package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackupCommand_Structure(t *testing.T) {
	cmd := newBackupCommand()
	
	// Test command structure
	assert.Equal(t, "backup", cmd.Use)
	assert.Equal(t, "Backup resources from Kubernetes", cmd.Short)
	assert.Contains(t, cmd.Long, "Available backup types")
	assert.NotNil(t, cmd.RunE)
	
	// Test subcommands
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 1)
	assert.Equal(t, "filesystem [pod] [path]", subcommands[0].Use)
}

func TestBackupProviderCmd_Filesystem_Flags(t *testing.T) {
	cmd := newProviderBackupCmd("filesystem")
	
	// Test command structure
	assert.Equal(t, "filesystem [pod] [path]", cmd.Use)
	assert.Equal(t, "Backup pod filesystem", cmd.Short)
	assert.Contains(t, cmd.Long, "Backup files and directories")
	assert.NotNil(t, cmd.RunE)
	
	// Test required flags exist
	flags := cmd.Flags()
	
	// Check required flags
	namespaceFlag := flags.Lookup("namespace")
	assert.NotNil(t, namespaceFlag)
	assert.Equal(t, "default", namespaceFlag.DefValue)
	
	compressionFlag := flags.Lookup("compression")
	assert.NotNil(t, compressionFlag)
	assert.Equal(t, "none", compressionFlag.DefValue)
	
	excludeFlag := flags.Lookup("exclude")
	assert.NotNil(t, excludeFlag)
	
	verboseFlag := flags.Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	
	dryRunFlag := flags.Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)
	
	outputFlag := flags.Lookup("output")
	assert.NotNil(t, outputFlag)
}

func TestBackupProviderCmd_Minio_Flags(t *testing.T) {
	cmd := newProviderBackupCmd("minio")
	
	// Test command structure
	assert.Equal(t, "minio [bucket]", cmd.Use)
	assert.Equal(t, "Backup MinIO bucket", cmd.Short)
	assert.Contains(t, cmd.Long, "Backup MinIO bucket contents")
	assert.NotNil(t, cmd.RunE)
	
	// Test MinIO-specific flags
	flags := cmd.Flags()
	
	serviceFlag := flags.Lookup("service")
	assert.NotNil(t, serviceFlag)
	
	accessKeyFlag := flags.Lookup("access-key")
	assert.NotNil(t, accessKeyFlag)
	
	secretKeyFlag := flags.Lookup("secret-key")
	assert.NotNil(t, secretKeyFlag)
}

func TestBackupProviderCmd_MongoDB_Flags(t *testing.T) {
	cmd := newProviderBackupCmd("mongodb")
	
	// Test command structure
	assert.Equal(t, "mongodb [database]", cmd.Use)
	assert.Equal(t, "Backup MongoDB database", cmd.Short)
	assert.Contains(t, cmd.Long, "Backup MongoDB database using mongodump")
	assert.NotNil(t, cmd.RunE)
	
	// Test MongoDB-specific flags
	flags := cmd.Flags()
	
	podFlag := flags.Lookup("pod")
	assert.NotNil(t, podFlag)
	
	uriFlag := flags.Lookup("uri")
	assert.NotNil(t, uriFlag)
	
	collectionsFlag := flags.Lookup("collections")
	assert.NotNil(t, collectionsFlag)
}

func TestBackupProviderCmd_UnknownProvider(t *testing.T) {
	cmd := newProviderBackupCmd("unknown")
	
	// Should return nil for unknown providers
	assert.Nil(t, cmd)
}

func TestBackupCommand_ShowsHelp(t *testing.T) {
	cmd := newBackupCommand()
	
	// Execute without subcommand should show help (no error expected as help is shown)
	err := cmd.RunE(cmd, []string{})
	
	// The RunE should return cmd.Help() which should not return an error
	assert.NoError(t, err)
}

func TestBackupProviderCmd_FilesystemArgs(t *testing.T) {
	cmd := newProviderBackupCmd("filesystem")
	
	// Test args validation by calling Args function directly
	assert.NotNil(t, cmd.Args)
	
	// Test with correct number of args (simulated)
	err := cmd.Args(cmd, []string{"pod-name", "/path"})
	assert.NoError(t, err)
	
	// Test with incorrect number of args (simulated)
	err = cmd.Args(cmd, []string{"pod-name"})
	assert.Error(t, err)
	
	err = cmd.Args(cmd, []string{})
	assert.Error(t, err)
	
	err = cmd.Args(cmd, []string{"pod-name", "/path", "extra"})
	assert.Error(t, err)
}

func TestBackupProviderCmd_MinioArgs(t *testing.T) {
	cmd := newProviderBackupCmd("minio")
	
	// Test args validation
	assert.NotNil(t, cmd.Args)
	
	// Test with correct args (simulated)
	err := cmd.Args(cmd, []string{"bucket-name"})
	assert.NoError(t, err)
	
	err = cmd.Args(cmd, []string{"bucket-name", "extra"})
	assert.NoError(t, err)
	
	// Test with no args (simulated)
	err = cmd.Args(cmd, []string{})
	assert.Error(t, err)
}

func TestBackupProviderCmd_MongoDBArgs(t *testing.T) {
	cmd := newProviderBackupCmd("mongodb")
	
	// Test args validation
	assert.NotNil(t, cmd.Args)
	
	// Test with correct args (simulated)
	err := cmd.Args(cmd, []string{"database-name"})
	assert.NoError(t, err)
	
	err = cmd.Args(cmd, []string{"database-name", "collection"})
	assert.NoError(t, err)
	
	// Test with no args (simulated)
	err = cmd.Args(cmd, []string{})
	assert.Error(t, err)
}
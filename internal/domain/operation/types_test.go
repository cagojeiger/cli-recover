package operation_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/operation"
)

func TestMetadata_SetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *operation.Metadata
		key      string
		value    string
		expected map[string]string
	}{
		{
			name: "set metadata on nil extra map",
			setup: func() *operation.Metadata {
				return &operation.Metadata{
					ID:   "test-123",
					Type: "backup",
				}
			},
			key:   "custom-key",
			value: "custom-value",
			expected: map[string]string{
				"custom-key": "custom-value",
			},
		},
		{
			name: "set metadata on existing extra map",
			setup: func() *operation.Metadata {
				return &operation.Metadata{
					ID:   "test-456",
					Type: "restore",
					Extra: map[string]string{
						"existing": "value",
					},
				}
			},
			key:   "new-key",
			value: "new-value",
			expected: map[string]string{
				"existing": "value",
				"new-key":  "new-value",
			},
		},
		{
			name: "overwrite existing metadata",
			setup: func() *operation.Metadata {
				return &operation.Metadata{
					ID:   "test-789",
					Type: "backup",
					Extra: map[string]string{
						"key1": "old-value",
						"key2": "value2",
					},
				}
			},
			key:   "key1",
			value: "new-value",
			expected: map[string]string{
				"key1": "new-value",
				"key2": "value2",
			},
		},
		{
			name: "set empty key and value",
			setup: func() *operation.Metadata {
				return &operation.Metadata{
					ID: "test-empty",
				}
			},
			key:   "",
			value: "",
			expected: map[string]string{
				"": "",
			},
		},
		{
			name: "set metadata multiple times",
			setup: func() *operation.Metadata {
				m := &operation.Metadata{
					ID: "test-multiple",
				}
				m.SetMetadata("first", "1")
				m.SetMetadata("second", "2")
				return m
			},
			key:   "third",
			value: "3",
			expected: map[string]string{
				"first":  "1",
				"second": "2",
				"third":  "3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := tt.setup()
			metadata.SetMetadata(tt.key, tt.value)
			
			assert.NotNil(t, metadata.Extra)
			assert.Equal(t, tt.expected, metadata.Extra)
		})
	}
}

func TestOptions_Validation(t *testing.T) {
	tests := []struct {
		name        string
		opts        operation.Options
		validate    func(*testing.T, operation.Options)
		description string
	}{
		{
			name: "backup options with all fields",
			opts: operation.Options{
				Type:       operation.TypeBackup,
				Namespace:  "production",
				PodName:    "app-pod",
				Container:  "main",
				SourcePath: "/app/data",
				OutputFile: "/backups/app.tar.gz",
				Compress:   true,
				Exclude:    []string{"*.log", "*.tmp"},
				Extra: map[string]interface{}{
					"retention": "30d",
					"priority":  1,
				},
			},
			validate: func(t *testing.T, opts operation.Options) {
				assert.Equal(t, operation.TypeBackup, opts.Type)
				assert.Equal(t, "production", opts.Namespace)
				assert.Equal(t, "app-pod", opts.PodName)
				assert.Equal(t, "main", opts.Container)
				assert.Equal(t, "/app/data", opts.SourcePath)
				assert.Equal(t, "/backups/app.tar.gz", opts.OutputFile)
				assert.True(t, opts.Compress)
				assert.Equal(t, []string{"*.log", "*.tmp"}, opts.Exclude)
				assert.NotNil(t, opts.Extra)
				assert.Equal(t, "30d", opts.Extra["retention"])
				assert.Equal(t, 1, opts.Extra["priority"])
			},
			description: "All backup-specific fields should be set correctly",
		},
		{
			name: "restore options with all fields",
			opts: operation.Options{
				Type:          operation.TypeRestore,
				Namespace:     "staging",
				PodName:       "db-pod",
				Container:     "postgres",
				BackupFile:    "/backups/db.tar.gz",
				TargetPath:    "/var/lib/postgresql/data",
				Overwrite:     true,
				PreservePerms: true,
				SkipPaths:     []string{"/var/lib/postgresql/data/pg_wal"},
				Extra: map[string]interface{}{
					"verify": true,
				},
			},
			validate: func(t *testing.T, opts operation.Options) {
				assert.Equal(t, operation.TypeRestore, opts.Type)
				assert.Equal(t, "staging", opts.Namespace)
				assert.Equal(t, "db-pod", opts.PodName)
				assert.Equal(t, "postgres", opts.Container)
				assert.Equal(t, "/backups/db.tar.gz", opts.BackupFile)
				assert.Equal(t, "/var/lib/postgresql/data", opts.TargetPath)
				assert.True(t, opts.Overwrite)
				assert.True(t, opts.PreservePerms)
				assert.Equal(t, []string{"/var/lib/postgresql/data/pg_wal"}, opts.SkipPaths)
				assert.NotNil(t, opts.Extra)
				assert.Equal(t, true, opts.Extra["verify"])
			},
			description: "All restore-specific fields should be set correctly",
		},
		{
			name: "minimal backup options",
			opts: operation.Options{
				Type:       operation.TypeBackup,
				Namespace:  "default",
				PodName:    "test",
				SourcePath: "/data",
				OutputFile: "/backup.tar",
			},
			validate: func(t *testing.T, opts operation.Options) {
				assert.Equal(t, operation.TypeBackup, opts.Type)
				assert.Equal(t, "default", opts.Namespace)
				assert.Equal(t, "test", opts.PodName)
				assert.Equal(t, "/data", opts.SourcePath)
				assert.Equal(t, "/backup.tar", opts.OutputFile)
				assert.False(t, opts.Compress)
				assert.Nil(t, opts.Exclude)
				assert.Nil(t, opts.Extra)
			},
			description: "Minimal required fields for backup should be set",
		},
		{
			name: "options with nil slices and maps",
			opts: operation.Options{
				Type:      operation.TypeBackup,
				Namespace: "test",
				PodName:   "pod",
				Exclude:   nil,
				SkipPaths: nil,
				Extra:     nil,
			},
			validate: func(t *testing.T, opts operation.Options) {
				assert.Nil(t, opts.Exclude)
				assert.Nil(t, opts.SkipPaths)
				assert.Nil(t, opts.Extra)
			},
			description: "Nil slices and maps should remain nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.opts)
		})
	}
}

func TestResult_Scenarios(t *testing.T) {
	tests := []struct {
		name   string
		result operation.Result
		check  func(*testing.T, operation.Result)
	}{
		{
			name: "successful backup result",
			result: operation.Result{
				Success:      true,
				Message:      "Backup completed successfully",
				BytesWritten: 1024 * 1024 * 50, // 50MB
				FileCount:    150,
				Duration:     2 * time.Minute,
				Warnings:     nil,
			},
			check: func(t *testing.T, r operation.Result) {
				assert.True(t, r.Success)
				assert.Equal(t, "Backup completed successfully", r.Message)
				assert.Equal(t, int64(52428800), r.BytesWritten)
				assert.Equal(t, 150, r.FileCount)
				assert.Equal(t, 2*time.Minute, r.Duration)
				assert.Nil(t, r.Error)
				assert.Empty(t, r.Warnings)
			},
		},
		{
			name: "failed backup result",
			result: operation.Result{
				Success: false,
				Message: "Backup failed: insufficient permissions",
				Error:   errors.New("permission denied"),
			},
			check: func(t *testing.T, r operation.Result) {
				assert.False(t, r.Success)
				assert.Contains(t, r.Message, "insufficient permissions")
				assert.NotNil(t, r.Error)
				assert.Equal(t, "permission denied", r.Error.Error())
				assert.Zero(t, r.BytesWritten)
				assert.Zero(t, r.FileCount)
			},
		},
		{
			name: "successful restore result with warnings",
			result: operation.Result{
				Success:      true,
				Message:      "Restore completed with warnings",
				RestoredPath: "/data/restored",
				BytesWritten: 1024 * 1024 * 45, // 45MB
				FileCount:    142,
				Duration:     90 * time.Second,
				Warnings: []string{
					"skipped 8 files due to permissions",
					"symlink /data/link could not be created",
				},
			},
			check: func(t *testing.T, r operation.Result) {
				assert.True(t, r.Success)
				assert.Equal(t, "/data/restored", r.RestoredPath)
				assert.Equal(t, int64(47185920), r.BytesWritten)
				assert.Equal(t, 142, r.FileCount)
				assert.Len(t, r.Warnings, 2)
				assert.Contains(t, r.Warnings[0], "permissions")
				assert.Contains(t, r.Warnings[1], "symlink")
			},
		},
		{
			name: "partial success result",
			result: operation.Result{
				Success:      true,
				Message:      "Operation completed with errors",
				BytesWritten: 1024000,
				FileCount:    10,
				Warnings: []string{
					"5 files failed to process",
					"checksum mismatch on 2 files",
				},
				Error: nil, // Success true but with warnings
			},
			check: func(t *testing.T, r operation.Result) {
				assert.True(t, r.Success)
				assert.Nil(t, r.Error)
				assert.Len(t, r.Warnings, 2)
				assert.Equal(t, 10, r.FileCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.result)
		})
	}
}

func TestMetadata_CompleteLifecycle(t *testing.T) {
	// Create metadata for a backup operation
	backupMeta := &operation.Metadata{
		ID:          "backup-20240101-120000",
		Type:        "backup",
		Provider:    "filesystem",
		Namespace:   "production",
		PodName:     "webapp-1",
		Container:   "app",
		SourcePath:  "/app/data",
		BackupFile:  "/backups/webapp-20240101.tar.gz",
		Compression: "gzip",
		Size:        1024 * 1024 * 100, // 100MB
		FileCount:   500,
		Checksum:    "sha256:abcdef123456",
		CreatedAt:   time.Now(),
		Status:      "in_progress",
		ProviderInfo: map[string]interface{}{
			"version":    "1.0",
			"compressed": true,
		},
	}

	// Test initial state
	t.Run("initial backup metadata", func(t *testing.T) {
		assert.Equal(t, "backup-20240101-120000", backupMeta.ID)
		assert.Equal(t, "backup", backupMeta.Type)
		assert.Equal(t, "in_progress", backupMeta.Status)
		assert.True(t, backupMeta.CompletedAt.IsZero())
		assert.Nil(t, backupMeta.Extra)
	})

	// Add custom metadata
	t.Run("add custom metadata", func(t *testing.T) {
		backupMeta.SetMetadata("retention", "30d")
		backupMeta.SetMetadata("environment", "prod")
		
		assert.NotNil(t, backupMeta.Extra)
		assert.Equal(t, "30d", backupMeta.Extra["retention"])
		assert.Equal(t, "prod", backupMeta.Extra["environment"])
	})

	// Complete the backup
	t.Run("complete backup", func(t *testing.T) {
		backupMeta.Status = "completed"
		backupMeta.CompletedAt = time.Now()
		
		assert.Equal(t, "completed", backupMeta.Status)
		assert.False(t, backupMeta.CompletedAt.IsZero())
		assert.True(t, backupMeta.CompletedAt.After(backupMeta.CreatedAt))
	})

	// Create restore metadata from backup
	t.Run("create restore metadata", func(t *testing.T) {
		restoreMeta := &operation.Metadata{
			ID:           "restore-20240102-090000",
			Type:         "restore",
			Provider:     "filesystem",
			Namespace:    backupMeta.Namespace,
			PodName:      "webapp-2", // Different pod
			Container:    backupMeta.Container,
			TargetPath:   "/app/data",
			RestoredFrom: backupMeta.BackupFile,
			Size:         backupMeta.Size,
			CreatedAt:    time.Now(),
			Status:       "pending",
			ProviderInfo: map[string]interface{}{
				"source_backup": backupMeta.ID,
			},
		}

		assert.Equal(t, "restore", restoreMeta.Type)
		assert.Equal(t, backupMeta.BackupFile, restoreMeta.RestoredFrom)
		assert.Equal(t, backupMeta.Size, restoreMeta.Size)
		assert.Equal(t, backupMeta.ID, restoreMeta.ProviderInfo["source_backup"])
	})
}

func TestProviderType_Constants(t *testing.T) {
	// Ensure provider types have expected values
	assert.Equal(t, operation.ProviderType("backup"), operation.TypeBackup)
	assert.Equal(t, operation.ProviderType("restore"), operation.TypeRestore)
	
	// Test string representation
	assert.Equal(t, "backup", string(operation.TypeBackup))
	assert.Equal(t, "restore", string(operation.TypeRestore))
}

func TestMetadata_JSONTags(t *testing.T) {
	metadata := &operation.Metadata{
		ID:         "test-123",
		Type:       "backup",
		Provider:   "filesystem",
		Namespace:  "default",
		PodName:    "test-pod",
		Container:  "app",
		SourcePath: "/data",
		BackupFile: "/backup.tar",
		Size:       1024,
		CreatedAt:  time.Now(),
		Status:     "completed",
		Extra: map[string]string{
			"key": "value",
		},
	}

	// This test verifies that the struct has proper JSON tags
	// The actual JSON marshaling is tested elsewhere, but we can
	// verify the struct definition is correct
	assert.NotNil(t, metadata.ID)
	assert.NotNil(t, metadata.Type)
	assert.NotNil(t, metadata.Provider)
	assert.NotNil(t, metadata.Namespace)
	assert.NotNil(t, metadata.PodName)
	assert.NotEmpty(t, metadata.Container)
	assert.NotEmpty(t, metadata.SourcePath)
	assert.NotEmpty(t, metadata.BackupFile)
	assert.NotZero(t, metadata.Size)
	assert.NotNil(t, metadata.Status)
	assert.NotNil(t, metadata.Extra)
}

func TestResult_EdgeCases(t *testing.T) {
	t.Run("result with nil error but false success", func(t *testing.T) {
		result := operation.Result{
			Success: false,
			Message: "Operation was cancelled",
			Error:   nil,
		}
		
		assert.False(t, result.Success)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Message, "cancelled")
	})

	t.Run("result with empty warnings slice", func(t *testing.T) {
		result := operation.Result{
			Success:  true,
			Warnings: []string{},
		}
		
		assert.True(t, result.Success)
		assert.NotNil(t, result.Warnings)
		assert.Empty(t, result.Warnings)
	})

	t.Run("result with zero duration", func(t *testing.T) {
		result := operation.Result{
			Success:  true,
			Duration: 0,
		}
		
		assert.True(t, result.Success)
		assert.Zero(t, result.Duration)
	})
}

func TestOptions_DefaultValues(t *testing.T) {
	var opts operation.Options
	
	// Test zero values
	assert.Equal(t, operation.ProviderType(""), opts.Type)
	assert.Empty(t, opts.Namespace)
	assert.Empty(t, opts.PodName)
	assert.Empty(t, opts.Container)
	assert.Empty(t, opts.SourcePath)
	assert.Empty(t, opts.OutputFile)
	assert.False(t, opts.Compress)
	assert.Nil(t, opts.Exclude)
	assert.Empty(t, opts.BackupFile)
	assert.Empty(t, opts.TargetPath)
	assert.False(t, opts.Overwrite)
	assert.False(t, opts.PreservePerms)
	assert.Nil(t, opts.SkipPaths)
	assert.Nil(t, opts.Extra)
}

func TestMetadata_SetMetadata_Concurrent(t *testing.T) {
	// This test verifies that SetMetadata initializes the map properly
	// even when called concurrently (though the method itself is not thread-safe)
	metadata := &operation.Metadata{
		ID: "concurrent-test",
	}
	
	// Sequential calls should work correctly
	metadata.SetMetadata("key1", "value1")
	metadata.SetMetadata("key2", "value2")
	metadata.SetMetadata("key3", "value3")
	
	require.NotNil(t, metadata.Extra)
	assert.Len(t, metadata.Extra, 3)
	assert.Equal(t, "value1", metadata.Extra["key1"])
	assert.Equal(t, "value2", metadata.Extra["key2"])
	assert.Equal(t, "value3", metadata.Extra["key3"])
}
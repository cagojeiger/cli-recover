package tui

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)

func TestCommandBuilder_BasicCommand(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*CommandBuilder)
		expected []string
	}{
		{
			name: "basic backup command",
			setup: func(cb *CommandBuilder) {
				cb.SetAction("backup")
				cb.SetPod("my-pod")
				cb.SetPath("/var/log")
			},
			expected: []string{"backup", "my-pod", "/var/log"},
		},
		{
			name: "restore command",
			setup: func(cb *CommandBuilder) {
				cb.SetAction("restore")
				cb.SetPod("my-pod")
				cb.SetPath("/var/log")
			},
			expected: []string{"restore", "my-pod", "/var/log"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCommandBuilder()
			tt.setup(cb)
			assert.Equal(t, tt.expected, cb.Build())
		})
	}
}

func TestCommandBuilder_WithNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		expected  []string
	}{
		{
			name:      "default namespace omitted",
			namespace: "default",
			expected:  []string{"backup", "my-pod", "/var/log"},
		},
		{
			name:      "custom namespace included",
			namespace: "production",
			expected:  []string{"backup", "my-pod", "/var/log", "--namespace", "production"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCommandBuilder()
			cb.SetAction("backup")
			cb.SetPod("my-pod")
			cb.SetPath("/var/log")
			cb.SetNamespace(tt.namespace)
			
			assert.Equal(t, tt.expected, cb.Build())
		})
	}
}

func TestCommandBuilder_WithOptions(t *testing.T) {
	cb := NewCommandBuilder()
	cb.SetAction("backup")
	cb.SetPod("my-pod")
	cb.SetPath("/var/log")
	
	options := kubernetes.BackupOptions{
		CompressionType: "xz",
		ExcludePatterns: []string{"*.log", "tmp/*"},
		ExcludeVCS:      true,
		Verbose:         true,
		ShowTotals:      false,
		PreservePerms:   true,
		Container:       "app",
		OutputFile:      "backup.tar.xz",
	}
	
	cb.SetOptions(options)
	
	result := cb.Build()
	
	// Verify required flags are present
	assert.Contains(t, result, "--compression")
	assert.Contains(t, result, "xz")
	assert.Contains(t, result, "--exclude")
	assert.Contains(t, result, "*.log")
	assert.Contains(t, result, "--exclude")
	assert.Contains(t, result, "tmp/*")
	assert.Contains(t, result, "--exclude-vcs")
	assert.Contains(t, result, "--verbose")
	assert.Contains(t, result, "--preserve-perms")
	assert.Contains(t, result, "--container")
	assert.Contains(t, result, "app")
	assert.Contains(t, result, "--output")
	assert.Contains(t, result, "backup.tar.xz")
	
	// Verify defaults are not included
	assert.NotContains(t, result, "--totals")
}

func TestCommandBuilder_Preview(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*CommandBuilder)
		expected string
	}{
		{
			name: "basic preview",
			setup: func(cb *CommandBuilder) {
				cb.SetAction("backup")
				cb.SetPod("my-pod")
				cb.SetPath("/var/log")
			},
			expected: "cli-recover backup my-pod /var/log",
		},
		{
			name: "preview with namespace",
			setup: func(cb *CommandBuilder) {
				cb.SetAction("backup")
				cb.SetPod("my-pod")
				cb.SetPath("/var/log")
				cb.SetNamespace("prod")
			},
			expected: "cli-recover backup my-pod /var/log --namespace prod",
		},
		{
			name: "preview with multiple options",
			setup: func(cb *CommandBuilder) {
				cb.SetAction("backup")
				cb.SetPod("my-pod")
				cb.SetPath("/var/log")
				cb.SetNamespace("prod")
				cb.SetOptions(kubernetes.BackupOptions{
					CompressionType: "xz",
					Verbose:         true,
				})
			},
			expected: "cli-recover backup my-pod /var/log --namespace prod --compression xz --verbose",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCommandBuilder()
			tt.setup(cb)
			assert.Equal(t, tt.expected, cb.Preview())
		})
	}
}

func TestCommandBuilder_Reset(t *testing.T) {
	cb := NewCommandBuilder()
	cb.SetAction("backup")
	cb.SetPod("my-pod")
	cb.SetPath("/var/log")
	cb.SetNamespace("prod")
	
	// Verify it has content
	assert.NotEmpty(t, cb.Build())
	
	// Reset
	cb.Reset()
	
	// Verify it's empty
	assert.Empty(t, cb.Build())
	assert.Equal(t, "cli-recover", cb.Preview())
}

func TestCommandBuilder_StepByStepBuild(t *testing.T) {
	cb := NewCommandBuilder()
	
	// Step 1: Action only
	cb.SetAction("backup")
	assert.Equal(t, "cli-recover backup", cb.Preview())
	
	// Step 2: Add pod
	cb.SetPod("my-pod")
	assert.Equal(t, "cli-recover backup my-pod", cb.Preview())
	
	// Step 3: Add path
	cb.SetPath("/var/log")
	assert.Equal(t, "cli-recover backup my-pod /var/log", cb.Preview())
	
	// Step 4: Add namespace
	cb.SetNamespace("prod")
	assert.Equal(t, "cli-recover backup my-pod /var/log --namespace prod", cb.Preview())
	
	// Step 5: Add options
	cb.SetOptions(kubernetes.BackupOptions{
		CompressionType: "xz",
		ExcludePatterns: []string{"*.tmp"},
	})
	preview := cb.Preview()
	assert.Contains(t, preview, "--compression xz")
	assert.Contains(t, preview, "--exclude *.tmp")
}

func TestCommandBuilder_WithBackupType(t *testing.T) {
	tests := []struct {
		name         string
		backupType   string
		expectedArgs []string
	}{
		{
			name:       "filesystem backup type",
			backupType: "filesystem",
			expectedArgs: []string{"backup", "filesystem", "test-pod", "/data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCommandBuilder()
			cb.SetAction("backup")
			cb.SetPod("test-pod")
			cb.SetBackupType(tt.backupType)
			
			// Set appropriate path based on backup type
			switch tt.backupType {
			case "filesystem":
				cb.SetPath("/data")
			case "minio":
				cb.SetPath("bucket/path")
			case "mongodb":
				cb.SetPath("db.collection")
			}
			
			args := cb.Build()
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestCommandBuilder_OptionsToFlags(t *testing.T) {
	tests := []struct {
		name     string
		options  kubernetes.BackupOptions
		expected []string
	}{
		{
			name:     "default options produce no flags",
			options:  kubernetes.BackupOptions{CompressionType: "gzip"},
			expected: []string{},
		},
		{
			name: "all options set",
			options: kubernetes.BackupOptions{
				CompressionType: "xz",
				ExcludePatterns: []string{"*.log", "tmp/*"},
				ExcludeVCS:      true,
				Verbose:         true,
				ShowTotals:      true,
				PreservePerms:   true,
				Container:       "sidecar",
				OutputFile:      "backup.tar",
				DryRun:          true,
			},
			expected: []string{
				"--compression", "xz",
				"--exclude", "*.log",
				"--exclude", "tmp/*",
				"--exclude-vcs",
				"--verbose",
				"--totals",
				"--preserve-perms",
				"--container", "sidecar",
				"--output", "backup.tar",
				"--dry-run",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCommandBuilder()
			cb.SetAction("backup")
			cb.SetPod("my-pod")
			cb.SetPath("/var/log")
			cb.SetOptions(tt.options)
			
			result := cb.Build()
			// Remove the base command parts
			flags := result[3:] // Skip "backup", "my-pod", "/var/log"
			
			if len(tt.expected) == 0 {
				assert.Empty(t, flags)
			} else {
				for i := 0; i < len(tt.expected); i++ {
					assert.Contains(t, flags, tt.expected[i])
				}
			}
		})
	}
}
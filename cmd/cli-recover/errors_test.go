package main

import (
	"errors"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
	"github.com/stretchr/testify/assert"
)

func TestCLIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		cliError *CLIError
		expected string
	}{
		{
			name: "error only",
			cliError: &CLIError{
				Message: "Something went wrong",
			},
			expected: "Something went wrong",
		},
		{
			name: "error with reason",
			cliError: &CLIError{
				Message: "File not found",
				Reason:  "Path does not exist",
			},
			expected: "File not found: Path does not exist",
		},
		{
			name: "error with cause",
			cliError: &CLIError{
				Message: "Operation failed",
				Cause:   errors.New("underlying error"),
			},
			expected: "Operation failed: underlying error",
		},
		{
			name: "full error",
			cliError: &CLIError{
				Message: "Backup failed",
				Reason:  "Disk full",
				Cause:   errors.New("write error"),
			},
			expected: "Backup failed: Disk full: write error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cliError.Error())
		})
	}
}

func TestCLIError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	cliErr := &CLIError{
		Message: "Wrapped error",
		Cause:   cause,
	}

	assert.Equal(t, cause, cliErr.Unwrap())
}

func TestConvertBackupError(t *testing.T) {
	tests := []struct {
		name        string
		backupError *backup.BackupError
		expectError string
		expectFix   string
	}{
		{
			name: "not found",
			backupError: &backup.BackupError{
				Code:    backup.ErrCodeNotFound,
				Message: "pod 'test' not found",
			},
			expectError: "Backup source not found",
			expectFix:   "Verify the pod name and path exist",
		},
		{
			name: "invalid input",
			backupError: &backup.BackupError{
				Code:    backup.ErrCodeInvalidInput,
				Message: "invalid compression type",
			},
			expectError: "Invalid backup parameters",
			expectFix:   "Check the command syntax and parameters",
		},
		{
			name: "timeout",
			backupError: &backup.BackupError{
				Code:    backup.ErrCodeTimeout,
				Message: "operation timed out",
			},
			expectError: "Backup operation timed out",
			expectFix:   "Try backing up smaller directories or check pod connectivity",
		},
		{
			name: "unauthorized",
			backupError: &backup.BackupError{
				Code:    backup.ErrCodeUnauthorized,
				Message: "access denied",
			},
			expectError: "Unauthorized to perform backup",
			expectFix:   "Check your Kubernetes RBAC permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliErr := convertBackupError(tt.backupError)
			assert.Equal(t, tt.expectError, cliErr.Message)
			assert.Equal(t, tt.expectFix, cliErr.Fix)
			assert.Equal(t, tt.backupError, cliErr.Cause)
		})
	}
}

func TestConvertRestoreError(t *testing.T) {
	tests := []struct {
		name         string
		restoreError *restore.RestoreError
		expectError  string
		expectFix    string
	}{
		{
			name: "not found",
			restoreError: &restore.RestoreError{
				Code:    "NOT_FOUND",
				Message: "backup file not found",
			},
			expectError: "Restore target not found",
			expectFix:   "Verify the backup file exists and pod is running",
		},
		{
			name: "invalid input",
			restoreError: &restore.RestoreError{
				Code:    "INVALID_INPUT",
				Message: "invalid target path",
			},
			expectError: "Invalid restore parameters",
			expectFix:   "Check the command syntax and parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliErr := convertRestoreError(tt.restoreError)
			assert.Equal(t, tt.expectError, cliErr.Message)
			assert.Equal(t, tt.expectFix, cliErr.Fix)
			assert.Equal(t, tt.restoreError, cliErr.Cause)
		})
	}
}

func TestConvertToCLIError_CommonErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectError string
		expectFix   string
	}{
		{
			name:        "file not found",
			err:         errors.New("open test.txt: no such file or directory"),
			expectError: "File or directory not found",
			expectFix:   "Check the file path and try again",
		},
		{
			name:        "permission denied",
			err:         errors.New("mkdir /root/test: permission denied"),
			expectError: "Permission denied",
			expectFix:   "Check file permissions or run with appropriate privileges",
		},
		{
			name:        "pod not found",
			err:         errors.New("pods \"test-pod\" not found"),
			expectError: "Pod not found",
			expectFix:   "Use 'kubectl get pods -n <namespace>' to list available pods",
		},
		{
			name:        "disk full",
			err:         errors.New("write test.tar: no space left on device"),
			expectError: "No space left on device",
			expectFix:   "Free up disk space and try again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cliErr := convertToCLIError(tt.err)
			assert.NotNil(t, cliErr)
			assert.Equal(t, tt.expectError, cliErr.Message)
			assert.Equal(t, tt.expectFix, cliErr.Fix)
			assert.Equal(t, tt.err, cliErr.Cause)
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewFileNotFoundError", func(t *testing.T) {
		err := NewFileNotFoundError("/tmp/test.txt")
		assert.Contains(t, err.Message, "/tmp/test.txt")
		assert.Contains(t, err.Fix, "Check the file path")
	})

	t.Run("NewPodNotFoundError", func(t *testing.T) {
		err := NewPodNotFoundError("test-pod", "default")
		assert.Contains(t, err.Message, "test-pod")
		assert.Contains(t, err.Message, "default")
		assert.Contains(t, err.Fix, "kubectl get pods -n default")
	})

	t.Run("NewInvalidFlagError", func(t *testing.T) {
		err := NewInvalidFlagError("--compression", "invalid", "gzip, none, or zstd")
		assert.Contains(t, err.Message, "--compression")
		assert.Contains(t, err.Message, "invalid")
		assert.Contains(t, err.Reason, "gzip, none, or zstd")
		assert.Contains(t, err.Fix, "--help")
	})
}

// Test that PrintError doesn't panic with various error types
func TestPrintError_NoPanic(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "nil error",
			err:  nil,
		},
		{
			name: "simple error",
			err:  errors.New("simple error"),
		},
		{
			name: "CLI error",
			err: &CLIError{
				Message: "CLI error",
				Fix:     "Do something",
			},
		},
		{
			name: "backup error",
			err: &backup.BackupError{
				Code:    backup.ErrCodeNotFound,
				Message: "not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			if tt.err != nil {
				assert.NotPanics(t, func() {
					PrintError(tt.err)
				})
			}
		})
	}
}
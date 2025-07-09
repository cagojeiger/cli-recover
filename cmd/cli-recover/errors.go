package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// CLIError represents a user-friendly error with structured information
type CLIError struct {
	// What went wrong
	Message string

	// Why it happened
	Reason string

	// How to fix it
	Fix string

	// Additional information (optional)
	Info string

	// Underlying error (if any)
	Cause error
}

// Error implements the error interface
func (e *CLIError) Error() string {
	var parts []string

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Reason != "" {
		parts = append(parts, e.Reason)
	}

	if e.Cause != nil {
		parts = append(parts, e.Cause.Error())
	}

	return strings.Join(parts, ": ")
}

// Unwrap returns the underlying error
func (e *CLIError) Unwrap() error {
	return e.Cause
}

// PrintError prints a formatted error message to stderr
func PrintError(err error) {
	// Try to convert to CLIError
	if cliErr, ok := err.(*CLIError); ok {
		printCLIError(cliErr)
		return
	}

	// Try to convert domain errors to CLI errors
	if cliErr := convertToCLIError(err); cliErr != nil {
		printCLIError(cliErr)
		return
	}

	// Fallback to simple error printing
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// printCLIError prints a structured CLI error
func printCLIError(e *CLIError) {
	// Main error message
	fmt.Fprintf(os.Stderr, "❌ Error: %s\n", e.Message)

	// Reason (if provided)
	if e.Reason != "" {
		fmt.Fprintf(os.Stderr, "   Reason: %s\n", e.Reason)
	}

	// Fix suggestion (if provided)
	if e.Fix != "" {
		fmt.Fprintf(os.Stderr, "   Fix: %s\n", e.Fix)
	}

	// Additional info (if provided)
	if e.Info != "" {
		fmt.Fprintf(os.Stderr, "   Info: %s\n", e.Info)
	}
}

// convertToCLIError converts domain errors to user-friendly CLI errors
func convertToCLIError(err error) *CLIError {
	switch e := err.(type) {
	case *backup.BackupError:
		return convertBackupError(e)
	case *restore.RestoreError:
		return convertRestoreError(e)
	default:
		// Check for common error patterns
		errStr := err.Error()

		// File not found
		if strings.Contains(errStr, "no such file or directory") {
			return &CLIError{
				Message: "File or directory not found",
				Reason:  "The specified path does not exist",
				Fix:     "Check the file path and try again",
				Cause:   err,
			}
		}

		// Permission denied
		if strings.Contains(errStr, "permission denied") {
			return &CLIError{
				Message: "Permission denied",
				Reason:  "You don't have permission to access this resource",
				Fix:     "Check file permissions or run with appropriate privileges",
				Cause:   err,
			}
		}

		// Kubernetes pod not found
		if strings.Contains(errStr, "pods") && strings.Contains(errStr, "not found") {
			return &CLIError{
				Message: "Pod not found",
				Reason:  "The specified pod does not exist in the namespace",
				Fix:     "Use 'kubectl get pods -n <namespace>' to list available pods",
				Cause:   err,
			}
		}

		// Disk space
		if strings.Contains(errStr, "no space left on device") {
			return &CLIError{
				Message: "No space left on device",
				Reason:  "The disk is full",
				Fix:     "Free up disk space and try again",
				Cause:   err,
			}
		}

		return nil
	}
}

// convertBackupError converts BackupError to CLIError
func convertBackupError(e *backup.BackupError) *CLIError {
	cliErr := &CLIError{
		Cause: e,
	}

	switch e.Code {
	case backup.ErrCodeNotFound:
		cliErr.Message = "Backup source not found"
		cliErr.Reason = e.Message
		cliErr.Fix = "Verify the pod name and path exist"

	case backup.ErrCodeInvalidInput:
		cliErr.Message = "Invalid backup parameters"
		cliErr.Reason = e.Message
		cliErr.Fix = "Check the command syntax and parameters"

	case backup.ErrCodeTimeout:
		cliErr.Message = "Backup operation timed out"
		cliErr.Reason = "The operation took too long to complete"
		cliErr.Fix = "Try backing up smaller directories or check pod connectivity"

	case backup.ErrCodeUnauthorized:
		cliErr.Message = "Unauthorized to perform backup"
		cliErr.Reason = "Insufficient permissions for the requested operation"
		cliErr.Fix = "Check your Kubernetes RBAC permissions"

	case backup.ErrCodeInternal:
		cliErr.Message = "Internal backup error"
		cliErr.Reason = e.Message
		cliErr.Fix = "Check the logs for more details"

	default:
		cliErr.Message = "Backup failed"
		cliErr.Reason = e.Message
	}

	return cliErr
}

// convertRestoreError converts RestoreError to CLIError
func convertRestoreError(e *restore.RestoreError) *CLIError {
	cliErr := &CLIError{
		Cause: e,
	}

	// RestoreError uses string codes, check common patterns
	switch e.Code {
	case "NOT_FOUND":
		cliErr.Message = "Restore target not found"
		cliErr.Reason = e.Message
		cliErr.Fix = "Verify the backup file exists and pod is running"

	case "INVALID_INPUT":
		cliErr.Message = "Invalid restore parameters"
		cliErr.Reason = e.Message
		cliErr.Fix = "Check the command syntax and parameters"

	case "TIMEOUT":
		cliErr.Message = "Restore operation timed out"
		cliErr.Reason = "The operation took too long to complete"
		cliErr.Fix = "Try restoring to a different pod or check connectivity"

	case "UNAUTHORIZED":
		cliErr.Message = "Unauthorized to perform restore"
		cliErr.Reason = "Insufficient permissions for the requested operation"
		cliErr.Fix = "Check your Kubernetes RBAC permissions"

	case "INTERNAL":
		cliErr.Message = "Internal restore error"
		cliErr.Reason = e.Message
		cliErr.Fix = "Check the logs for more details"

	default:
		cliErr.Message = "Restore failed"
		cliErr.Reason = e.Message
	}

	return cliErr
}

// Common error constructors

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(path string) *CLIError {
	return &CLIError{
		Message: fmt.Sprintf("File not found: %s", path),
		Reason:  "The specified file does not exist",
		Fix:     "Check the file path and try again",
	}
}

// NewPodNotFoundError creates a pod not found error
func NewPodNotFoundError(pod, namespace string) *CLIError {
	return &CLIError{
		Message: fmt.Sprintf("Pod '%s' not found in namespace '%s'", pod, namespace),
		Reason:  "The specified pod does not exist",
		Fix:     fmt.Sprintf("Use 'kubectl get pods -n %s' to list available pods", namespace),
	}
}

// NewInvalidFlagError creates an invalid flag error
func NewInvalidFlagError(flag, value, expected string) *CLIError {
	return &CLIError{
		Message: fmt.Sprintf("Invalid value for flag '%s': %s", flag, value),
		Reason:  fmt.Sprintf("Expected %s", expected),
		Fix:     "Use --help to see valid options",
	}
}

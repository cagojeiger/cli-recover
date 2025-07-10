package kubernetes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// RestoreExecutor handles binary-safe restore operations with kubectl
type RestoreExecutor struct {
	progressCh chan<- RestoreProgress
	mu         sync.Mutex
	fileCount  int
}

// RestoreProgress represents restore operation progress
type RestoreProgress struct {
	FileCount int
	FileName  string
	Error     error
}

// NewRestoreExecutor creates a new restore executor
func NewRestoreExecutor(progressCh chan<- RestoreProgress) *RestoreExecutor {
	return &RestoreExecutor{
		progressCh: progressCh,
	}
}

// ExecuteRestore performs a binary-safe restore operation
func (r *RestoreExecutor) ExecuteRestore(ctx context.Context, backupFile string, kubectlArgs []string) error {
	// Open backup file
	file, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	// Get file info for validation
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat backup file: %w", err)
	}

	// Validate it's a regular file
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("backup file is not a regular file")
	}

	// Build kubectl command
	cmd := exec.CommandContext(ctx, "kubectl", kubectlArgs...)

	// Connect backup file directly to stdin (binary-safe)
	cmd.Stdin = file

	// Capture stderr for progress monitoring
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Capture stdout for any output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start kubectl command: %w", err)
	}

	// Monitor progress from stderr in separate goroutine
	var wg sync.WaitGroup
	var progressErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		progressErr = r.monitorProgress(stderr)
	}()

	// Discard stdout (tar usually doesn't output to stdout)
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(io.Discard, stdout)
	}()

	// Wait for command to complete
	cmdErr := cmd.Wait()

	// Wait for goroutines to finish
	wg.Wait()

	// Check for errors
	if cmdErr != nil {
		return fmt.Errorf("kubectl exec failed: %w", cmdErr)
	}

	if progressErr != nil {
		return fmt.Errorf("progress monitoring failed: %w", progressErr)
	}

	return nil
}

// monitorProgress monitors tar verbose output from stderr
func (r *RestoreExecutor) monitorProgress(stderr io.Reader) error {
	scanner := bufio.NewScanner(stderr)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse tar verbose output
		// Format: "x path/to/file" or "x path/to/dir/"
		if strings.HasPrefix(line, "x ") {
			fileName := strings.TrimPrefix(line, "x ")
			fileName = strings.TrimSpace(fileName)

			if fileName != "" {
				r.mu.Lock()
				r.fileCount++
				count := r.fileCount
				r.mu.Unlock()

				// Send progress update
				if r.progressCh != nil {
					select {
					case r.progressCh <- RestoreProgress{
						FileCount: count,
						FileName:  fileName,
					}:
					default:
						// Channel full, skip this update
					}
				}
			}
		} else if strings.Contains(line, "tar:") {
			// Handle tar errors/warnings
			if r.progressCh != nil {
				select {
				case r.progressCh <- RestoreProgress{
					Error: fmt.Errorf("tar: %s", line),
				}:
				default:
				}
			}
		}
		// Ignore other stderr output
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stderr: %w", err)
	}

	return nil
}

// StreamRestore executes restore with streaming support (alternative implementation)
func StreamRestore(ctx context.Context, backupFile string, namespace, pod, container, targetPath string,
	preservePerms bool, overwrite bool, skipPaths []string) (<-chan RestoreProgress, error) {

	progressCh := make(chan RestoreProgress, 100)

	go func() {
		defer close(progressCh)

		// Build kubectl args
		kubectlArgs := []string{"exec", "-i", "-n", namespace, pod}

		if container != "" {
			kubectlArgs = append(kubectlArgs, "-c", container)
		}

		kubectlArgs = append(kubectlArgs, "--", "tar")

		// Detect compression from file extension
		compression := detectCompression(backupFile)
		switch compression {
		case "gzip":
			kubectlArgs = append(kubectlArgs, "-xzvf")
		case "bzip2":
			kubectlArgs = append(kubectlArgs, "-xjvf")
		case "xz":
			kubectlArgs = append(kubectlArgs, "-xJvf")
		default:
			kubectlArgs = append(kubectlArgs, "-xvf")
		}

		// Add stdin flag
		kubectlArgs = append(kubectlArgs, "-")

		// Add target directory
		kubectlArgs = append(kubectlArgs, "-C", targetPath)

		// Add overwrite flag if needed
		if !overwrite {
			kubectlArgs = append(kubectlArgs, "--keep-old-files")
		}

		// Add preserve permissions flag
		if preservePerms {
			kubectlArgs = append(kubectlArgs, "-p")
		}

		// Add skip paths
		for _, skip := range skipPaths {
			kubectlArgs = append(kubectlArgs, "--exclude="+skip)
		}

		// Execute restore
		executor := NewRestoreExecutor(progressCh)
		if err := executor.ExecuteRestore(ctx, backupFile, kubectlArgs); err != nil {
			progressCh <- RestoreProgress{Error: err}
		}
	}()

	return progressCh, nil
}

// detectCompression detects compression type from file extension
func detectCompression(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return "gzip"
	case strings.HasSuffix(lower, ".tar.bz2"):
		return "bzip2"
	case strings.HasSuffix(lower, ".tar.xz"):
		return "xz"
	default:
		return "none"
	}
}

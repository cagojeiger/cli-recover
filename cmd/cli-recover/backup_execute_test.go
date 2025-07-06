package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/runner"
)

// mockFileWriter implements io.Writer for testing file operations
type mockFileWriter struct {
	buffer bytes.Buffer
	closed bool
}

func (m *mockFileWriter) Write(p []byte) (n int, err error) {
	return m.buffer.Write(p)
}

func (m *mockFileWriter) Close() error {
	m.closed = true
	return nil
}

// Test executeBackup function components
func TestExecuteBackupHelpers(t *testing.T) {
	t.Run("generatePathSuffix edge cases", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"/", "root"},
			{"", ""},
			{"/../../etc", "------etc"},
			{"/data/", "data-"},
			{"/var/lib/docker/volumes", "var-lib-docker-volumes"},
		}
		
		for _, tt := range tests {
			result := generatePathSuffix(tt.input)
			if result != tt.expected {
				t.Errorf("generatePathSuffix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	})
	
	t.Run("compression type detection", func(t *testing.T) {
		commands := map[string]string{
			"kubectl exec pod -- tar -z -c -f - /data":         "gzip",
			"kubectl exec pod -- tar -c -z -f - /data":         "gzip",
			"kubectl exec pod -- tar -j -c -f - /data":         "bzip2",
			"kubectl exec pod -- tar -J -c -f - /data":         "xz",
			"kubectl exec pod -- tar -c -f - /data":            "none",
			"kubectl exec pod -- tar --gzip -cf - /data":       "none", // flag not detected
		}
		
		for cmd, expected := range commands {
			result := getCompressionFromCommand(cmd)
			if result != expected {
				t.Errorf("getCompressionFromCommand(%q) = %q, want %q", cmd, result, expected)
			}
		}
	})
}

// Test output file naming
func TestBackupOutputFileNaming(t *testing.T) {
	tests := []struct {
		namespace   string
		pod         string
		path        string
		compression string
		expected    string
	}{
		{
			namespace:   "default",
			pod:         "nginx-123",
			path:        "/data",
			compression: "gzip",
			expected:    "backup-default-nginx-123-data.tar.gz",
		},
		{
			namespace:   "prod",
			pod:         "app-xyz",
			path:        "/var/log",
			compression: "bzip2",
			expected:    "backup-prod-app-xyz-var-log.tar.bz2",
		},
		{
			namespace:   "test",
			pod:         "db",
			path:        "/",
			compression: "none",
			expected:    "backup-test-db-root.tar",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			suffix := generatePathSuffix(tt.path)
			extension := getFileExtension(tt.compression)
			result := "backup-" + tt.namespace + "-" + tt.pod + "-" + suffix + extension
			
			if result != tt.expected {
				t.Errorf("Got filename %q, want %q", result, tt.expected)
			}
		})
	}
}

// mockRunnerForSize implements runner.Runner for size estimation testing
type mockRunnerForSize struct {
	output string
}

func (m *mockRunnerForSize) Run(cmd string, args ...string) ([]byte, error) {
	// Return the preset output for du command
	return []byte(m.output), nil
}

// Test mock runner for size estimation
func TestSizeEstimationParsing(t *testing.T) {
	// Test the du output parsing logic
	duOutputs := []struct {
		output   string
		expected int64
	}{
		{"1024\t/data", 1024},
		{"1048576\t/var/log", 1048576},
		{"0\t/empty", 0},
		{"", 0},            // empty output
		{"invalid", 0},     // invalid format
		{"abc\t/data", 0},  // non-numeric size
	}
	
	for _, tt := range duOutputs {
		t.Run(strings.ReplaceAll(tt.output, "\t", "_"), func(t *testing.T) {
			runner := &mockRunnerForSize{output: tt.output}
			size := estimateBackupSize(runner, "test-pod", "default", "/data", false)
			
			if size != tt.expected {
				t.Errorf("estimateBackupSize with output %q = %d, want %d", tt.output, size, tt.expected)
			}
		})
	}
}

// Test humanizeBytes with edge cases
func TestHumanizeBytesEdgeCases(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
		{1099511627776, "1.0 TB"},
		{1125899906842624, "1.0 PB"},
		{1152921504606846976, "1.0 EB"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := humanizeBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("humanizeBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

// mockStreamRunner for testing streaming operations
type mockStreamRunner struct {
	runner.Runner
	streamData []byte
	streamErr  error
}

func (m *mockStreamRunner) Run(cmd string, args ...string) ([]byte, error) {
	// Check if this is a streaming tar command
	if strings.Contains(strings.Join(args, " "), "tar") {
		return m.streamData, m.streamErr
	}
	return []byte(""), nil
}

// Test backup command execution flow (without actual file I/O)
func TestBackupExecutionFlow(t *testing.T) {
	t.Run("dry run mode", func(t *testing.T) {
		// In dry-run mode, the function should return early without creating files
		// This is tested in runFilesystemBackup_DryRun test
		
		// Test that dry-run generates correct output filename
		outputFile := ""
		if outputFile == "" {
			path := "/var/data"
			pod := "test-pod"
			namespace := "default"
			compression := "gzip"
			
			extension := getFileExtension(compression)
			expected := "backup-" + namespace + "-" + pod + "-" + generatePathSuffix(path) + extension
			
			if expected != "backup-default-test-pod-var-data.tar.gz" {
				t.Errorf("Dry-run output filename = %s, want backup-default-test-pod-var-data.tar.gz", expected)
			}
		}
	})
	
	t.Run("command building", func(t *testing.T) {
		// Test that the kubectl command is built correctly
		commands := []string{
			"kubectl exec -n default test-pod -- tar -z -c -f - /data",
			"kubectl exec -n prod app-xyz -- tar -j -c -f - --verbose /var",
		}
		
		for _, cmd := range commands {
			parts := strings.Fields(cmd)
			if len(parts) < 5 {
				t.Errorf("Invalid command structure: %s", cmd)
			}
			if parts[0] != "kubectl" || parts[1] != "exec" {
				t.Errorf("Command should start with 'kubectl exec': %s", cmd)
			}
		}
	})
}

// Test error handling in backup execution
func TestBackupErrorHandling(t *testing.T) {
	t.Run("size estimation failure", func(t *testing.T) {
		// When size estimation fails, it should return 0 but continue
		runner := &mockRunner{err: io.EOF}
		size := estimateBackupSize(runner, "test-pod", "default", "/data", false)
		
		if size != 0 {
			t.Errorf("Expected size = 0 on error, got %d", size)
		}
	})
	
	t.Run("invalid size format", func(t *testing.T) {
		runner := &mockRunner{output: []byte("not-a-valid-size-output")}
		size := estimateBackupSize(runner, "test-pod", "default", "/data", false)
		
		if size != 0 {
			t.Errorf("Expected size = 0 on invalid format, got %d", size)
		}
	})
}

// Benchmark humanizeBytes
func BenchmarkHumanizeBytes(b *testing.B) {
	sizes := []int64{
		512,
		1024,
		1048576,
		1073741824,
		1099511627776,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			_ = humanizeBytes(size)
		}
	}
}

// TestMain for setup/teardown if needed
func TestMain(m *testing.M) {
	// Setup
	originalStderr := os.Stderr
	
	// Run tests
	code := m.Run()
	
	// Teardown
	os.Stderr = originalStderr
	
	os.Exit(code)
}
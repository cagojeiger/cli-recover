package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)

// mockRunnerWithPods implements runner.Runner for testing with pod data
type mockRunnerWithPods struct {
	pods      []kubernetes.Pod
	shouldErr bool
	cmdLog    []string // log of commands run
}

func (m *mockRunnerWithPods) Run(cmd string, args ...string) ([]byte, error) {
	// Log the command
	m.cmdLog = append(m.cmdLog, fmt.Sprintf("%s %s", cmd, strings.Join(args, " ")))
	
	if m.shouldErr {
		return nil, errors.New("mock error")
	}

	// Mock kubectl get namespaces
	if len(args) >= 3 && args[0] == "get" && args[1] == "namespaces" {
		return []byte(`{"items":[{"metadata":{"name":"default"}},{"metadata":{"name":"kube-system"}}]}`), nil
	}

	// Mock kubectl get pods
	if len(args) >= 3 && args[0] == "get" && args[1] == "pods" {
		if len(m.pods) == 0 {
			return []byte(`{"items":[]}`), nil
		}
		
		// Build a simple JSON response with our mock pods
		items := []string{}
		for _, pod := range m.pods {
			items = append(items, fmt.Sprintf(`{
				"metadata":{"name":"%s","namespace":"%s"},
				"spec":{"containers":[{"name":"main"}]},
				"status":{"phase":"%s","containerStatuses":[{"ready":true,"name":"main"}]}
			}`, pod.Name, pod.Namespace, pod.Status))
		}
		return []byte(fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))), nil
	}

	// Mock kubectl exec for du command (size estimation)
	if len(args) >= 2 && args[0] == "exec" && strings.Contains(strings.Join(args, " "), "du -sb") {
		return []byte("1048576\t/data\n"), nil // 1MB
	}

	// Default response
	return []byte(""), nil
}

func TestNewFilesystemBackupCmd(t *testing.T) {
	cmd := newFilesystemBackupCmd()
	
	// Test basic command properties
	if cmd.Use != "filesystem [pod] [path]" {
		t.Errorf("Expected Use = 'filesystem [pod] [path]', got %s", cmd.Use)
	}
	
	if cmd.Short != "Backup pod filesystem" {
		t.Errorf("Expected Short = 'Backup pod filesystem', got %s", cmd.Short)
	}
	
	// Test that all expected flags are present
	expectedFlags := []string{
		"namespace", "compression", "exclude", "exclude-vcs",
		"verbose", "totals", "preserve-perms", "container",
		"output", "dry-run",
	}
	
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", flagName)
		}
	}
	
	// Test flag defaults
	ns, _ := cmd.Flags().GetString("namespace")
	if ns != "default" {
		t.Errorf("Expected namespace default = 'default', got %s", ns)
	}
	
	comp, _ := cmd.Flags().GetString("compression")
	if comp != "gzip" {
		t.Errorf("Expected compression default = 'gzip', got %s", comp)
	}
}

func TestGeneratePathSuffix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/", "root"},
		{"/data", "data"},
		{"/var/log", "var-log"},
		{"/var/lib/mysql", "var-lib-mysql"},
		{"/path with spaces", "path-with-spaces"},
		{"/path.with.dots", "path-with-dots"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := generatePathSuffix(tt.input)
			if result != tt.expected {
				t.Errorf("generatePathSuffix(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetCompressionFromCommand(t *testing.T) {
	tests := []struct {
		command  string
		expected string
	}{
		{"kubectl exec -- tar -z -c -f - /data", "gzip"},
		{"kubectl exec -- tar -j -c -f - /data", "bzip2"},
		{"kubectl exec -- tar -J -c -f - /data", "xz"},
		{"kubectl exec -- tar -c -f - /data", "none"},
		{"kubectl exec -- tar -zcf - /data", "gzip"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getCompressionFromCommand(tt.command)
			if result != tt.expected {
				t.Errorf("getCompressionFromCommand(%s) = %s, want %s", tt.command, result, tt.expected)
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		compression string
		expected    string
	}{
		{"gzip", ".tar.gz"},
		{"bzip2", ".tar.bz2"},
		{"xz", ".tar.xz"},
		{"none", ".tar"},
		{"unknown", ".tar.gz"},
	}
	
	for _, tt := range tests {
		t.Run(tt.compression, func(t *testing.T) {
			result := getFileExtension(tt.compression)
			if result != tt.expected {
				t.Errorf("getFileExtension(%s) = %s, want %s", tt.compression, result, tt.expected)
			}
		})
	}
}

func TestRunFilesystemBackup_PodNotFound(t *testing.T) {
	// Create command with mock runner that has no pods
	cmd := &cobra.Command{}
	
	// Add required flags
	cmd.Flags().String("namespace", "default", "")
	cmd.Flags().String("compression", "gzip", "")
	cmd.Flags().StringSlice("exclude", []string{}, "")
	cmd.Flags().Bool("exclude-vcs", false, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("totals", false, "")
	cmd.Flags().Bool("preserve-perms", false, "")
	cmd.Flags().String("container", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	
	// Override the runner creation in the function
	// Since we can't easily inject the runner, we'll test the error case
	// by providing a non-existent pod name
	err := runFilesystemBackup(cmd, []string{"nonexistent-pod", "/data"})
	
	if err == nil {
		t.Error("Expected error for non-existent pod, got nil")
	}
	
	if !strings.Contains(err.Error(), "failed to get pods") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error about pod not found, got: %v", err)
	}
}

func TestRunFilesystemBackup_DryRun(t *testing.T) {
	// This test would require more complex mocking or refactoring
	// to inject the runner. For now, we can test that the function
	// accepts the correct arguments
	cmd := &cobra.Command{}
	
	// Add required flags
	cmd.Flags().String("namespace", "default", "")
	cmd.Flags().String("compression", "gzip", "")
	cmd.Flags().StringSlice("exclude", []string{}, "")
	cmd.Flags().Bool("exclude-vcs", false, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("totals", false, "")
	cmd.Flags().Bool("preserve-perms", false, "")
	cmd.Flags().String("container", "", "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().Bool("debug", false, "")
	cmd.Flags().Bool("dry-run", true, "") // Enable dry-run
	
	// Test will fail due to pod not found, but that's expected
	// In a real scenario with proper dependency injection, we would
	// verify that no actual backup is performed in dry-run mode
	_ = runFilesystemBackup(cmd, []string{"test-pod", "/data"})
}

// Additional test for estimateBackupSize with valid output
func TestEstimateBackupSize_ValidOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int64
	}{
		{
			name:     "valid size output",
			output:   "1048576\t/data\n",
			expected: 1048576,
		},
		{
			name:     "size with spaces",
			output:   "2097152    /var/log\n",
			expected: 2097152,
		},
		{
			name:     "empty output",
			output:   "",
			expected: 0,
		},
		{
			name:     "invalid format",
			output:   "not-a-number\t/data\n",
			expected: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRunner{output: []byte(tt.output)}
			size := estimateBackupSize(runner, "test-pod", "default", "/data", false)
			
			if size != tt.expected {
				t.Errorf("estimateBackupSize() = %d, want %d", size, tt.expected)
			}
		})
	}
}
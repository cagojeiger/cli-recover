package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoldenRunner_Run(t *testing.T) {
	// Create a temporary directory for test golden files
	tempDir := t.TempDir()
	
	// Create test golden files
	testCases := []struct {
		name     string
		command  string
		args     []string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "successful command",
			command:  "kubectl",
			args:     []string{"get", "namespaces"},
			filename: "get-namespaces.golden",
			content:  "default\nkube-system\n",
			wantErr:  false,
		},
		{
			name:     "command with complex args",
			command:  "kubectl",
			args:     []string{"get", "pods", "-n", "default"},
			filename: "get-pods-n-default.golden",
			content:  "test-pod\n",
			wantErr:  false,
		},
		{
			name:     "missing golden file",
			command:  "missing",
			args:     []string{},
			filename: ".golden",
			content:  "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create command subdirectory
			cmdDir := filepath.Join(tempDir, tc.command)
			err := os.MkdirAll(cmdDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create command directory: %v", err)
			}

			// Create golden file if content is provided
			if tc.content != "" {
				err := os.WriteFile(filepath.Join(cmdDir, tc.filename), []byte(tc.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create golden file: %v", err)
				}
			}

			runner := &GoldenRunner{dir: tempDir}
			output, err := runner.Run(tc.command, tc.args...)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if string(output) != tc.content {
				t.Errorf("Expected %q, got %q", tc.content, string(output))
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		want    string
	}{
		{
			name:    "simple args",
			command: "kubectl",
			args:    []string{"get", "pods"},
			want:    "get-pods.golden",
		},
		{
			name:    "args with namespace",
			command: "kubectl",
			args:    []string{"get", "pods", "-n", "default"},
			want:    "get-pods-n-default.golden",
		},
		{
			name:    "args with special characters",
			command: "kubectl",
			args:    []string{"get", "pods", "--selector=app=test"},
			want:    "get-pods--selector=app=test.golden",
		},
		{
			name:    "empty args",
			command: "kubectl",
			args:    []string{},
			want:    ".golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.command, tt.args)
			if got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGoldenRunner(t *testing.T) {
	testDir := "/test/path"
	runner := NewGoldenRunner(testDir)
	
	if runner == nil {
		t.Fatal("NewGoldenRunner returned nil")
	}
	
	if runner.dir != testDir {
		t.Errorf("Expected dir to be %s, got %s", testDir, runner.dir)
	}
}
package kubernetes

import (
	"strings"
	"testing"
)

func TestGenerateBackupCommand(t *testing.T) {
	tests := []struct {
		name      string
		pod       string
		namespace string
		path      string
		options   BackupOptions
		want      string
	}{
		{
			name:      "basic backup command with gzip",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "gzip",
			},
			want: "kubectl exec -n default test-pod -- tar -z -c -f - /app",
		},
		{
			name:      "backup with bzip2 compression",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "bzip2",
			},
			want: "kubectl exec -n default test-pod -- tar -j -c -f - /app",
		},
		{
			name:      "backup with xz compression",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "xz",
			},
			want: "kubectl exec -n default test-pod -- tar -J -c -f - /app",
		},
		{
			name:      "backup with no compression",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "none",
			},
			want: "kubectl exec -n default test-pod -- tar -c -f - /app",
		},
		{
			name:      "backup with exclude patterns",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "gzip",
				ExcludePatterns: []string{"*.log", "*.tmp"},
			},
			want: "kubectl exec -n default test-pod -- tar -z -c -f - --exclude=*.log --exclude=*.tmp /app",
		},
		{
			name:      "backup with VCS exclusion",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "gzip",
				ExcludeVCS:      true,
			},
			want: "kubectl exec -n default test-pod -- tar -z -c -f - --exclude-vcs /app",
		},
		{
			name:      "backup with verbose and totals",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "gzip",
				Verbose:         true,
				ShowTotals:      true,
				PreservePerms:   true,
			},
			want: "kubectl exec -n default test-pod -- tar -z -c -f - --verbose --totals --preserve-permissions /app",
		},
		{
			name:      "backup with container name",
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			options: BackupOptions{
				CompressionType: "gzip",
				Container:       "web",
			},
			want: "kubectl exec -n default test-pod -c web -- tar -z -c -f - /app",
		},
		{
			name:      "backup with all options",
			pod:       "api-server",
			namespace: "prod",
			path:      "/data",
			options: BackupOptions{
				CompressionType: "xz",
				ExcludePatterns: []string{"*.log", "temp/*"},
				ExcludeVCS:      true,
				Verbose:         true,
				ShowTotals:      true,
				PreservePerms:   true,
				Container:       "api",
			},
			want: "kubectl exec -n prod api-server -c api -- tar -J -c -f - --verbose --totals --preserve-permissions --exclude=*.log --exclude=temp/* --exclude-vcs /data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateBackupCommand(tt.pod, tt.namespace, tt.path, tt.options)

			if got != tt.want {
				t.Errorf("GenerateBackupCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackupOptions_Validation(t *testing.T) {
	tests := []struct {
		name    string
		options BackupOptions
		valid   bool
	}{
		{
			name: "valid gzip compression",
			options: BackupOptions{
				CompressionType: "gzip",
			},
			valid: true,
		},
		{
			name: "valid bzip2 compression",
			options: BackupOptions{
				CompressionType: "bzip2",
			},
			valid: true,
		},
		{
			name: "valid xz compression",
			options: BackupOptions{
				CompressionType: "xz",
			},
			valid: true,
		},
		{
			name: "valid no compression",
			options: BackupOptions{
				CompressionType: "none",
			},
			valid: true,
		},
		{
			name: "empty compression type defaults to valid",
			options: BackupOptions{
				CompressionType: "",
			},
			valid: true,
		},
		{
			name: "valid exclude patterns",
			options: BackupOptions{
				CompressionType: "gzip",
				ExcludePatterns: []string{"*.log", "*.tmp", ".git/*"},
			},
			valid: true,
		},
		{
			name: "valid container name",
			options: BackupOptions{
				CompressionType: "gzip",
				Container:       "web-server",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates that the options structure is correct
			// and can be used to generate commands without errors
			cmd := GenerateBackupCommand("test-pod", "default", "/app", tt.options)
			
			if tt.valid {
				if cmd == "" {
					t.Error("Expected non-empty command for valid options")
				}
				if !strings.Contains(cmd, "kubectl exec") {
					t.Error("Expected command to contain 'kubectl exec'")
				}
			}
		})
	}
}
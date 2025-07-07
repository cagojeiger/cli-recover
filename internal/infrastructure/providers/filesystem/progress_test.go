package filesystem

import (
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/stretchr/testify/assert"
)

func TestMonitorProgress(t *testing.T) {
	provider := &Provider{
		progressCh: make(chan backup.Progress, 10),
	}
	
	// Create test output channel
	outputCh := make(chan string)
	opts := backup.Options{
		SourcePath: "/data",
	}
	
	// Start monitoring in goroutine
	go provider.monitorProgress(outputCh, opts)
	
	// Send test data
	outputCh <- "tar: data/file1.txt"
	outputCh <- "tar: data/file2.txt"
	outputCh <- "some other output"
	outputCh <- "tar: data/dir/file3.txt"
	close(outputCh)
	
	// Collect progress updates
	var updates []backup.Progress
	for i := 0; i < 3; i++ {
		update := <-provider.progressCh
		updates = append(updates, update)
	}
	
	// Verify progress updates
	assert.Len(t, updates, 3)
	assert.Equal(t, int64(1), updates[0].Current)
	assert.Equal(t, "Backing up: data/file1.txt", updates[0].Message)
	assert.Equal(t, int64(2), updates[1].Current)
	assert.Equal(t, "Backing up: data/file2.txt", updates[1].Message)
	assert.Equal(t, int64(3), updates[2].Current)
	assert.Equal(t, "Backing up: data/dir/file3.txt", updates[2].Message)
}

func TestBuildTarCommand(t *testing.T) {
	provider := &Provider{}
	
	tests := []struct {
		name     string
		opts     backup.Options
		expected []string
	}{
		{
			name: "basic tar without compression",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar",
				Compress:   false,
			},
			expected: []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
				"tar", "-cf", "-", "-C", "/", "data", ">", "backup.tar"},
		},
		{
			name: "tar with compression",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar.gz",
				Compress:   true,
			},
			expected: []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
				"tar", "-czf", "-", "-C", "/", "data", ">", "backup.tar.gz"},
		},
		{
			name: "tar with excludes",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar",
				Exclude:    []string{"*.log", "tmp/"},
			},
			expected: []string{"kubectl", "exec", "-n", "default", "test-pod", "--", 
				"tar", "-cf", "-", "--exclude=*.log", "--exclude=tmp/", "-C", "/", "data", ">", "backup.tar"},
		},
		{
			name: "tar with container specified",
			opts: backup.Options{
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar",
				Extra: map[string]interface{}{
					"container": "app",
				},
			},
			expected: []string{"kubectl", "exec", "-n", "default", "test-pod", "-c", "app", "--", 
				"tar", "-cf", "-", "-C", "/", "data", ">", "backup.tar"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.buildTarCommand(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}
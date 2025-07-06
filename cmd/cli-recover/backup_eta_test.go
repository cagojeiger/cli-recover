package main

import (
	"testing"
)

// mockRunner implements the runner.Runner interface for testing
type mockRunner struct {
	output []byte
	err    error
}

func (m *mockRunner) Run(cmd string, args ...string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

func TestEstimateBackupSize(t *testing.T) {
	tests := []struct {
		name      string
		pod       string
		namespace string
		path      string
		expectZero bool // expect 0 size (error case)
	}{
		{
			name:       "invalid pod should return zero",
			pod:        "nonexistent-pod",
			namespace:  "default",
			path:       "/tmp",
			expectZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock runner - this will fail for non-existent pods
			runner := &mockRunner{}
			
			size := estimateBackupSize(runner, tt.pod, tt.namespace, tt.path, true)
			
			if tt.expectZero && size != 0 {
				t.Errorf("Expected zero size for invalid pod, got %d", size)
			}
		})
	}
}

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
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

// Mock runner for testing - updated to match runner.Runner interface
func (m *mockRunner) Stream(cmd string, args ...string) error {
	return nil
}
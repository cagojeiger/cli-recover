package runner

import (
	"os"
	"testing"
)

func TestNewRunner(t *testing.T) {
	tests := []struct {
		name        string
		useGoldenEnv string
		wantType    string
	}{
		{
			name:        "creates shell runner when USE_GOLDEN is not set",
			useGoldenEnv: "",
			wantType:    "*runner.ShellRunner",
		},
		{
			name:        "creates golden runner when USE_GOLDEN is true",
			useGoldenEnv: "true",
			wantType:    "*runner.GoldenRunner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			oldEnv := os.Getenv("USE_GOLDEN")
			defer os.Setenv("USE_GOLDEN", oldEnv)
			
			os.Setenv("USE_GOLDEN", tt.useGoldenEnv)
			
			runner := NewRunner()
			if runner == nil {
				t.Fatal("NewRunner returned nil")
			}
			
			// Check the type by attempting to cast
			switch tt.wantType {
			case "*runner.ShellRunner":
				if _, ok := runner.(*ShellRunner); !ok {
					t.Errorf("Expected ShellRunner, got %T", runner)
				}
			case "*runner.GoldenRunner":
				if _, ok := runner.(*GoldenRunner); !ok {
					t.Errorf("Expected GoldenRunner, got %T", runner)
				}
			}
		})
	}
}

func TestShellRunner_Run(t *testing.T) {
	runner := &ShellRunner{}
	
	// Test a simple command that should exist on most systems
	output, err := runner.Run("echo", "test")
	if err != nil {
		t.Fatalf("ShellRunner.Run failed: %v", err)
	}
	
	if string(output) != "test\n" {
		t.Errorf("Expected 'test\\n', got %q", string(output))
	}
}

func TestShellRunner_Run_NonExistentCommand(t *testing.T) {
	runner := &ShellRunner{}
	
	_, err := runner.Run("non-existent-command-12345")
	if err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}
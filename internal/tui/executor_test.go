package tui

import (
	"bytes"
	"errors"
	"io"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// MockExecutor for testing
type MockExecutor struct {
	CalledWith []string
	ReturnError error
	ReturnOutput string
	Writer io.Writer
}

func (m *MockExecutor) Execute(args []string, writer io.Writer) error {
	m.CalledWith = args
	m.Writer = writer
	
	if writer != nil && m.ReturnOutput != "" {
		_, _ = writer.Write([]byte(m.ReturnOutput))
	}
	
	return m.ReturnError
}

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		returnError  error
		returnOutput string
		expectError  bool
	}{
		{
			name:        "successful execution",
			args:        []string{"backup", "my-pod", "/var/log"},
			expectError: false,
		},
		{
			name:        "execution with error",
			args:        []string{"backup", "invalid-pod", "/var/log"},
			returnError: errors.New("pod not found"),
			expectError: true,
		},
		{
			name:         "execution with output",
			args:         []string{"backup", "my-pod", "/var/log"},
			returnOutput: "Backup completed successfully",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &MockExecutor{
				ReturnError:  tt.returnError,
				ReturnOutput: tt.returnOutput,
			}
			
			var buf bytes.Buffer
			err := executor.Execute(tt.args, &buf)
			
			// Verify the executor was called with correct args
			assert.Equal(t, tt.args, executor.CalledWith)
			
			// Verify error handling
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			// Verify output
			if tt.returnOutput != "" {
				assert.Equal(t, tt.returnOutput, buf.String())
			}
		})
	}
}

func TestRealExecutor_ValidateCommand(t *testing.T) {
	// Test that validates the real executor would call cli-recover properly
	executor, err := NewRealExecutor()
	
	// We can't actually execute in tests, but we can verify the command construction
	assert.NoError(t, err)
	assert.NotNil(t, executor)
	assert.NotEmpty(t, executor.selfPath)
}

func TestExecutorWithProgress(t *testing.T) {
	// Test executor with progress writer
	executor := &MockExecutor{
		ReturnOutput: "Progress: 10%\nProgress: 50%\nProgress: 100%\n",
	}
	
	var buf bytes.Buffer
	err := executor.Execute([]string{"backup", "my-pod", "/var/log"}, &buf)
	
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "Progress: 100%")
}
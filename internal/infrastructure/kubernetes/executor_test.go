package kubernetes_test

import (
	"context"
	"testing"
	"time"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewOSCommandExecutor(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	assert.NotNil(t, executor)
}

func TestOSCommandExecutor_Execute_EmptyCommand(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	
	_, err := executor.Execute(context.Background(), []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no command provided")
}

func TestOSCommandExecutor_Execute_Success(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	
	// Use echo command which should be available on most systems
	output, err := executor.Execute(context.Background(), []string{"echo", "hello world"})
	assert.NoError(t, err)
	assert.Contains(t, output, "hello world")
}

func TestOSCommandExecutor_Execute_CommandNotFound(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	
	_, err := executor.Execute(context.Background(), []string{"nonexistent-command-xyz"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
}

func TestOSCommandExecutor_Execute_WithTimeout(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	
	// Use a short timeout with a command that would normally take longer
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// sleep command should be available on Unix systems
	_, err := executor.Execute(ctx, []string{"sleep", "1"})
	assert.Error(t, err)
}

func TestOSCommandExecutor_Stream_EmptyCommand(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	
	_, errorCh := executor.Stream(context.Background(), []string{})
	
	// Should receive an error
	select {
	case err := <-errorCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no command provided")
	case <-time.After(time.Second):
		t.Fatal("Expected error but got timeout")
	}
}

func TestOSCommandExecutor_Stream_Success(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	
	outputCh, errorCh := executor.Stream(context.Background(), []string{"echo", "line1"})
	
	// Collect all output
	var outputs []string
	var errors []error
	
	done := false
	for !done {
		select {
		case output, ok := <-outputCh:
			if !ok {
				done = true
			} else {
				outputs = append(outputs, output)
			}
		case err, ok := <-errorCh:
			if ok && err != nil {
				errors = append(errors, err)
			}
		case <-time.After(2 * time.Second):
			done = true
		}
	}
	
	assert.Empty(t, errors, "Should not have errors")
	assert.NotEmpty(t, outputs, "Should have output")
	assert.Contains(t, outputs[0], "line1")
}

func TestOSCommandExecutor_Stream_WithCancellation(t *testing.T) {
	executor := kubernetes.NewOSCommandExecutor()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	_, errorCh := executor.Stream(ctx, []string{"sleep", "10"})
	
	// Cancel immediately
	cancel()
	
	// Should receive cancellation error
	select {
	case err := <-errorCh:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	case <-time.After(2 * time.Second):
		t.Fatal("Expected cancellation error but got timeout")
	}
	
	// Note: We don't test the output channel here since it was not captured
	// due to the cancellation focus of this test
}

func TestEstimateSize_Success(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	// Mock du command output
	mockExecutor.On("Execute", mock.Anything, mock.MatchedBy(func(cmd []string) bool {
		return len(cmd) >= 5 && cmd[len(cmd)-3] == "du" && cmd[len(cmd)-2] == "-sb"
	})).Return("1024000\t/data\n", nil)
	
	size, err := kubernetes.EstimateSize(context.Background(), mockExecutor, "default", "test-pod", "/data")
	
	assert.NoError(t, err)
	assert.Equal(t, int64(1024000), size)
	mockExecutor.AssertExpectations(t)
}

func TestEstimateSize_CommandFailed(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	mockExecutor.On("Execute", mock.Anything, mock.Anything).Return("", assert.AnError)
	
	_, err := kubernetes.EstimateSize(context.Background(), mockExecutor, "default", "test-pod", "/data")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to estimate size")
	mockExecutor.AssertExpectations(t)
}

func TestEstimateSize_InvalidOutput(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	// Mock invalid du output
	mockExecutor.On("Execute", mock.Anything, mock.Anything).Return("invalid output", nil)
	
	_, err := kubernetes.EstimateSize(context.Background(), mockExecutor, "default", "test-pod", "/data")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse size")
	mockExecutor.AssertExpectations(t)
}

func TestEstimateSize_UnparsableSize(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	
	// Mock du output with non-numeric size
	mockExecutor.On("Execute", mock.Anything, mock.Anything).Return("not-a-number\t/data\n", nil)
	
	_, err := kubernetes.EstimateSize(context.Background(), mockExecutor, "default", "test-pod", "/data")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse size")
	mockExecutor.AssertExpectations(t)
}
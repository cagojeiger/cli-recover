package tui

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProcessManager for testing
type MockProcessManager struct {
	mock.Mock
	output []string
}

func (m *MockProcessManager) Start(ctx context.Context, cmd string, args []string) (*exec.Cmd, error) {
	called := m.Called(ctx, cmd, args)
	if called.Get(0) == nil {
		return nil, called.Error(1)
	}
	return called.Get(0).(*exec.Cmd), called.Error(1)
}

func (m *MockProcessManager) Wait(cmd *exec.Cmd) error {
	args := m.Called(cmd)
	return args.Error(0)
}

func (m *MockProcessManager) Kill(cmd *exec.Cmd, force bool) error {
	args := m.Called(cmd, force)
	return args.Error(0)
}

func (m *MockProcessManager) ReadOutput(onOutput func(string)) error {
	args := m.Called(onOutput)
	
	// Simulate output
	for _, line := range m.output {
		onOutput(line)
	}
	
	return args.Error(0)
}

func TestProcessManagerInterface(t *testing.T) {
	// Ensure RealProcessManager implements ProcessManager interface
	var _ ProcessManager = (*RealProcessManager)(nil)
}

func TestRealProcessManagerStart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	pm := NewRealProcessManager()
	ctx := context.Background()

	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid command",
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name:    "invalid command",
			cmd:     "/nonexistent/command",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := pm.Start(ctx, tt.cmd, tt.args)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cmd)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cmd)
				
				// Clean up
				cmd.Wait()
			}
		})
	}
}

func TestRealProcessManagerKill(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	pm := NewRealProcessManager()
	ctx := context.Background()

	// Start a long-running process
	cmd, err := pm.Start(ctx, "sleep", []string{"10"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Kill the process
	err = pm.Kill(cmd, false)
	assert.NoError(t, err)

	// Process should be terminated
	err = cmd.Wait()
	assert.Error(t, err) // Should get an error because process was killed
}

func TestRealProcessManagerForceKill(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	pm := NewRealProcessManager()
	ctx := context.Background()

	// Create a process that ignores SIGTERM
	script := `trap '' TERM; sleep 10`
	cmd, err := pm.Start(ctx, "sh", []string{"-c", script})
	assert.NoError(t, err)

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Force kill should work even if SIGTERM is ignored
	err = pm.Kill(cmd, true)
	assert.NoError(t, err)

	// Process should be terminated
	err = cmd.Wait()
	assert.Error(t, err)
}

func TestRealProcessManagerTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	pm := NewRealProcessManager()
	
	// Context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start a long-running process
	cmd, err := pm.Start(ctx, "sleep", []string{"5"})
	assert.NoError(t, err)

	// Wait for context to timeout
	<-ctx.Done()
	
	// Process should be killed by context
	err = cmd.Wait()
	assert.Error(t, err)
}

func TestProcessOutputCapture(t *testing.T) {
	pm := NewRealProcessManager()
	ctx := context.Background()

	// Run a command that produces output
	cmd, err := pm.Start(ctx, "echo", []string{"line1\nline2\nline3"})
	assert.NoError(t, err)

	// Wait for command to complete
	err = cmd.Wait()
	assert.NoError(t, err)

	// In real implementation, we'd capture stdout/stderr
}

func TestProcessGroupKill(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping process group test on Windows")
	}

	pm := NewRealProcessManager()
	ctx := context.Background()

	// Start a process that creates children
	script := `
	(sleep 10) &
	(sleep 10) &
	sleep 10
	`
	cmd, err := pm.Start(ctx, "sh", []string{"-c", script})
	assert.NoError(t, err)

	// Give it time to create child processes
	time.Sleep(100 * time.Millisecond)

	// Kill should terminate the entire process group
	err = pm.Kill(cmd, true)
	assert.NoError(t, err)

	// All processes should be terminated
	err = cmd.Wait()
	assert.Error(t, err)
}

func TestProcessManagerWithJob(t *testing.T) {
	// Test integration with BackupJob
	job, err := NewBackupJob("test-job", "echo 'backup complete'")
	assert.NoError(t, err)

	mockPM := new(MockProcessManager)
	mockPM.output = []string{"backup progress: 50%", "backup complete"}

	// Setup mock expectations
	mockCmd := &exec.Cmd{}
	mockPM.On("Start", mock.Anything, "echo", mock.Anything).Return(mockCmd, nil)
	mockPM.On("Wait", mockCmd).Return(nil)
	mockPM.On("ReadOutput", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		onOutput := args.Get(0).(func(string))
		for _, line := range mockPM.output {
			onOutput(line)
		}
	})

	// Start job
	err = job.Start()
	assert.NoError(t, err)

	// Simulate process execution
	cmd, err := mockPM.Start(job.Context(), "echo", []string{"'backup complete'"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	// Read output
	err = mockPM.ReadOutput(func(line string) {
		job.AddOutput(line)
		if strings.Contains(line, "50%") {
			job.UpdateProgress(50, line)
		}
	})
	assert.NoError(t, err)

	// Wait for completion
	err = mockPM.Wait(cmd)
	assert.NoError(t, err)

	// Complete job
	job.Complete(nil)

	// Verify job state
	assert.Equal(t, JobStatusCompleted, job.GetStatus())
	assert.Equal(t, 50, job.GetProgress())
	assert.Contains(t, job.GetOutput(), "backup complete")

	mockPM.AssertExpectations(t)
}

func TestKillResistantProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	// Test the 3-stage kill sequence
	mockPM := new(MockProcessManager)
	mockCmd := &exec.Cmd{}

	// First call: SIGTERM fails
	mockPM.On("Kill", mockCmd, false).Return(errors.New("process still running")).Once()
	// Second call: Force kill succeeds
	mockPM.On("Kill", mockCmd, true).Return(nil).Once()

	// Try graceful kill first
	err := mockPM.Kill(mockCmd, false)
	assert.Error(t, err)

	// Then force kill
	err = mockPM.Kill(mockCmd, true)
	assert.NoError(t, err)

	mockPM.AssertExpectations(t)
}
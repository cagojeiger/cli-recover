package kubernetes

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

// MockKubeClient is a mock implementation of KubeClient
type MockKubeClient struct {
	mock.Mock
}

// GetNamespaces mocks the GetNamespaces method
func (m *MockKubeClient) GetNamespaces(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// GetPods mocks the GetPods method
func (m *MockKubeClient) GetPods(ctx context.Context, namespace string) ([]Pod, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]Pod), args.Error(1)
}

// GetContainers mocks the GetContainers method
func (m *MockKubeClient) GetContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	args := m.Called(ctx, namespace, podName)
	return args.Get(0).([]string), args.Error(1)
}

// ExecCommand mocks the ExecCommand method
func (m *MockKubeClient) ExecCommand(ctx context.Context, namespace, podName, container string, command []string) error {
	args := m.Called(ctx, namespace, podName, container, command)
	return args.Error(0)
}

// MockCommandExecutor is a mock implementation of CommandExecutor
type MockCommandExecutor struct {
	mock.Mock
}

// Execute mocks the Execute method
func (m *MockCommandExecutor) Execute(ctx context.Context, command []string) (string, error) {
	args := m.Called(ctx, command)
	return args.String(0), args.Error(1)
}

// Stream mocks the Stream method
func (m *MockCommandExecutor) Stream(ctx context.Context, command []string) (<-chan string, <-chan error) {
	args := m.Called(ctx, command)

	// Handle nil returns
	outputCh := args.Get(0)
	errorCh := args.Get(1)

	if outputCh == nil {
		ch := make(chan string)
		close(ch)
		outputCh = ch
	}

	if errorCh == nil {
		ch := make(chan error)
		close(ch)
		errorCh = ch
	}

	// Type assertions with proper channel types
	outChan, ok := outputCh.(chan string)
	if ok {
		// Convert bidirectional to receive-only
		errChan, errOk := errorCh.(chan error)
		if errOk {
			return (<-chan string)(outChan), (<-chan error)(errChan)
		}
		return (<-chan string)(outChan), errorCh.(<-chan error)
	}

	return outputCh.(<-chan string), errorCh.(<-chan error)
}

// StreamBinary mocks the StreamBinary method
func (m *MockCommandExecutor) StreamBinary(ctx context.Context, command []string) (stdout io.ReadCloser, stderr io.ReadCloser, wait func() error, err error) {
	args := m.Called(ctx, command)

	stdout = args.Get(0).(io.ReadCloser)
	stderr = args.Get(1).(io.ReadCloser)
	wait = args.Get(2).(func() error)
	err = args.Error(3)

	return
}

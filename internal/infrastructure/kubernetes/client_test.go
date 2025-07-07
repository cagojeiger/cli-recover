package kubernetes_test

import (
	"context"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKubeClient is a mock implementation of KubeClient
type MockKubeClient struct {
	mock.Mock
}

func (m *MockKubeClient) GetNamespaces(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockKubeClient) GetPods(ctx context.Context, namespace string) ([]kubernetes.Pod, error) {
	args := m.Called(ctx, namespace)
	if pods := args.Get(0); pods != nil {
		return pods.([]kubernetes.Pod), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockKubeClient) GetContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	args := m.Called(ctx, namespace, podName)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockKubeClient) ExecCommand(ctx context.Context, namespace, podName, container string, command []string) error {
	args := m.Called(ctx, namespace, podName, container, command)
	return args.Error(0)
}

func TestMockKubeClient_GetNamespaces(t *testing.T) {
	mockClient := new(MockKubeClient)
	ctx := context.Background()
	
	expectedNamespaces := []string{"default", "kube-system", "production"}
	mockClient.On("GetNamespaces", ctx).Return(expectedNamespaces, nil)

	namespaces, err := mockClient.GetNamespaces(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedNamespaces, namespaces)
	mockClient.AssertExpectations(t)
}

func TestMockKubeClient_GetPods(t *testing.T) {
	mockClient := new(MockKubeClient)
	ctx := context.Background()
	
	expectedPods := []kubernetes.Pod{
		{Name: "app-1", Status: "Running", Ready: true},
		{Name: "app-2", Status: "Running", Ready: true},
	}
	mockClient.On("GetPods", ctx, "default").Return(expectedPods, nil)

	pods, err := mockClient.GetPods(ctx, "default")
	assert.NoError(t, err)
	assert.Len(t, pods, 2)
	assert.Equal(t, "app-1", pods[0].Name)
	mockClient.AssertExpectations(t)
}

func TestCommandExecutor_Interface(t *testing.T) {
	executor := &MockCommandExecutor{}
	ctx := context.Background()
	
	executor.On("Execute", ctx, []string{"echo", "hello"}).Return("hello\n", nil)

	output, err := executor.Execute(ctx, []string{"echo", "hello"})
	assert.NoError(t, err)
	assert.Equal(t, "hello\n", output)
	executor.AssertExpectations(t)
}

type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Execute(ctx context.Context, command []string) (string, error) {
	args := m.Called(ctx, command)
	return args.String(0), args.Error(1)
}

func (m *MockCommandExecutor) Stream(ctx context.Context, command []string) (<-chan string, <-chan error) {
	args := m.Called(ctx, command)
	return args.Get(0).(<-chan string), args.Get(1).(<-chan error)
}

func TestCommandExecutor_Stream(t *testing.T) {
	executor := &MockCommandExecutor{}
	ctx := context.Background()
	
	outputCh := make(chan string, 2)
	errorCh := make(chan error, 1)
	
	outputCh <- "line1"
	outputCh <- "line2"
	close(outputCh)
	close(errorCh)
	
	executor.On("Stream", ctx, []string{"tail", "-f", "log"}).Return(
		(<-chan string)(outputCh),
		(<-chan error)(errorCh),
	)

	outCh, errCh := executor.Stream(ctx, []string{"tail", "-f", "log"})
	
	var lines []string
	for line := range outCh {
		lines = append(lines, line)
	}
	
	assert.Equal(t, []string{"line1", "line2"}, lines)
	assert.Nil(t, <-errCh) // No error
	executor.AssertExpectations(t)
}
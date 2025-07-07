package kubernetes_test

import (
	"context"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMocksFile_KubeClient_GetNamespaces(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	
	expectedNamespaces := []string{"default", "kube-system", "production"}
	
	// Set up expectation
	mockClient.On("GetNamespaces", ctx).Return(expectedNamespaces, nil)
	
	// Call the method
	namespaces, err := mockClient.GetNamespaces(ctx)
	
	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedNamespaces, namespaces)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_KubeClient_GetNamespaces_Error(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	
	expectedError := assert.AnError
	
	// Set up expectation for error case
	mockClient.On("GetNamespaces", ctx).Return([]string(nil), expectedError)
	
	// Call the method
	namespaces, err := mockClient.GetNamespaces(ctx)
	
	// Assert results
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, namespaces)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_KubeClient_GetPods(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"
	
	expectedPods := []kubernetes.Pod{
		{Name: "pod1", Namespace: "default", Status: "Running", Ready: true},
		{Name: "pod2", Namespace: "default", Status: "Pending", Ready: false},
	}
	
	// Set up expectation
	mockClient.On("GetPods", ctx, namespace).Return(expectedPods, nil)
	
	// Call the method
	pods, err := mockClient.GetPods(ctx, namespace)
	
	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedPods, pods)
	assert.Len(t, pods, 2)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_KubeClient_GetPods_EmptyResult(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "empty-namespace"
	
	expectedPods := []kubernetes.Pod{}
	
	// Set up expectation
	mockClient.On("GetPods", ctx, namespace).Return(expectedPods, nil)
	
	// Call the method
	pods, err := mockClient.GetPods(ctx, namespace)
	
	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedPods, pods)
	assert.Empty(t, pods)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_KubeClient_GetContainers(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"
	podName := "test-pod"
	
	expectedContainers := []string{"web", "sidecar", "proxy"}
	
	// Set up expectation
	mockClient.On("GetContainers", ctx, namespace, podName).Return(expectedContainers, nil)
	
	// Call the method
	containers, err := mockClient.GetContainers(ctx, namespace, podName)
	
	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedContainers, containers)
	assert.Len(t, containers, 3)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_KubeClient_GetContainers_Error(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"
	podName := "nonexistent-pod"
	
	expectedError := assert.AnError
	
	// Set up expectation for error case
	mockClient.On("GetContainers", ctx, namespace, podName).Return([]string(nil), expectedError)
	
	// Call the method
	containers, err := mockClient.GetContainers(ctx, namespace, podName)
	
	// Assert results
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, containers)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_KubeClient_ExecCommand(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"
	podName := "test-pod"
	container := "web"
	command := []string{"ls", "-la", "/var/log"}
	
	// Set up expectation
	mockClient.On("ExecCommand", ctx, namespace, podName, container, command).Return(nil)
	
	// Call the method
	err := mockClient.ExecCommand(ctx, namespace, podName, container, command)
	
	// Assert results
	assert.NoError(t, err)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_KubeClient_ExecCommand_Error(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"
	podName := "test-pod"
	container := "web"
	command := []string{"invalid", "command"}
	
	expectedError := assert.AnError
	
	// Set up expectation for error case
	mockClient.On("ExecCommand", ctx, namespace, podName, container, command).Return(expectedError)
	
	// Call the method
	err := mockClient.ExecCommand(ctx, namespace, podName, container, command)
	
	// Assert results
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	
	// Assert that expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_CommandExecutor_Execute(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	ctx := context.Background()
	command := []string{"echo", "hello world"}
	
	expectedOutput := "hello world\n"
	
	// Set up expectation
	mockExecutor.On("Execute", ctx, command).Return(expectedOutput, nil)
	
	// Call the method
	output, err := mockExecutor.Execute(ctx, command)
	
	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
	
	// Assert that expectations were met
	mockExecutor.AssertExpectations(t)
}

func TestMocksFile_CommandExecutor_Execute_Error(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	ctx := context.Background()
	command := []string{"nonexistent-command"}
	
	expectedError := assert.AnError
	
	// Set up expectation for error case
	mockExecutor.On("Execute", ctx, command).Return("", expectedError)
	
	// Call the method
	output, err := mockExecutor.Execute(ctx, command)
	
	// Assert results
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, output)
	
	// Assert that expectations were met
	mockExecutor.AssertExpectations(t)
}

func TestMocksFile_CommandExecutor_Stream(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	ctx := context.Background()
	command := []string{"tail", "-f", "/var/log/app.log"}
	
	// Create channels for the mock
	outputCh := make(chan string, 2)
	errorCh := make(chan error, 1)
	
	// Send some test data
	outputCh <- "log line 1"
	outputCh <- "log line 2"
	close(outputCh)
	close(errorCh)
	
	// Set up expectation
	mockExecutor.On("Stream", ctx, command).Return((<-chan string)(outputCh), (<-chan error)(errorCh))
	
	// Call the method
	outCh, errCh := mockExecutor.Stream(ctx, command)
	
	// Assert results
	assert.NotNil(t, outCh)
	assert.NotNil(t, errCh)
	
	// Read from channels
	var outputs []string
	var errors []error
	
	// Collect all output
	for output := range outCh {
		outputs = append(outputs, output)
	}
	
	// Collect any errors
	for err := range errCh {
		errors = append(errors, err)
	}
	
	assert.Len(t, outputs, 2)
	assert.Equal(t, "log line 1", outputs[0])
	assert.Equal(t, "log line 2", outputs[1])
	assert.Empty(t, errors)
	
	// Assert that expectations were met
	mockExecutor.AssertExpectations(t)
}

func TestMocksFile_CommandExecutor_Stream_WithError(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	ctx := context.Background()
	command := []string{"invalid", "streaming", "command"}
	
	// Create channels for the mock
	outputCh := make(chan string)
	errorCh := make(chan error, 1)
	
	// Send error and close channels
	errorCh <- assert.AnError
	close(outputCh)
	close(errorCh)
	
	// Set up expectation
	mockExecutor.On("Stream", ctx, command).Return((<-chan string)(outputCh), (<-chan error)(errorCh))
	
	// Call the method
	outCh, errCh := mockExecutor.Stream(ctx, command)
	
	// Assert results
	assert.NotNil(t, outCh)
	assert.NotNil(t, errCh)
	
	// Read from channels
	var outputs []string
	var errors []error
	
	// Collect all output
	for output := range outCh {
		outputs = append(outputs, output)
	}
	
	// Collect any errors
	for err := range errCh {
		errors = append(errors, err)
	}
	
	assert.Empty(t, outputs)
	assert.Len(t, errors, 1)
	assert.Equal(t, assert.AnError, errors[0])
	
	// Assert that expectations were met
	mockExecutor.AssertExpectations(t)
}

func TestMocksFile_KubeClient_MultipleMethodCalls(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	
	// Set up multiple expectations
	mockClient.On("GetNamespaces", ctx).Return([]string{"default"}, nil)
	mockClient.On("GetPods", ctx, "default").Return([]kubernetes.Pod{
		{Name: "pod1", Namespace: "default"},
	}, nil)
	mockClient.On("GetContainers", ctx, "default", "pod1").Return([]string{"web"}, nil)
	mockClient.On("ExecCommand", ctx, "default", "pod1", "web", mock.Anything).Return(nil)
	
	// Call multiple methods
	namespaces, err := mockClient.GetNamespaces(ctx)
	assert.NoError(t, err)
	assert.Len(t, namespaces, 1)
	
	pods, err := mockClient.GetPods(ctx, "default")
	assert.NoError(t, err)
	assert.Len(t, pods, 1)
	
	containers, err := mockClient.GetContainers(ctx, "default", "pod1")
	assert.NoError(t, err)
	assert.Len(t, containers, 1)
	
	err = mockClient.ExecCommand(ctx, "default", "pod1", "web", []string{"ls"})
	assert.NoError(t, err)
	
	// Assert that all expectations were met
	mockClient.AssertExpectations(t)
}

func TestMocksFile_CommandExecutor_MultipleExecutions(t *testing.T) {
	mockExecutor := new(kubernetes.MockCommandExecutor)
	ctx := context.Background()
	
	// Set up multiple expectations
	mockExecutor.On("Execute", ctx, []string{"echo", "first"}).Return("first\n", nil)
	mockExecutor.On("Execute", ctx, []string{"echo", "second"}).Return("second\n", nil)
	
	// Call multiple times
	output1, err1 := mockExecutor.Execute(ctx, []string{"echo", "first"})
	assert.NoError(t, err1)
	assert.Equal(t, "first\n", output1)
	
	output2, err2 := mockExecutor.Execute(ctx, []string{"echo", "second"})
	assert.NoError(t, err2)
	assert.Equal(t, "second\n", output2)
	
	// Assert that all expectations were met
	mockExecutor.AssertExpectations(t)
}
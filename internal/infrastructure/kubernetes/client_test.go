package kubernetes_test

import (
	"context"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/stretchr/testify/assert"
)

func TestMockKubeClient_GetNamespaces(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()

	expectedNamespaces := []string{"default", "kube-system", "production"}
	mockClient.On("GetNamespaces", ctx).Return(expectedNamespaces, nil)

	namespaces, err := mockClient.GetNamespaces(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedNamespaces, namespaces)
	mockClient.AssertExpectations(t)
}

func TestMockKubeClient_GetPods(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"

	expectedPods := []kubernetes.Pod{
		{Name: "pod1", Namespace: namespace, Status: "Running", Ready: true},
		{Name: "pod2", Namespace: namespace, Status: "Running", Ready: true},
	}
	mockClient.On("GetPods", ctx, namespace).Return(expectedPods, nil)

	pods, err := mockClient.GetPods(ctx, namespace)
	assert.NoError(t, err)
	assert.Equal(t, expectedPods, pods)
	mockClient.AssertExpectations(t)
}

func TestMockKubeClient_GetContainers(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"
	podName := "test-pod"

	expectedContainers := []string{"container1", "container2"}
	mockClient.On("GetContainers", ctx, namespace, podName).Return(expectedContainers, nil)

	containers, err := mockClient.GetContainers(ctx, namespace, podName)
	assert.NoError(t, err)
	assert.Equal(t, expectedContainers, containers)
	mockClient.AssertExpectations(t)
}

func TestMockKubeClient_ExecCommand(t *testing.T) {
	mockClient := new(kubernetes.MockKubeClient)
	ctx := context.Background()
	namespace := "default"
	podName := "test-pod"
	container := "container1"
	command := []string{"ls", "-la"}

	mockClient.On("ExecCommand", ctx, namespace, podName, container, command).Return(nil)

	err := mockClient.ExecCommand(ctx, namespace, podName, container, command)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

package kubernetes_test

import (
	"context"
	"strings"
	"testing"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/stretchr/testify/assert"
)

// MockExecutor for testing kubectl wrapper
type MockExecutor struct {
	outputs map[string]string
	errors  map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		outputs: make(map[string]string),
		errors:  make(map[string]error),
	}
}

func (m *MockExecutor) Execute(ctx context.Context, command []string) (string, error) {
	key := strings.Join(command, " ")
	if err, exists := m.errors[key]; exists {
		return "", err
	}
	return m.outputs[key], nil
}

func (m *MockExecutor) Stream(ctx context.Context, command []string) (<-chan string, <-chan error) {
	outputCh := make(chan string, 1)
	errorCh := make(chan error, 1)
	
	output, err := m.Execute(ctx, command)
	if err != nil {
		errorCh <- err
	} else {
		outputCh <- output
	}
	close(outputCh)
	close(errorCh)
	
	return outputCh, errorCh
}

func TestKubectlClient_GetNamespaces(t *testing.T) {
	executor := NewMockExecutor()
	executor.outputs["kubectl get namespaces -o json"] = `{
		"items": [
			{"metadata": {"name": "default"}},
			{"metadata": {"name": "kube-system"}},
			{"metadata": {"name": "production"}}
		]
	}`

	client := kubernetes.NewKubectlClient(executor)
	ctx := context.Background()

	namespaces, err := client.GetNamespaces(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"default", "kube-system", "production"}, namespaces)
}

func TestKubectlClient_GetPods(t *testing.T) {
	executor := NewMockExecutor()
	executor.outputs["kubectl get pods -n default -o json"] = `{
		"items": [
			{
				"metadata": {"name": "app-1", "namespace": "default"},
				"status": {
					"phase": "Running",
					"conditions": [{"type": "Ready", "status": "True"}]
				},
				"spec": {"nodeName": "node-1"}
			},
			{
				"metadata": {"name": "app-2", "namespace": "default"},
				"status": {
					"phase": "Pending",
					"conditions": [{"type": "Ready", "status": "False"}]
				},
				"spec": {"nodeName": "node-2"}
			}
		]
	}`

	client := kubernetes.NewKubectlClient(executor)
	ctx := context.Background()

	pods, err := client.GetPods(ctx, "default")
	assert.NoError(t, err)
	assert.Len(t, pods, 2)
	
	assert.Equal(t, "app-1", pods[0].Name)
	assert.Equal(t, "Running", pods[0].Status)
	assert.True(t, pods[0].Ready)
	
	assert.Equal(t, "app-2", pods[1].Name)
	assert.Equal(t, "Pending", pods[1].Status)
	assert.False(t, pods[1].Ready)
}

func TestKubectlClient_GetContainers(t *testing.T) {
	executor := NewMockExecutor()
	executor.outputs["kubectl get pod test-pod -n default -o json"] = `{
		"spec": {
			"containers": [
				{"name": "main"},
				{"name": "sidecar"}
			]
		}
	}`

	client := kubernetes.NewKubectlClient(executor)
	ctx := context.Background()

	containers, err := client.GetContainers(ctx, "default", "test-pod")
	assert.NoError(t, err)
	assert.Equal(t, []string{"main", "sidecar"}, containers)
}

func TestBuildKubectlCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "get namespaces",
			args:     []string{"get", "namespaces", "-o", "json"},
			expected: []string{"kubectl", "get", "namespaces", "-o", "json"},
		},
		{
			name:     "exec command",
			args:     []string{"exec", "-n", "default", "pod-1", "--", "ls", "-la"},
			expected: []string{"kubectl", "exec", "-n", "default", "pod-1", "--", "ls", "-la"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kubernetes.BuildKubectlCommand(tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
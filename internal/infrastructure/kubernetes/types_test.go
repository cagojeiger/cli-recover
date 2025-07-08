package kubernetes_test

import (
	"testing"
	"time"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/kubernetes"
	"github.com/stretchr/testify/assert"
)

func TestPod(t *testing.T) {
	pod := kubernetes.Pod{
		Name:      "test-pod",
		Namespace: "default",
		Status:    "Running",
		Ready:     true,
		Age:       time.Duration(24 * time.Hour),
		Node:      "worker-node-1",
	}

	assert.Equal(t, "test-pod", pod.Name)
	assert.Equal(t, "default", pod.Namespace)
	assert.Equal(t, "Running", pod.Status)
	assert.True(t, pod.Ready)
	assert.Equal(t, time.Duration(24*time.Hour), pod.Age)
	assert.Equal(t, "worker-node-1", pod.Node)
}

func TestPod_EmptyValues(t *testing.T) {
	pod := kubernetes.Pod{}

	assert.Empty(t, pod.Name)
	assert.Empty(t, pod.Namespace)
	assert.Empty(t, pod.Status)
	assert.False(t, pod.Ready)
	assert.Equal(t, time.Duration(0), pod.Age)
	assert.Empty(t, pod.Node)
}

func TestPod_MultipleStates(t *testing.T) {
	tests := []struct {
		name     string
		pod      kubernetes.Pod
		expected struct {
			status string
			ready  bool
		}
	}{
		{
			name: "running and ready",
			pod: kubernetes.Pod{
				Name:   "running-pod",
				Status: "Running",
				Ready:  true,
			},
			expected: struct {
				status string
				ready  bool
			}{"Running", true},
		},
		{
			name: "running but not ready",
			pod: kubernetes.Pod{
				Name:   "starting-pod",
				Status: "Running",
				Ready:  false,
			},
			expected: struct {
				status string
				ready  bool
			}{"Running", false},
		},
		{
			name: "pending",
			pod: kubernetes.Pod{
				Name:   "pending-pod",
				Status: "Pending",
				Ready:  false,
			},
			expected: struct {
				status string
				ready  bool
			}{"Pending", false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.status, tt.pod.Status)
			assert.Equal(t, tt.expected.ready, tt.pod.Ready)
		})
	}
}

func TestContainer(t *testing.T) {
	container := kubernetes.Container{
		Name:  "web-server",
		Image: "nginx:1.21",
		Ready: true,
	}

	assert.Equal(t, "web-server", container.Name)
	assert.Equal(t, "nginx:1.21", container.Image)
	assert.True(t, container.Ready)
}

func TestContainer_EmptyValues(t *testing.T) {
	container := kubernetes.Container{}

	assert.Empty(t, container.Name)
	assert.Empty(t, container.Image)
	assert.False(t, container.Ready)
}

func TestContainer_MultipleContainers(t *testing.T) {
	containers := []kubernetes.Container{
		{
			Name:  "nginx",
			Image: "nginx:latest",
			Ready: true,
		},
		{
			Name:  "sidecar",
			Image: "busybox:1.35",
			Ready: false,
		},
	}

	assert.Len(t, containers, 2)
	assert.Equal(t, "nginx", containers[0].Name)
	assert.Equal(t, "nginx:latest", containers[0].Image)
	assert.True(t, containers[0].Ready)

	assert.Equal(t, "sidecar", containers[1].Name)
	assert.Equal(t, "busybox:1.35", containers[1].Image)
	assert.False(t, containers[1].Ready)
}

// Tests for type structures only - interface tests are in client_test.go

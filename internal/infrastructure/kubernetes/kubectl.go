package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
)

// KubectlClient implements KubeClient using kubectl
type KubectlClient struct {
	executor CommandExecutor
}

// NewKubectlClient creates a new kubectl-based client
func NewKubectlClient(executor CommandExecutor) *KubectlClient {
	return &KubectlClient{
		executor: executor,
	}
}

// GetNamespaces returns all namespaces
func (k *KubectlClient) GetNamespaces(ctx context.Context) ([]string, error) {
	output, err := k.executor.Execute(ctx, BuildKubectlCommand("get", "namespaces", "-o", "json"))
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse namespaces: %w", err)
	}

	namespaces := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		namespaces = append(namespaces, item.Metadata.Name)
	}

	return namespaces, nil
}

// GetPods returns pods in a namespace
func (k *KubectlClient) GetPods(ctx context.Context, namespace string) ([]Pod, error) {
	output, err := k.executor.Execute(ctx, BuildKubectlCommand("get", "pods", "-n", namespace, "-o", "json"))
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Status struct {
				Phase      string `json:"phase"`
				Conditions []struct {
					Type   string `json:"type"`
					Status string `json:"status"`
				} `json:"conditions"`
			} `json:"status"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse pods: %w", err)
	}

	pods := make([]Pod, 0, len(result.Items))
	for _, item := range result.Items {
		pod := Pod{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Status:    item.Status.Phase,
			Node:      item.Spec.NodeName,
		}

		// Check if pod is ready
		for _, condition := range item.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				pod.Ready = true
				break
			}
		}

		pods = append(pods, pod)
	}

	return pods, nil
}

// GetContainers returns container names in a pod
func (k *KubectlClient) GetContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	output, err := k.executor.Execute(ctx, BuildKubectlCommand("get", "pod", podName, "-n", namespace, "-o", "json"))
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	var result struct {
		Spec struct {
			Containers []struct {
				Name string `json:"name"`
			} `json:"containers"`
		} `json:"spec"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse pod: %w", err)
	}

	containers := make([]string, 0, len(result.Spec.Containers))
	for _, container := range result.Spec.Containers {
		containers = append(containers, container.Name)
	}

	return containers, nil
}

// ExecCommand executes a command in a container
func (k *KubectlClient) ExecCommand(ctx context.Context, namespace, podName, container string, command []string) error {
	args := []string{"exec", "-n", namespace, podName}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, "--")
	args = append(args, command...)

	_, err := k.executor.Execute(ctx, BuildKubectlCommand(args...))
	if err != nil {
		return fmt.Errorf("failed to exec command: %w", err)
	}

	return nil
}

// BuildKubectlCommand builds a kubectl command with arguments
func BuildKubectlCommand(args ...string) []string {
	cmd := []string{"kubectl"}
	cmd = append(cmd, args...)
	return cmd
}

package kubectl

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PodInfo represents basic pod information
type PodInfo struct {
	Name      string
	Namespace string
	Status    string
	Ready     string
}

// GetNamespaces returns list of available namespaces
func GetNamespaces() ([]string, error) {
	cmd := exec.Command("kubectl", "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	namespaces := strings.Fields(string(output))
	if len(namespaces) == 0 {
		return []string{"default"}, nil
	}
	return namespaces, nil
}

// GetPods returns list of pods in the specified namespace
func GetPods(namespace string) ([]PodInfo, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pods in namespace %s: %w", namespace, err)
	}

	var podList struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
				ContainerStatuses []struct {
					Ready bool `json:"ready"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podList); err != nil {
		return nil, fmt.Errorf("failed to parse pod list: %w", err)
	}

	var pods []PodInfo
	for _, item := range podList.Items {
		ready := 0
		total := len(item.Status.ContainerStatuses)
		for _, container := range item.Status.ContainerStatuses {
			if container.Ready {
				ready++
			}
		}

		pods = append(pods, PodInfo{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Status:    item.Status.Phase,
			Ready:     fmt.Sprintf("%d/%d", ready, total),
		})
	}

	return pods, nil
}

// CheckKubectl verifies kubectl is installed and accessible
func CheckKubectl() error {
	cmd := exec.Command("kubectl", "version", "--client", "--short")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubectl not found or not accessible: %w", err)
	}
	return nil
}

// CheckClusterAccess verifies cluster connectivity
func CheckClusterAccess() error {
	cmd := exec.Command("kubectl", "cluster-info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot access Kubernetes cluster: %w", err)
	}
	return nil
}

// GetCommonPaths returns common backup paths
func GetCommonPaths() []string {
	return []string{
		"/data",
		"/logs",
		"/config",
		"/app",
		"/var/lib",
		"/usr/share",
		"/etc",
		"Custom path...",
	}
}
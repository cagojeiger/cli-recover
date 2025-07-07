package kubernetes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cagojeiger/cli-recover/internal/runner"
)

// GetNamespaces returns list of available namespaces
func GetNamespaces(runner runner.Runner) ([]string, error) {
	output, err := runner.Run("kubectl", "get", "namespaces", "-o", "json")
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

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse namespaces: %w", err)
	}

	namespaces := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		namespaces = append(namespaces, item.Metadata.Name)
	}

	return namespaces, nil
}

// GetPods returns list of pods in namespace
func GetPods(runner runner.Runner, namespace string) ([]Pod, error) {
	output, err := runner.Run("kubectl", "get", "pods", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Spec struct {
				Containers []struct {
					Name string `json:"name"`
				} `json:"containers"`
			} `json:"spec"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					Ready bool   `json:"ready"`
					Name  string `json:"name"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pods: %w", err)
	}

	pods := make([]Pod, 0, len(result.Items))
	for _, item := range result.Items {
		ready := 0
		total := len(item.Status.ContainerStatuses)
		for _, container := range item.Status.ContainerStatuses {
			if container.Ready {
				ready++
			}
		}

		// Extract container names from spec
		var containers []string
		for _, container := range item.Spec.Containers {
			containers = append(containers, container.Name)
		}

		pods = append(pods, Pod{
			Name:       item.Metadata.Name,
			Namespace:  item.Metadata.Namespace,
			Status:     item.Status.Phase,
			Ready:      fmt.Sprintf("%d/%d", ready, total),
			Containers: containers,
		})
	}

	return pods, nil
}

// GetDirectoryContents returns list of files and directories in the pod's path
func GetDirectoryContents(runner runner.Runner, pod, namespace, path, container string) ([]DirectoryEntry, error) {
	args := []string{"exec", "-n", namespace}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, pod, "--", "ls", "-la", path)
	
	output, err := runner.Run("kubectl", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var entries []DirectoryEntry
	
	for _, line := range lines {
		// Skip empty lines and total line
		if line == "" || strings.HasPrefix(line, "total") {
			continue
		}
		
		// Parse ls -la output
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		
		// Skip . and .. entries
		name := fields[8]
		if name == "." || name == ".." {
			continue
		}
		
		entryType := "file"
		if strings.HasPrefix(fields[0], "d") {
			entryType = "dir"
		}
		
		entries = append(entries, DirectoryEntry{
			Name:     name,
			Type:     entryType,
			Size:     fields[4],
			Modified: strings.Join(fields[5:8], " "),
		})
	}
	
	return entries, nil
}
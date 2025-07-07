package kubernetes

import (
	"context"
	"io"
	"time"
)

// Pod represents a Kubernetes pod
type Pod struct {
	Name      string
	Namespace string
	Status    string
	Ready     bool
	Age       time.Duration
	Node      string
}

// Container represents a container in a pod
type Container struct {
	Name  string
	Image string
	Ready bool
}

// KubeClient defines the interface for Kubernetes operations
type KubeClient interface {
	// GetNamespaces returns all namespaces
	GetNamespaces(ctx context.Context) ([]string, error)
	
	// GetPods returns pods in a namespace
	GetPods(ctx context.Context, namespace string) ([]Pod, error)
	
	// GetContainers returns container names in a pod
	GetContainers(ctx context.Context, namespace, podName string) ([]string, error)
	
	// ExecCommand executes a command in a container
	ExecCommand(ctx context.Context, namespace, podName, container string, command []string) error
}

// CommandExecutor defines the interface for executing commands
type CommandExecutor interface {
	// Execute runs a command and returns the output
	Execute(ctx context.Context, command []string) (string, error)
	
	// Stream runs a command and streams the output
	Stream(ctx context.Context, command []string) (<-chan string, <-chan error)
	
	// StreamBinary runs a command and streams binary output safely
	// Returns stdout, stderr readers and a wait function for command completion
	StreamBinary(ctx context.Context, command []string) (stdout io.ReadCloser, stderr io.ReadCloser, wait func() error, err error)
}
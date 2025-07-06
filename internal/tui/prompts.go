package tui

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cagojeiger/cli-restore/internal/kubectl"
)

// BackupOptions represents user-selected backup options
type BackupOptions struct {
	Namespace string
	Pod       string
	Path      string
	SplitSize string
}

// RunInteractiveBackup runs the interactive TUI for backup configuration
func RunInteractiveBackup() (*BackupOptions, error) {
	options := &BackupOptions{}

	// Check dependencies first
	if err := kubectl.CheckKubectl(); err != nil {
		return nil, fmt.Errorf("dependency check failed: %w", err)
	}

	if err := kubectl.CheckClusterAccess(); err != nil {
		return nil, fmt.Errorf("cluster access check failed: %w", err)
	}

	// Step 1: Select namespace
	namespaces, err := kubectl.GetNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	prompt := &survey.Select{
		Message: "Select namespace:",
		Options: namespaces,
		Default: "default",
	}
	if err := survey.AskOne(prompt, &options.Namespace); err != nil {
		return nil, fmt.Errorf("namespace selection failed: %w", err)
	}

	// Step 2: Select pod
	pods, err := kubectl.GetPods(options.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	if len(pods) == 0 {
		return nil, fmt.Errorf("no pods found in namespace %s", options.Namespace)
	}

	// Format pod options with status
	var podOptions []string
	for _, pod := range pods {
		podOptions = append(podOptions, fmt.Sprintf("%-20s (%s, %s ready)", pod.Name, pod.Status, pod.Ready))
	}

	var selectedPodDisplay string
	prompt = &survey.Select{
		Message: "Select pod:",
		Options: podOptions,
	}
	if err := survey.AskOne(prompt, &selectedPodDisplay); err != nil {
		return nil, fmt.Errorf("pod selection failed: %w", err)
	}

	// Extract pod name from display string
	options.Pod = strings.Fields(selectedPodDisplay)[0]

	// Step 3: Select path
	pathOptions := kubectl.GetCommonPaths()
	var selectedPath string
	prompt = &survey.Select{
		Message: "Select path to backup:",
		Options: pathOptions,
	}
	if err := survey.AskOne(prompt, &selectedPath); err != nil {
		return nil, fmt.Errorf("path selection failed: %w", err)
	}

	if selectedPath == "Custom path..." {
		input := &survey.Input{
			Message: "Enter custom path:",
			Default: "/data",
		}
		if err := survey.AskOne(input, &options.Path); err != nil {
			return nil, fmt.Errorf("custom path input failed: %w", err)
		}
	} else {
		options.Path = selectedPath
	}

	// Step 4: Select split size
	sizeOptions := []string{"1G", "2G", "5G", "Custom..."}
	var selectedSize string
	prompt = &survey.Select{
		Message: "Split size:",
		Options: sizeOptions,
		Default: "1G",
	}
	if err := survey.AskOne(prompt, &selectedSize); err != nil {
		return nil, fmt.Errorf("split size selection failed: %w", err)
	}

	if selectedSize == "Custom..." {
		input := &survey.Input{
			Message: "Enter custom split size (e.g., 500M, 2G):",
			Default: "1G",
		}
		if err := survey.AskOne(input, &options.SplitSize); err != nil {
			return nil, fmt.Errorf("custom split size input failed: %w", err)
		}
	} else {
		options.SplitSize = selectedSize
	}

	// Step 5: Confirm settings
	if err := confirmSettings(options); err != nil {
		return nil, err
	}

	return options, nil
}

// confirmSettings shows final confirmation and CLI command
func confirmSettings(options *BackupOptions) error {
	cliCommand := fmt.Sprintf("cli-restore backup %s %s --namespace %s --split-size %s",
		options.Pod, options.Path, options.Namespace, options.SplitSize)

	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("Backup Settings:\n")
	fmt.Printf("  Pod:       %s\n", options.Pod)
	fmt.Printf("  Namespace: %s\n", options.Namespace)
	fmt.Printf("  Path:      %s\n", options.Path)
	fmt.Printf("  Split:     %s\n", options.SplitSize)
	fmt.Printf("\nGenerated CLI command:\n")
	fmt.Printf("  %s\n", cliCommand)
	fmt.Printf(strings.Repeat("=", 50) + "\n\n")

	var choice string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Execute backup now",
			"Show CLI command only",
			"Cancel",
		},
		Default: "Execute backup now",
	}

	if err := survey.AskOne(prompt, &choice); err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}

	switch choice {
	case "Execute backup now":
		return nil
	case "Show CLI command only":
		fmt.Printf("\nCLI Command:\n%s\n\n", cliCommand)
		return fmt.Errorf("showing CLI command only - no backup executed")
	case "Cancel":
		return fmt.Errorf("backup cancelled by user")
	default:
		return fmt.Errorf("invalid choice")
	}
}
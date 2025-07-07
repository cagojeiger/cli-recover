package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/cagojeiger/cli-recover/internal/domain/metadata"
	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// ListAdapter handles listing backup metadata
type ListAdapter struct {
	store metadata.Store
}

// NewListAdapter creates a new list adapter
func NewListAdapter() *ListAdapter {
	return &ListAdapter{
		store: metadata.DefaultStore,
	}
}

// ExecuteList executes the list command
func (a *ListAdapter) ExecuteList(cmd *cobra.Command, args []string) error {
	// Get flags
	namespace, _ := cmd.Flags().GetString("namespace")
	outputFormat, _ := cmd.Flags().GetString("output")
	showDetails, _ := cmd.Flags().GetBool("details")

	// Retrieve metadata
	var metadataList []*restore.Metadata
	var err error

	if namespace != "" {
		metadataList, err = a.store.ListByNamespace(namespace)
	} else {
		metadataList, err = a.store.List()
	}

	if err != nil {
		return fmt.Errorf("failed to retrieve backups: %w", err)
	}

	// Handle empty list
	if len(metadataList) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	// Output based on format
	switch outputFormat {
	case "json":
		return a.outputJSON(metadataList)
	case "yaml":
		return a.outputYAML(metadataList)
	case "table":
		if showDetails {
			return a.outputDetails(metadataList)
		}
		return a.outputTable(metadataList)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// outputTable outputs metadata in table format
func (a *ListAdapter) outputTable(metadataList []*restore.Metadata) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "ID\tType\tNamespace\tPod\tPath\tSize\tCreated")
	fmt.Fprintln(w, strings.Repeat("-", 80))

	// Rows
	for _, m := range metadataList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			m.ID,
			m.Type,
			m.Namespace,
			m.PodName,
			m.SourcePath,
			a.formatSize(m.Size),
			m.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	// Summary
	fmt.Fprintf(w, "\nTotal: %d backup", len(metadataList))
	if len(metadataList) != 1 {
		fmt.Fprint(w, "s")
	}
	fmt.Fprintln(w)

	return nil
}

// outputDetails outputs detailed metadata information
func (a *ListAdapter) outputDetails(metadataList []*restore.Metadata) error {
	for i, m := range metadataList {
		if i > 0 {
			fmt.Println(strings.Repeat("-", 60))
		}

		fmt.Printf("Backup ID:    %s\n", m.ID)
		fmt.Printf("Type:         %s\n", m.Type)
		fmt.Printf("Namespace:    %s\n", m.Namespace)
		fmt.Printf("Pod:          %s\n", m.PodName)
		fmt.Printf("Source Path:  %s\n", m.SourcePath)
		fmt.Printf("Backup File:  %s\n", m.BackupFile)
		fmt.Printf("Size:         %s\n", a.formatSize(m.Size))
		fmt.Printf("Compression:  %s\n", m.Compression)
		fmt.Printf("Status:       %s\n", m.Status)
		fmt.Printf("Created:      %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Completed:    %s\n", m.CompletedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Duration:     %s\n", a.formatDuration(m.CreatedAt, m.CompletedAt))
		
		if m.Checksum != "" {
			fmt.Printf("Checksum:     %s\n", m.Checksum)
		}

		// Provider-specific info
		if len(m.ProviderInfo) > 0 {
			fmt.Println("Provider Info:")
			for k, v := range m.ProviderInfo {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}

	return nil
}

// outputJSON outputs metadata in JSON format
func (a *ListAdapter) outputJSON(metadataList []*restore.Metadata) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadataList)
}

// outputYAML outputs metadata in YAML format
func (a *ListAdapter) outputYAML(metadataList []*restore.Metadata) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(metadataList)
}

// formatSize formats bytes to human-readable format
func (a *ListAdapter) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats duration between two times
func (a *ListAdapter) formatDuration(start, end time.Time) string {
	duration := end.Sub(start)
	return duration.Round(time.Second).String()
}
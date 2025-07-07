package kubernetes

import (
	"fmt"
	"strings"
)

// GenerateBackupCommand creates the kubectl backup command with options
func GenerateBackupCommand(pod, namespace, path string, options BackupOptions) string {
	var tarFlags []string
	
	// Compression flags
	switch options.CompressionType {
	case "gzip":
		tarFlags = append(tarFlags, "-z")
	case "bzip2":
		tarFlags = append(tarFlags, "-j")
	case "xz":
		tarFlags = append(tarFlags, "-J")
	case "none":
		// No compression flag
	}
	
	// Basic flags
	tarFlags = append(tarFlags, "-c") // create
	tarFlags = append(tarFlags, "-f", "-") // file to stdout
	
	// Advanced options
	if options.Verbose {
		tarFlags = append(tarFlags, "--verbose")
	}
	if options.ShowTotals {
		tarFlags = append(tarFlags, "--totals")
	}
	if options.PreservePerms {
		tarFlags = append(tarFlags, "--preserve-permissions")
	}
	
	// Exclude patterns
	for _, pattern := range options.ExcludePatterns {
		tarFlags = append(tarFlags, "--exclude="+pattern)
	}
	if options.ExcludeVCS {
		tarFlags = append(tarFlags, "--exclude-vcs")
	}
	
	// Build command
	containerFlag := ""
	if options.Container != "" {
		containerFlag = fmt.Sprintf("-c %s ", options.Container)
	}
	
	cmd := fmt.Sprintf("kubectl exec -n %s %s %s-- tar %s %s",
		namespace, pod, containerFlag, strings.Join(tarFlags, " "), path)
	
	return cmd
}
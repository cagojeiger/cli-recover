package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version will be set by ldflags during build
var version = "dev"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cli-restore",
		Short: "Kubernetes Pod backup utility",
		Long: `CLI-Restore is a tool for backing up files and directories from Kubernetes pods.
It creates tar archives with optional splitting for large files.`,
		Version: version,
	}

	// Customize version template to show only version string
	rootCmd.SetVersionTemplate("cli-restore version {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
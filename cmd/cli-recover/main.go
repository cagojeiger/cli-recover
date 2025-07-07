package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cagojeiger/cli-recover/internal/application/config"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/logger"
)

// version will be set by ldflags during build
var version = "dev"

// expandPath is a helper to expand ~ in paths
func expandPath(path string) string {
	if path == "" {
		return path
	}
	
	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	
	return path
}

func main() {
	var rootCmd = &cobra.Command{
		Use:     "cli-recover",
		Short:   "Kubernetes integrated backup and restore tool",
		Long:    `CLI-Recover provides backup and restore capabilities for Kubernetes environments including pod filesystems, databases, and object storage.`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip config loading for init command
			if cmd.Name() == "init" {
				return
			}
			
			// Load configuration from file
			loader := config.NewLoader(config.ConfigPath())
			appConfig, err := loader.Load()
			if err != nil {
				// If config doesn't exist, use defaults
				appConfig = config.DefaultConfig()
			}
			
			// Override with flags (flags have highest priority)
			logLevel, _ := cmd.Flags().GetString("log-level")
			logFile, _ := cmd.Flags().GetString("log-file")
			logFormat, _ := cmd.Flags().GetString("log-format")
			
			// Convert app config to logger config
			loggerCfg := logger.Config{
				Level:      appConfig.Logger.Level,
				Output:     appConfig.Logger.Output,
				FilePath:   expandPath(appConfig.Logger.File.Path),
				MaxSize:    int64(appConfig.Logger.File.MaxSize),
				MaxAge:     appConfig.Logger.File.MaxAge,
				JSONFormat: appConfig.Logger.File.Format == "json",
				UseColor:   appConfig.Logger.Console.Color,
			}
			
			// Apply flag overrides
			if logLevel != "" {
				loggerCfg.Level = logLevel
			}
			if logFile != "" {
				loggerCfg.Output = "both"
				loggerCfg.FilePath = logFile
			}
			if logFormat == "json" {
				loggerCfg.JSONFormat = true
			}
			
			// Set logger level based on debug flag
			debug, _ := cmd.Flags().GetBool("debug")
			if debug && loggerCfg.Level == "info" {
				loggerCfg.Level = "debug"
			}
			
			// Initialize global logger
			if err := logger.InitializeFromConfig(loggerCfg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			}
			
			// Store loaded config for use by commands
			cmd.SetContext(config.WithConfig(cmd.Context(), appConfig))
		},
		Run: func(cmd *cobra.Command, args []string) {
			// If no args, show help
			cmd.Help()
		},
	}

	// Customize version template
	rootCmd.SetVersionTemplate("cli-recover version {{.Version}}\n")
	
	// Add global flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-file", "", "Log file path (logs to console if not specified)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (text, json)")

	// Add new provider-based backup command (recommended)
	rootCmd.AddCommand(newBackupCommand())
	
	// Add new provider-based restore command
	rootCmd.AddCommand(newRestoreCommand())
	
	// Add list command
	rootCmd.AddCommand(newListCommand())
	
	// Add init command
	rootCmd.AddCommand(newInitCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
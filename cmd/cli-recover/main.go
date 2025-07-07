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
	rootCmd := createRootCommand()
	setupPersistentPreRun(rootCmd)
	addGlobalFlags(rootCmd)
	addSubcommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "cli-recover",
		Short:   "Kubernetes integrated backup and restore tool",
		Long:    `CLI-Recover provides backup and restore capabilities for Kubernetes environments including pod filesystems, databases, and object storage.`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	
	rootCmd.SetVersionTemplate("cli-recover version {{.Version}}\n")
	return rootCmd
}

func setupPersistentPreRun(rootCmd *cobra.Command) {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "init" {
			return
		}
		
		appConfig := loadAppConfig()
		loggerCfg := buildLoggerConfig(cmd, appConfig)
		
		if err := logger.InitializeFromConfig(loggerCfg); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		}
		
		cmd.SetContext(config.WithConfig(cmd.Context(), appConfig))
	}
}

func loadAppConfig() *config.Config {
	loader := config.NewLoader(config.ConfigPath())
	appConfig, err := loader.Load()
	if err != nil {
		appConfig = config.DefaultConfig()
	}
	return appConfig
}

func buildLoggerConfig(cmd *cobra.Command, appConfig *config.Config) logger.Config {
	logLevel, _ := cmd.Flags().GetString("log-level")
	logFile, _ := cmd.Flags().GetString("log-file")
	logFormat, _ := cmd.Flags().GetString("log-format")
	debug, _ := cmd.Flags().GetBool("debug")
	
	loggerCfg := logger.Config{
		Level:      appConfig.Logger.Level,
		Output:     appConfig.Logger.Output,
		FilePath:   expandPath(appConfig.Logger.File.Path),
		MaxSize:    int64(appConfig.Logger.File.MaxSize),
		MaxAge:     appConfig.Logger.File.MaxAge,
		JSONFormat: appConfig.Logger.File.Format == "json",
		UseColor:   appConfig.Logger.Console.Color,
	}
	
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
	if debug && loggerCfg.Level == "info" {
		loggerCfg.Level = "debug"
	}
	
	return loggerCfg
}

func addGlobalFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-file", "", "Log file path (logs to console if not specified)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (text, json)")
}

func addSubcommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newBackupCommand())
	rootCmd.AddCommand(newRestoreCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newInitCommand())
}
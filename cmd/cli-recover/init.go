package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cagojeiger/cli-recover/internal/infrastructure/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newInitCommand creates the init command
func newInitCommand() *cobra.Command {
	var (
		showConfig  bool
		resetConfig bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize CLI configuration",
		Long: `Initialize CLI-Recover configuration and directories.

This command will:
- Create configuration directory (~/.cli-recover)
- Generate default configuration file
- Set up log and metadata directories
- Create initial settings for logger, backup, and metadata`,
		Example: `  # Initialize with default settings
  cli-recover init

  # Show current configuration
  cli-recover init --show

  # Reset configuration to defaults
  cli-recover init --reset

  # Interactive configuration setup
  cli-recover init --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := config.ConfigPath()

			// Show configuration
			if showConfig {
				return showConfiguration(configPath)
			}

			// Reset configuration
			if resetConfig {
				return resetConfiguration(configPath)
			}

			// Interactive setup
			if interactive {
				return interactiveSetup(configPath)
			}

			// Default: create configuration
			return createConfiguration(configPath)
		},
	}

	cmd.Flags().BoolVar(&showConfig, "show", false, "Show current configuration")
	cmd.Flags().BoolVar(&resetConfig, "reset", false, "Reset configuration to defaults")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive configuration setup")

	return cmd
}

// createConfiguration creates the default configuration
func createConfiguration(configPath string) error {
	writer := config.NewWriter(configPath)

	// Check if config already exists
	if writer.Exists() {
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		fmt.Println("Use --reset to overwrite with defaults")
		return nil
	}

	// Create default configuration
	if err := writer.CreateDefaultConfig(); err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	// Create necessary directories
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".cli-recover")

	directories := []string{
		filepath.Join(baseDir, "logs"),
		filepath.Join(baseDir, "metadata"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("Created directory: %s\n", dir)
	}

	fmt.Printf("\nConfiguration initialized successfully!\n")
	fmt.Printf("Config file: %s\n", configPath)
	fmt.Printf("\nYou can now use cli-recover with the default settings.\n")
	fmt.Printf("To customize settings, edit the config file or run 'cli-recover init --interactive'\n")

	return nil
}

// showConfiguration displays the current configuration
func showConfiguration(configPath string) error {
	loader := config.NewLoader(configPath)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Display configuration as YAML
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)

	fmt.Printf("Configuration file: %s\n", configPath)
	fmt.Println("---")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to display configuration: %w", err)
	}

	return nil
}

// resetConfiguration resets the configuration to defaults
func resetConfiguration(configPath string) error {
	writer := config.NewWriter(configPath)

	// Confirm reset
	if writer.Exists() {
		fmt.Printf("This will reset your configuration at: %s\n", configPath)
		fmt.Print("Are you sure? (y/N): ")

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Reset cancelled")
			return nil
		}
	}

	// Create default configuration
	if err := writer.CreateDefaultConfig(); err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}

	fmt.Println("Configuration reset to defaults successfully!")
	return nil
}

// interactiveSetup provides an interactive configuration setup
func interactiveSetup(configPath string) error {
	fmt.Println("Interactive CLI-Recover Setup")
	fmt.Println("=============================")
	fmt.Println()

	// Load existing config or use defaults
	loader := config.NewLoader(configPath)
	cfg, _ := loader.Load()
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Logger settings
	fmt.Println("Logger Configuration:")
	cfg.Logger.Level = promptWithDefault("  Log level (debug/info/warn/error/fatal)", cfg.Logger.Level)
	cfg.Logger.Output = promptWithDefault("  Output (console/file/both)", cfg.Logger.Output)

	if cfg.Logger.Output == "file" || cfg.Logger.Output == "both" {
		cfg.Logger.File.Path = promptWithDefault("  Log file path", cfg.Logger.File.Path)
		cfg.Logger.File.Format = promptWithDefault("  Log format (text/json)", cfg.Logger.File.Format)
	}

	// Backup settings
	fmt.Println("\nBackup Configuration:")
	cfg.Backup.DefaultCompression = promptWithDefault("  Default compression (gzip/bzip2/xz/none)", cfg.Backup.DefaultCompression)
	excludeVCS := promptWithDefault("  Exclude VCS directories (true/false)", fmt.Sprintf("%v", cfg.Backup.ExcludeVCS))
	cfg.Backup.ExcludeVCS = excludeVCS == "true"

	// Metadata settings
	fmt.Println("\nMetadata Configuration:")
	cfg.Metadata.Path = promptWithDefault("  Metadata path", cfg.Metadata.Path)
	cfg.Metadata.Format = promptWithDefault("  Metadata format (json/yaml)", cfg.Metadata.Format)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save configuration
	writer := config.NewWriter(configPath)
	if err := writer.Write(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\nConfiguration saved to: %s\n", configPath)

	// Create necessary directories
	dirs := []string{
		filepath.Dir(cfg.Logger.File.Path),
		cfg.Metadata.Path,
	}

	for _, dir := range dirs {
		expandedDir := expandPath(dir)
		if err := os.MkdirAll(expandedDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create directory %s: %v\n", expandedDir, err)
		}
	}

	fmt.Println("\nSetup completed successfully!")
	return nil
}

// promptWithDefault prompts for input with a default value
func promptWithDefault(prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)

	var input string
	fmt.Scanln(&input)

	if input == "" {
		return defaultValue
	}
	return input
}

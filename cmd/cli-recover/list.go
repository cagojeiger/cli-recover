package main

import (
	"github.com/cagojeiger/cli-recover/internal/domain/flags"
	"github.com/spf13/cobra"
)

// newListCommand creates the list command
func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backups and other resources",
		Long:  `List various resources like backups, restore jobs, etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newListBackupsCommand())
	// TODO: Add more subcommands like 'list jobs', 'list restores', etc.

	return cmd
}

// newListBackupsCommand creates the 'list backups' command
func newListBackupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backups",
		Aliases: []string{"backup"},
		Short:   "List all backups",
		Long:    `List all backups stored in the metadata repository`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeList(cmd, args)
		},
	}

	// Add flags using registry
	cmd.Flags().StringP(flags.LongNames.Namespace, flags.Registry.Namespace, "", "Filter by namespace")
	cmd.Flags().StringP(flags.LongNames.Format, flags.Registry.Format, "table", "Output format (table, json, yaml)")
	cmd.Flags().Bool(flags.LongNames.Details, false, "Show detailed information")

	return cmd
}

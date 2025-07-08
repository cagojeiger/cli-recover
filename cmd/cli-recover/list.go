package main

import (
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
	
	// Add flags
	cmd.Flags().StringP("namespace", "n", "", "Filter by namespace")
	cmd.Flags().StringP("output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().Bool("details", false, "Show detailed information")
	
	return cmd
}
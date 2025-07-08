package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cagojeiger/cli-recover/cmd/cli-recover/tui"
)

func newTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch Terminal User Interface mode",
		Long:  `Launch an interactive Terminal User Interface for cli-recover that provides a menu-driven interface for backup and restore operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}

			app := tui.NewApp(execPath)
			return app.Run()
		},
	}

	return cmd
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/cagojeiger/cli-recover/internal/domain/log"
	"github.com/cagojeiger/cli-recover/internal/infrastructure/log/storage"
)

// newLogsCommand creates the logs command
func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Manage operation logs",
		Long:  `View and manage logs from backup and restore operations`,
	}

	// Add subcommands
	cmd.AddCommand(newLogsListCommand())
	cmd.AddCommand(newLogsShowCommand())
	cmd.AddCommand(newLogsTailCommand())
	cmd.AddCommand(newLogsCleanCommand())

	return cmd
}

// newLogsListCommand creates the logs list command
func newLogsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all logs",
		Long:  `List all operation logs with their status and metadata`,
		RunE:  runLogsList,
	}

	cmd.Flags().String("type", "", "Filter by type (backup, restore)")
	cmd.Flags().String("provider", "", "Filter by provider (filesystem, minio, mongodb)")
	cmd.Flags().String("status", "", "Filter by status (running, completed, failed)")
	cmd.Flags().Int("limit", 20, "Maximum number of logs to display")
	cmd.Flags().String("log-dir", "", "Directory containing logs (default: ~/.cli-recover/logs)")

	return cmd
}

// newLogsShowCommand creates the logs show command
func newLogsShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [log-id]",
		Short: "Show log content",
		Long:  `Display the full content of a specific log file`,
		Args:  cobra.ExactArgs(1),
		RunE:  runLogsShow,
	}

	cmd.Flags().String("log-dir", "", "Directory containing logs (default: ~/.cli-recover/logs)")

	return cmd
}

// newLogsTailCommand creates the logs tail command
func newLogsTailCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail [log-id]",
		Short: "Tail log file",
		Long:  `Display the last lines of a log file (similar to tail -f)`,
		Args:  cobra.ExactArgs(1),
		RunE:  runLogsTail,
	}

	cmd.Flags().IntP("lines", "n", 50, "Number of lines to display")
	cmd.Flags().BoolP("follow", "f", false, "Follow log output (not yet implemented)")
	cmd.Flags().String("log-dir", "", "Directory containing logs (default: ~/.cli-recover/logs)")

	return cmd
}

// newLogsCleanCommand creates the logs clean command
func newLogsCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean old logs",
		Long:  `Remove logs older than specified days`,
		RunE:  runLogsClean,
	}

	cmd.Flags().Int("days", 30, "Remove logs older than this many days")
	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")
	cmd.Flags().String("log-dir", "", "Directory containing logs (default: ~/.cli-recover/logs)")

	return cmd
}

// runLogsList lists all logs
func runLogsList(cmd *cobra.Command, args []string) error {
	logDir := getLogDir(cmd)

	// Initialize repository
	repo, err := storage.NewFileRepository(logDir)
	if err != nil {
		return fmt.Errorf("failed to access log repository: %w", err)
	}

	// Build filter
	filter := log.ListFilter{}

	if typeFilter, _ := cmd.Flags().GetString("type"); typeFilter != "" {
		filter.Type = log.Type(typeFilter)
	}
	if provider, _ := cmd.Flags().GetString("provider"); provider != "" {
		filter.Provider = provider
	}
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		filter.Status = log.Status(status)
	}
	filter.Limit, _ = cmd.Flags().GetInt("limit")

	// List logs
	logs, err := repo.List(filter)
	if err != nil {
		return fmt.Errorf("failed to list logs: %w", err)
	}

	if len(logs) == 0 {
		fmt.Println("No logs found")
		return nil
	}

	// Print logs in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tPROVIDER\tSTATUS\tSTART TIME\tDURATION")
	fmt.Fprintln(w, "----\t----\t--------\t------\t----------\t--------")

	for _, l := range logs {
		duration := l.Duration().Round(time.Second).String()
		if l.Status == log.StatusRunning {
			duration = "running"
		}

		// Shorten ID for display
		shortID := l.ID
		if len(shortID) > 15 {
			shortID = shortID[:15] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			shortID,
			l.Type,
			l.Provider,
			l.Status,
			l.StartTime.Format("2006-01-02 15:04:05"),
			duration,
		)
	}
	w.Flush()

	return nil
}

// runLogsShow displays log content
func runLogsShow(cmd *cobra.Command, args []string) error {
	logID := args[0]
	logDir := getLogDir(cmd)

	// Initialize repository and service
	repo, err := storage.NewFileRepository(logDir)
	if err != nil {
		return fmt.Errorf("failed to access log repository: %w", err)
	}

	service := log.NewService(repo, filepath.Join(logDir, "files"))

	// Try to find log with partial ID
	fullLogID, err := findLogByPartialID(repo, logID)
	if err != nil {
		return err
	}

	// Read log content
	content, err := service.ReadLogFile(fullLogID)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Display content
	fmt.Print(string(content))

	return nil
}

// runLogsTail shows the tail of a log file
func runLogsTail(cmd *cobra.Command, args []string) error {
	logID := args[0]
	logDir := getLogDir(cmd)
	numLines, _ := cmd.Flags().GetInt("lines")
	follow, _ := cmd.Flags().GetBool("follow")

	if follow {
		return fmt.Errorf("follow mode not yet implemented")
	}

	// Initialize repository and service
	repo, err := storage.NewFileRepository(logDir)
	if err != nil {
		return fmt.Errorf("failed to access log repository: %w", err)
	}

	service := log.NewService(repo, filepath.Join(logDir, "files"))

	// Try to find log with partial ID
	fullLogID, err := findLogByPartialID(repo, logID)
	if err != nil {
		return err
	}

	// Read log content
	content, err := service.ReadLogFile(fullLogID)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Get last N lines
	lines := strings.Split(string(content), "\n")
	start := len(lines) - numLines
	if start < 0 {
		start = 0
	}

	for i := start; i < len(lines); i++ {
		if lines[i] != "" || i < len(lines)-1 {
			fmt.Println(lines[i])
		}
	}

	return nil
}

// runLogsClean removes old logs
func runLogsClean(cmd *cobra.Command, args []string) error {
	logDir := getLogDir(cmd)
	days, _ := cmd.Flags().GetInt("days")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Initialize repository
	repo, err := storage.NewFileRepository(logDir)
	if err != nil {
		return fmt.Errorf("failed to access log repository: %w", err)
	}

	// Calculate cutoff time
	maxAge := time.Duration(days) * 24 * time.Hour
	cutoff := time.Now().Add(-maxAge)

	// List old logs
	filter := log.ListFilter{}
	logs, err := repo.List(filter)
	if err != nil {
		return fmt.Errorf("failed to list logs: %w", err)
	}

	// Count and optionally delete old logs
	var oldLogs []*log.Log
	for _, l := range logs {
		if l.StartTime.Before(cutoff) {
			oldLogs = append(oldLogs, l)
		}
	}

	if len(oldLogs) == 0 {
		fmt.Printf("No logs older than %d days found\n", days)
		return nil
	}

	if dryRun {
		fmt.Printf("Would delete %d logs older than %d days:\n", len(oldLogs), days)
		for _, l := range oldLogs {
			fmt.Printf("  - %s (%s %s from %s)\n",
				l.ID, l.Type, l.Provider, l.StartTime.Format("2006-01-02"))
		}
	} else {
		fmt.Printf("Deleting %d logs older than %d days...\n", len(oldLogs), days)

		// Delete old logs
		deleted := 0
		for _, l := range oldLogs {
			// Remove log file if it exists
			if l.FilePath != "" {
				os.Remove(l.FilePath)
			}

			// Remove metadata
			if err := repo.Delete(l.ID); err == nil {
				deleted++
			}
		}

		fmt.Printf("Deleted %d logs\n", deleted)
	}

	return nil
}

// getLogDir returns the log directory from flags or default
func getLogDir(cmd *cobra.Command) string {
	logDir, _ := cmd.Flags().GetString("log-dir")
	if logDir == "" {
		homeDir, _ := os.UserHomeDir()
		logDir = filepath.Join(homeDir, ".cli-recover", "logs")
	}
	return logDir
}

// findLogByPartialID finds a log by partial ID match
func findLogByPartialID(repo log.Repository, partialID string) (string, error) {
	// First try exact match
	if _, err := repo.Get(partialID); err == nil {
		return partialID, nil
	}

	// Try partial match
	logs, err := repo.List(log.ListFilter{})
	if err != nil {
		return "", fmt.Errorf("failed to list logs: %w", err)
	}

	var matches []*log.Log
	for _, l := range logs {
		if strings.HasPrefix(l.ID, partialID) {
			matches = append(matches, l)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no log found with ID: %s", partialID)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple logs found with ID prefix: %s", partialID)
	}

	return matches[0].ID, nil
}

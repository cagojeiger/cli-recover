package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
	"github.com/cagojeiger/cli-recover/internal/runner"
)

// newMongoDBBackupCmd creates the MongoDB backup command
func newMongoDBBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mongodb [pod] [database]",
		Short: "Backup MongoDB database",
		Long:  `Backup MongoDB databases or collections`,
		Args:  cobra.ExactArgs(2),
		RunE:  runMongoDBBackup,
	}
	
	// Add MongoDB-specific flags
	cmd.Flags().StringP("namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringP("host", "", "localhost:27017", "MongoDB host:port")
	cmd.Flags().StringP("username", "u", "", "MongoDB username")
	cmd.Flags().StringP("password", "p", "", "MongoDB password")
	cmd.Flags().StringP("auth-db", "", "admin", "Authentication database")
	cmd.Flags().StringP("output", "o", "", "Output file path")
	cmd.Flags().BoolP("dry-run", "", false, "Show what would be executed without running")
	cmd.Flags().StringP("container", "", "", "Container name (for multi-container pods)")
	cmd.Flags().BoolP("gzip", "z", true, "Compress with gzip")
	cmd.Flags().StringSlice("collection", []string{}, "Specific collections to backup (can be repeated)")
	cmd.Flags().BoolP("oplog", "", false, "Include oplog for point-in-time restore")
	
	return cmd
}

func runMongoDBBackup(cmd *cobra.Command, args []string) error {
	pod := args[0]
	database := args[1]
	
	// Get all flags
	namespace, _ := cmd.Flags().GetString("namespace")
	host, _ := cmd.Flags().GetString("host")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	authDB, _ := cmd.Flags().GetString("auth-db")
	outputFile, _ := cmd.Flags().GetString("output")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	debug, _ := cmd.Flags().GetBool("debug")
	container, _ := cmd.Flags().GetString("container")
	gzip, _ := cmd.Flags().GetBool("gzip")
	collections, _ := cmd.Flags().GetStringSlice("collection")
	oplog, _ := cmd.Flags().GetBool("oplog")
	
	if debug {
		fmt.Printf("Debug: MongoDB backup\n")
		fmt.Printf("  pod: %s\n", pod)
		fmt.Printf("  database: %s\n", database)
		fmt.Printf("  namespace: %s\n", namespace)
		fmt.Printf("  host: %s\n", host)
		fmt.Printf("  username: %s\n", username)
		fmt.Printf("  auth-db: %s\n", authDB)
		fmt.Printf("  output: %s\n", outputFile)
		fmt.Printf("  container: %s\n", container)
		fmt.Printf("  gzip: %v\n", gzip)
		fmt.Printf("  collections: %v\n", collections)
		fmt.Printf("  oplog: %v\n", oplog)
		fmt.Printf("  dry-run: %v\n", dryRun)
	}
	
	runner := runner.NewRunner()
	
	// Verify pod exists
	if debug {
		fmt.Printf("Debug: Verifying pod exists in namespace %s\n", namespace)
	}
	pods, err := kubernetes.GetPods(runner, namespace)
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}
	
	found := false
	for _, p := range pods {
		if p.Name == pod {
			found = true
			if debug {
				fmt.Printf("Debug: Found pod %s (status: %s, ready: %s)\n", p.Name, p.Status, p.Ready)
			}
			break
		}
	}
	
	if !found {
		return fmt.Errorf("pod %s not found in namespace %s", pod, namespace)
	}
	
	// Generate output filename if not provided
	if outputFile == "" {
		if gzip {
			outputFile = fmt.Sprintf("mongodb-backup-%s-%s-%s.gz", namespace, pod, database)
		} else {
			outputFile = fmt.Sprintf("mongodb-backup-%s-%s-%s.bson", namespace, pod, database)
		}
	}
	
	// Build mongodump command
	command := buildMongoBackupCommand(pod, namespace, database, host, username, password, authDB, container, gzip, collections, oplog)
	
	if dryRun {
		fmt.Printf("Dry run - would execute: %s\n", command)
		fmt.Printf("Output would be saved to: %s\n", outputFile)
		return nil
	}
	
	fmt.Printf("Starting MongoDB backup...\n")
	fmt.Printf("Executing: %s\n", command)
	
	// Execute backup
	return executeMongoBackup(runner, command, outputFile, pod, namespace, database, debug)
}

// buildMongoBackupCommand creates the mongodump command
func buildMongoBackupCommand(pod, namespace, database, host, username, password, authDB, container string, gzip bool, collections []string, oplog bool) string {
	// Build kubectl exec prefix
	kubectlPrefix := fmt.Sprintf("kubectl exec -n %s %s", namespace, pod)
	if container != "" {
		kubectlPrefix += fmt.Sprintf(" -c %s", container)
	}
	kubectlPrefix += " --"
	
	// Build mongodump command
	mongodumpCmd := fmt.Sprintf("%s mongodump", kubectlPrefix)
	
	// Add host
	mongodumpCmd += fmt.Sprintf(" --host %s", host)
	
	// Add database
	mongodumpCmd += fmt.Sprintf(" --db %s", database)
	
	// Add authentication if provided
	if username != "" && password != "" {
		mongodumpCmd += fmt.Sprintf(" --username %s --password %s --authenticationDatabase %s", username, password, authDB)
	}
	
	// Add collections if specified
	for _, collection := range collections {
		mongodumpCmd += fmt.Sprintf(" --collection %s", collection)
	}
	
	// Add oplog if requested
	if oplog {
		mongodumpCmd += " --oplog"
	}
	
	// Add gzip if requested
	if gzip {
		mongodumpCmd += " --gzip"
	}
	
	// Output to stdout for streaming
	mongodumpCmd += " --archive"
	
	return mongodumpCmd
}

// executeMongoBackup performs the MongoDB backup operation
func executeMongoBackup(runner runner.Runner, command, outputFile, pod, namespace, database string, debug bool) error {
	if debug {
		fmt.Printf("Debug: Starting MongoDB backup execution\n")
	}
	
	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputFile, err)
	}
	defer outFile.Close()
	
	if debug {
		fmt.Printf("Debug: Created output file, executing mongodump command\n")
	}
	
	// Execute mongodump command and stream to file
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	
	// Execute the command and get output
	fmt.Printf("Dumping database '%s'...\n", database)
	output, err := runner.Run(parts[0], parts[1:]...)
	if err != nil {
		return fmt.Errorf("mongodump command failed: %w", err)
	}
	
	// Write output to file
	_, err = outFile.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}
	
	// Get file info for success message
	fileInfo, err := outFile.Stat()
	if err != nil {
		fmt.Printf("MongoDB backup completed successfully: %s\n", outputFile)
	} else {
		fmt.Printf("MongoDB backup completed successfully: %s (%d bytes)\n", outputFile, fileInfo.Size())
	}
	
	if debug {
		fmt.Printf("Debug: MongoDB backup execution completed\n")
	}
	
	return nil
}
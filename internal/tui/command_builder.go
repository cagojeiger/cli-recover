package tui

import (
	"strings"
	
	"github.com/cagojeiger/cli-recover/internal/kubernetes"
)

// CommandBuilder builds cli-recover commands incrementally
type CommandBuilder struct {
	action       string
	backupType   string
	pod          string
	path         string
	namespace    string
	options      kubernetes.BackupOptions
	minioOptions kubernetes.MinioBackupOptions
	mongoOptions kubernetes.MongoBackupOptions
}

// NewCommandBuilder creates a new command builder
func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{
		namespace: "default",
		options: kubernetes.BackupOptions{
			CompressionType: "gzip",
		},
	}
}

// Reset clears all builder state
func (cb *CommandBuilder) Reset() {
	cb.action = ""
	cb.backupType = ""
	cb.pod = ""
	cb.path = ""
	cb.namespace = "default"
	cb.options = kubernetes.BackupOptions{
		CompressionType: "gzip",
	}
}

// SetAction sets the action (backup/restore)
func (cb *CommandBuilder) SetAction(action string) {
	cb.action = action
}

// SetPod sets the pod name
func (cb *CommandBuilder) SetPod(pod string) {
	cb.pod = pod
}

// SetPath sets the path
func (cb *CommandBuilder) SetPath(path string) {
	cb.path = path
}

// SetNamespace sets the namespace
func (cb *CommandBuilder) SetNamespace(namespace string) {
	cb.namespace = namespace
}

// SetBackupType sets the backup type (filesystem, minio, mongodb)
func (cb *CommandBuilder) SetBackupType(backupType string) {
	cb.backupType = backupType
}

// SetOptions sets the backup options
func (cb *CommandBuilder) SetOptions(options kubernetes.BackupOptions) {
	cb.options = options
}

// SetMinioOptions sets the MinIO backup options
func (cb *CommandBuilder) SetMinioOptions(options kubernetes.MinioBackupOptions) {
	cb.minioOptions = options
}

// SetMongoOptions sets the MongoDB backup options
func (cb *CommandBuilder) SetMongoOptions(options kubernetes.MongoBackupOptions) {
	cb.mongoOptions = options
}

// Build returns the command as a slice of arguments
func (cb *CommandBuilder) Build() []string {
	if cb.action == "" {
		return []string{}
	}
	
	args := []string{cb.action}
	
	// Add backup type as subcommand for backup action
	if cb.action == "backup" && cb.backupType != "" {
		args = append(args, cb.backupType)
	}
	
	if cb.pod != "" {
		args = append(args, cb.pod)
	}
	
	// For MinIO: bucket/path, for MongoDB: database, for filesystem: path
	if cb.path != "" {
		args = append(args, cb.path)
	}
	
	// Add namespace if not default
	if cb.namespace != "" && cb.namespace != "default" {
		args = append(args, "--namespace", cb.namespace)
	}
	
	// Add type-specific options as flags
	var flags []string
	switch cb.backupType {
	case "minio":
		flags = cb.minioOptionsToFlags()
	case "mongodb":
		flags = cb.mongoOptionsToFlags()
	default:
		flags = cb.optionsToFlags()
	}
	args = append(args, flags...)
	
	return args
}

// Preview returns the command as a string for display
func (cb *CommandBuilder) Preview() string {
	args := cb.Build()
	if len(args) == 0 {
		return "cli-recover"
	}
	return "cli-recover " + strings.Join(args, " ")
}

// optionsToFlags converts BackupOptions to CLI flags
func (cb *CommandBuilder) optionsToFlags() []string {
	var flags []string
	opts := cb.options
	
	// Compression (only if not default gzip)
	if opts.CompressionType != "" && opts.CompressionType != "gzip" {
		flags = append(flags, "--compression", opts.CompressionType)
	}
	
	// Exclude patterns
	for _, pattern := range opts.ExcludePatterns {
		flags = append(flags, "--exclude", pattern)
	}
	
	// Boolean flags
	if opts.ExcludeVCS {
		flags = append(flags, "--exclude-vcs")
	}
	
	if opts.Verbose {
		flags = append(flags, "--verbose")
	}
	
	if opts.ShowTotals {
		flags = append(flags, "--totals")
	}
	
	if opts.PreservePerms {
		flags = append(flags, "--preserve-perms")
	}
	
	// String options
	if opts.Container != "" {
		flags = append(flags, "--container", opts.Container)
	}
	
	if opts.OutputFile != "" {
		flags = append(flags, "--output", opts.OutputFile)
	}
	
	return flags
}

// minioOptionsToFlags converts MinioBackupOptions to CLI flags
func (cb *CommandBuilder) minioOptionsToFlags() []string {
	var flags []string
	opts := cb.minioOptions
	
	if opts.Endpoint != "" && opts.Endpoint != "http://localhost:9000" {
		flags = append(flags, "--endpoint", opts.Endpoint)
	}
	
	if opts.AccessKey != "" {
		flags = append(flags, "--access-key", opts.AccessKey)
	}
	
	if opts.SecretKey != "" {
		flags = append(flags, "--secret-key", opts.SecretKey)
	}
	
	if opts.Format != "" && opts.Format != "tar" {
		flags = append(flags, "--format", opts.Format)
	}
	
	if !opts.Recursive {
		flags = append(flags, "--recursive=false")
	}
	
	if opts.Container != "" {
		flags = append(flags, "--container", opts.Container)
	}
	
	if opts.OutputFile != "" {
		flags = append(flags, "--output", opts.OutputFile)
	}
	
	return flags
}

// mongoOptionsToFlags converts MongoBackupOptions to CLI flags
func (cb *CommandBuilder) mongoOptionsToFlags() []string {
	var flags []string
	opts := cb.mongoOptions
	
	if opts.Host != "" && opts.Host != "localhost:27017" {
		flags = append(flags, "--host", opts.Host)
	}
	
	if opts.Username != "" {
		flags = append(flags, "--username", opts.Username)
	}
	
	if opts.Password != "" {
		flags = append(flags, "--password", opts.Password)
	}
	
	if opts.AuthDB != "" && opts.AuthDB != "admin" {
		flags = append(flags, "--auth-db", opts.AuthDB)
	}
	
	for _, collection := range opts.Collections {
		flags = append(flags, "--collection", collection)
	}
	
	if opts.Oplog {
		flags = append(flags, "--oplog")
	}
	
	if !opts.Gzip {
		flags = append(flags, "--gzip=false")
	}
	
	if opts.Container != "" {
		flags = append(flags, "--container", opts.Container)
	}
	
	if opts.OutputFile != "" {
		flags = append(flags, "--output", opts.OutputFile)
	}
	
	return flags
}
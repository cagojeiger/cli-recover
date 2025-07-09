package flags

// Registry defines all flag shortcuts in a centralized location
// to prevent conflicts and ensure consistency across commands
var Registry = struct {
	// Global flags used across all commands
	Namespace string
	Verbose   string
	Debug     string
	LogLevel  string

	// Backup command flags
	Output        string
	Compression   string
	Exclude       string
	Totals        string
	PreservePerms string
	DryRun        string

	// Restore command flags
	Force      string
	TargetPath string
	Container  string
	SkipPaths  string

	// List command flags
	Format  string
	Details string
}{
	// Global flags
	Namespace: "n",
	Verbose:   "v",
	Debug:     "d",
	LogLevel:  "", // No shortcut

	// Backup command flags
	Output:        "o",
	Compression:   "c",
	Exclude:       "e",
	Totals:        "T", // Uppercase to avoid conflict with target-path
	PreservePerms: "p",
	DryRun:        "", // No shortcut for safety

	// Restore command flags
	Force:      "f", // Replaces old overwrite flag
	TargetPath: "t",
	Container:  "C", // Uppercase to avoid conflict with compression
	SkipPaths:  "s",

	// List command flags
	Format:  "", // No shortcut to avoid conflict with output
	Details: "", // No shortcut
}

// LongNames defines the long flag names
var LongNames = struct {
	// Global
	Namespace string
	Verbose   string
	Debug     string
	LogLevel  string
	LogFile   string

	// Backup
	Output        string
	Compression   string
	Exclude       string
	Totals        string
	PreservePerms string
	Container     string
	DryRun        string

	// Restore
	Force      string
	TargetPath string
	SkipPaths  string

	// List
	Format  string
	Details string
}{
	// Global
	Namespace: "namespace",
	Verbose:   "verbose",
	Debug:     "debug",
	LogLevel:  "log-level",
	LogFile:   "log-file",

	// Backup
	Output:        "output",
	Compression:   "compression",
	Exclude:       "exclude",
	Totals:        "totals",
	PreservePerms: "preserve-perms",
	Container:     "container",
	DryRun:        "dry-run",

	// Restore
	Force:      "force",
	TargetPath: "target-path",
	SkipPaths:  "skip-paths",

	// List
	Format:  "output",
	Details: "details",
}

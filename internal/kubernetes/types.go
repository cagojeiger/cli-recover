package kubernetes

// Pod represents basic pod information
type Pod struct {
	Name      string
	Namespace string
	Status    string
	Ready     string
}

// DirectoryEntry represents a file or directory in the pod
type DirectoryEntry struct {
	Name     string
	Type     string // "dir" or "file"
	Size     string
	Modified string
}

// BackupOptions represents backup configuration options
type BackupOptions struct {
	// Compression settings
	CompressionType string // "gzip", "bzip2", "xz", "none"
	
	// Exclude patterns
	ExcludePatterns []string // ["*.log", "tmp/*", ".git"]
	ExcludeVCS      bool     // Exclude version control systems
	
	// Output settings
	Verbose       bool // Show progress
	ShowTotals    bool // Show total bytes
	PreservePerms bool // Preserve permissions
	
	// Container settings
	Container string // For multi-container pods
	
	// File settings
	OutputFile string // Output filename (auto-generated if empty)
}
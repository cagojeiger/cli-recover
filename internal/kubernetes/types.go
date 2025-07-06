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

// MinioBackupOptions represents MinIO-specific backup options
type MinioBackupOptions struct {
	// Connection settings
	Endpoint   string // MinIO endpoint URL
	AccessKey  string // MinIO access key
	SecretKey  string // MinIO secret key
	
	// Backup settings
	Recursive bool   // Recursive backup
	Format    string // Output format (tar, zip)
	
	// Common settings
	Container  string // For multi-container pods
	OutputFile string // Output filename
}

// MongoBackupOptions represents MongoDB-specific backup options
type MongoBackupOptions struct {
	// Connection settings
	Host     string // MongoDB host:port
	Username string // MongoDB username
	Password string // MongoDB password
	AuthDB   string // Authentication database
	
	// Backup settings
	Collections []string // Specific collections to backup
	Oplog       bool     // Include oplog for point-in-time restore
	Gzip        bool     // Compress with gzip
	
	// Common settings
	Container  string // For multi-container pods
	OutputFile string // Output filename
}
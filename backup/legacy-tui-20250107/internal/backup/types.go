package backup

import (
	"fmt"
	"strings"
)

// Type represents the type of backup
type Type string

const (
	// TypeFilesystem represents filesystem backup
	TypeFilesystem Type = "filesystem"
	// TypeMinio represents MinIO object storage backup
	TypeMinio Type = "minio"
	// TypeMongoDB represents MongoDB database backup
	TypeMongoDB Type = "mongodb"
)

// String returns the string representation of the backup type
func (t Type) String() string {
	return string(t)
}

// IsValid checks if the backup type is valid
func (t Type) IsValid() bool {
	switch t {
	case TypeFilesystem, TypeMinio, TypeMongoDB:
		return true
	default:
		return false
	}
}

// ParseType parses a string into a BackupType
func ParseType(s string) (Type, error) {
	t := Type(strings.ToLower(s))
	if !t.IsValid() {
		return "", fmt.Errorf("invalid backup type: %s", s)
	}
	return t, nil
}

// AllTypes returns all available backup types
func AllTypes() []Type {
	return []Type{TypeFilesystem, TypeMinio, TypeMongoDB}
}

// BackupSource represents the source for backup based on type
type BackupSource struct {
	Type   Type
	Pod    string
	Source string // path for filesystem, bucket/path for minio, db.collection for mongodb
}

// Validate validates the backup source based on its type
func (bs *BackupSource) Validate() error {
	if bs.Pod == "" {
		return fmt.Errorf("pod name is required")
	}
	
	switch bs.Type {
	case TypeFilesystem:
		if !strings.HasPrefix(bs.Source, "/") {
			return fmt.Errorf("filesystem path must be absolute (start with /)")
		}
	case TypeMinio:
		if !strings.Contains(bs.Source, "/") {
			return fmt.Errorf("minio source must be in format: bucket/path")
		}
	case TypeMongoDB:
		if !strings.Contains(bs.Source, ".") {
			return fmt.Errorf("mongodb source must be in format: database.collection")
		}
	default:
		return fmt.Errorf("invalid backup type: %s", bs.Type)
	}
	
	return nil
}

// InferType tries to infer the backup type from the source string
func InferType(source string) Type {
	// MongoDB pattern: database.collection (no slashes)
	if strings.Contains(source, ".") && !strings.Contains(source, "/") {
		return TypeMongoDB
	}
	
	// MinIO pattern: bucket/path (no leading slash)
	if !strings.HasPrefix(source, "/") && strings.Contains(source, "/") {
		return TypeMinio
	}
	
	// Default to filesystem (absolute path)
	return TypeFilesystem
}
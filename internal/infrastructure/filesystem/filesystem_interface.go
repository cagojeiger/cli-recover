package filesystem

import (
	"io"
	"os"
)

// FileSystem is an interface for file system operations
// This allows us to mock file operations in tests
type FileSystem interface {
	Create(name string) (File, error)
	Open(name string) (File, error)
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (os.FileInfo, error)
	Exists(name string) bool
}

// File is an interface that combines io.WriteCloser with Sync
type File interface {
	io.WriteCloser
	Sync() error
}

// OSFileSystem is the production implementation using real OS operations
type OSFileSystem struct{}

// Create creates a new file
func (fs *OSFileSystem) Create(name string) (File, error) {
	return os.Create(name)
}

// Open opens an existing file
func (fs *OSFileSystem) Open(name string) (File, error) {
	return os.Open(name)
}

// Remove removes a file
func (fs *OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

// Rename renames a file atomically
func (fs *OSFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Stat returns file info
func (fs *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Exists checks if a file exists
func (fs *OSFileSystem) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
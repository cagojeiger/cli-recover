package filesystem

import (
	"bytes"
	"errors"
	"os"
	"sync"
	"time"
)

// MockFileSystem is a mock implementation for testing
type MockFileSystem struct {
	mu         sync.Mutex
	files      map[string]*mockFile
	shouldFail map[string]error
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files:      make(map[string]*mockFile),
		shouldFail: make(map[string]error),
	}
}

// SetFailure sets a specific operation to fail
func (fs *MockFileSystem) SetFailure(filename string, err error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.shouldFail[filename] = err
}

// SetWriteFailureAfterBytes sets write to fail after n bytes
func (fs *MockFileSystem) SetWriteFailureAfterBytes(filename string, n int64) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	// Create a pre-configured file that will fail after n bytes
	fs.files[filename] = &mockFile{
		name:         filename,
		buffer:       &bytes.Buffer{},
		failAfterBytes: n,
		failError:    errors.New("write failed: simulated error"),
	}
}

// Create creates a mock file
func (fs *MockFileSystem) Create(name string) (File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if err, ok := fs.shouldFail[name]; ok {
		return nil, err
	}
	
	// If file was pre-configured (e.g., for failure testing), use it
	if f, exists := fs.files[name]; exists {
		return f, nil
	}
	
	f := &mockFile{
		name:   name,
		buffer: &bytes.Buffer{},
	}
	fs.files[name] = f
	return f, nil
}

// Open opens a mock file
func (fs *MockFileSystem) Open(name string) (File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if err, ok := fs.shouldFail[name]; ok {
		return nil, err
	}
	
	f, exists := fs.files[name]
	if !exists {
		return nil, os.ErrNotExist
	}
	
	// Create a new reader file
	return &mockFile{
		name:   name,
		buffer: bytes.NewBuffer(f.buffer.Bytes()),
		readonly: true,
	}, nil
}

// Remove removes a mock file
func (fs *MockFileSystem) Remove(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if err, ok := fs.shouldFail[name]; ok {
		return err
	}
	
	delete(fs.files, name)
	return nil
}

// Rename renames a file atomically
func (fs *MockFileSystem) Rename(oldpath, newpath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if err, ok := fs.shouldFail[oldpath]; ok {
		return err
	}
	
	f, exists := fs.files[oldpath]
	if !exists {
		return os.ErrNotExist
	}
	
	fs.files[newpath] = f
	f.name = newpath
	delete(fs.files, oldpath)
	return nil
}

// Stat returns mock file info
func (fs *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	f, exists := fs.files[name]
	if !exists {
		return nil, os.ErrNotExist
	}
	
	return &mockFileInfo{
		name: name,
		size: int64(f.buffer.Len()),
	}, nil
}

// Exists checks if a file exists in the mock
func (fs *MockFileSystem) Exists(name string) bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	_, exists := fs.files[name]
	return exists
}

// GetFileContent returns the content of a file (for testing)
func (fs *MockFileSystem) GetFileContent(name string) ([]byte, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	f, exists := fs.files[name]
	if !exists {
		return nil, os.ErrNotExist
	}
	
	return f.buffer.Bytes(), nil
}

// mockFile implements the File interface
type mockFile struct {
	name           string
	buffer         *bytes.Buffer
	closed         bool
	synced         bool
	readonly       bool
	bytesWritten   int64
	failAfterBytes int64
	failError      error
}

// Write writes to the mock file
func (f *mockFile) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, errors.New("file closed")
	}
	
	if f.readonly {
		return 0, errors.New("file is readonly")
	}
	
	// Check if we should fail after certain bytes
	if f.failAfterBytes > 0 && f.bytesWritten+int64(len(p)) > f.failAfterBytes {
		// Write partial data up to the failure point
		canWrite := int(f.failAfterBytes - f.bytesWritten)
		if canWrite > 0 {
			f.buffer.Write(p[:canWrite])
			f.bytesWritten += int64(canWrite)
		}
		return canWrite, f.failError
	}
	
	n, err = f.buffer.Write(p)
	f.bytesWritten += int64(n)
	return n, err
}

// Read reads from the mock file
func (f *mockFile) Read(p []byte) (n int, err error) {
	if f.closed {
		return 0, errors.New("file closed")
	}
	return f.buffer.Read(p)
}

// Close closes the mock file
func (f *mockFile) Close() error {
	f.closed = true
	return nil
}

// Sync syncs the mock file
func (f *mockFile) Sync() error {
	if f.closed {
		return errors.New("file closed")
	}
	f.synced = true
	return nil
}

// mockFileInfo implements os.FileInfo
type mockFileInfo struct {
	name string
	size int64
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return fi.size }
func (fi *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (fi *mockFileInfo) ModTime() time.Time { return time.Now() }
func (fi *mockFileInfo) IsDir() bool        { return false }
func (fi *mockFileInfo) Sys() interface{}   { return nil }
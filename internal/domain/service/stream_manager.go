package service

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

// StreamManager manages the lifecycle of streams in a pipeline
type StreamManager struct {
	streams map[string]*managedStream
	mu      sync.RWMutex
}

// managedStream wraps a pipe with metadata
type managedStream struct {
	reader *io.PipeReader
	writer *io.PipeWriter
	closed bool
}

// NewStreamManager creates a new StreamManager instance
func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*managedStream),
	}
}

// CreateStream creates a new named stream and returns a writer
func (sm *StreamManager) CreateStream(name string) (io.WriteCloser, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, exists := sm.streams[name]; exists {
		return nil, fmt.Errorf("stream '%s' already exists", name)
	}
	
	reader, writer := io.Pipe()
	sm.streams[name] = &managedStream{
		reader: reader,
		writer: writer,
		closed: false,
	}
	
	return writer, nil
}

// GetStream returns a reader for the named stream
func (sm *StreamManager) GetStream(name string) (io.Reader, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	stream, exists := sm.streams[name]
	if !exists {
		return nil, fmt.Errorf("stream '%s' not found", name)
	}
	
	return stream.reader, nil
}

// Connect connects an existing stream to a new stream
func (sm *StreamManager) Connect(source, destination string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sourceStream, exists := sm.streams[source]
	if !exists {
		return fmt.Errorf("source stream '%s' not found", source)
	}
	
	if _, exists := sm.streams[destination]; exists {
		return fmt.Errorf("destination stream '%s' already exists", destination)
	}
	
	// Create new pipe for destination
	reader, writer := io.Pipe()
	sm.streams[destination] = &managedStream{
		reader: reader,
		writer: writer,
		closed: false,
	}
	
	// Copy data from source to destination
	go func() {
		defer writer.Close()
		io.Copy(writer, sourceStream.reader)
	}()
	
	return nil
}

// ConnectWithTransform returns a reader and writer for transformation
func (sm *StreamManager) ConnectWithTransform(source, destination string) (io.Reader, io.WriteCloser, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sourceStream, exists := sm.streams[source]
	if !exists {
		return nil, nil, fmt.Errorf("source stream '%s' not found", source)
	}
	
	if _, exists := sm.streams[destination]; exists {
		return nil, nil, fmt.Errorf("destination stream '%s' already exists", destination)
	}
	
	// Create new pipe for destination
	reader, writer := io.Pipe()
	sm.streams[destination] = &managedStream{
		reader: reader,
		writer: writer,
		closed: false,
	}
	
	return sourceStream.reader, writer, nil
}

// CloseAll closes all open streams
func (sm *StreamManager) CloseAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	var errs []error
	
	for name, stream := range sm.streams {
		if !stream.closed {
			if stream.writer != nil {
				if err := stream.writer.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close writer for stream '%s': %w", name, err))
				}
			}
			stream.closed = true
		}
	}
	
	if len(errs) > 0 {
		return errors.New("failed to close some streams")
	}
	
	return nil
}

// HasStream checks if a stream exists
func (sm *StreamManager) HasStream(name string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	_, exists := sm.streams[name]
	return exists
}
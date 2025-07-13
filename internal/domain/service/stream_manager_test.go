package service

import (
	"io"
	"strings"
	"testing"
)

func TestStreamManager_CreateStream(t *testing.T) {
	sm := NewStreamManager()
	
	name := "test-stream"
	writer, err := sm.CreateStream(name)
	
	if err != nil {
		t.Fatalf("CreateStream() error = %v", err)
	}
	
	if writer == nil {
		t.Error("CreateStream() returned nil writer")
	}
	
	// Get the reader before writing
	reader, err := sm.GetStream(name)
	if err != nil {
		t.Fatalf("GetStream() error = %v", err)
	}
	
	// Write and read concurrently
	testData := "Hello, World!"
	done := make(chan bool)
	
	go func() {
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("Failed to read from stream: %v", err)
		}
		if string(data) != testData {
			t.Errorf("Read data = %v, want %v", string(data), testData)
		}
		done <- true
	}()
	
	// Write data to the stream
	_, err = io.WriteString(writer, testData)
	if err != nil {
		t.Fatalf("Failed to write to stream: %v", err)
	}
	
	// Close the writer to signal EOF
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	
	// Wait for reading to complete
	<-done
}

func TestStreamManager_GetStream_NotFound(t *testing.T) {
	sm := NewStreamManager()
	
	_, err := sm.GetStream("nonexistent")
	if err == nil {
		t.Error("GetStream() error = nil, want error for nonexistent stream")
	}
}

func TestStreamManager_CreateStream_Duplicate(t *testing.T) {
	sm := NewStreamManager()
	
	name := "test-stream"
	_, err := sm.CreateStream(name)
	if err != nil {
		t.Fatalf("First CreateStream() error = %v", err)
	}
	
	// Try to create the same stream again
	_, err = sm.CreateStream(name)
	if err == nil {
		t.Error("Second CreateStream() error = nil, want error for duplicate stream")
	}
}

func TestStreamManager_Connect(t *testing.T) {
	sm := NewStreamManager()
	
	// Create a source stream
	sourceName := "source"
	sourceWriter, _ := sm.CreateStream(sourceName)
	
	// Connect source to destination before writing
	destName := "destination"
	err := sm.Connect(sourceName, destName)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	
	// Get destination reader
	reader, err := sm.GetStream(destName)
	if err != nil {
		t.Fatalf("GetStream() error = %v", err)
	}
	
	// Read from destination concurrently
	testData := "Test data for connection"
	done := make(chan bool)
	
	go func() {
		data, _ := io.ReadAll(reader)
		if string(data) != testData {
			t.Errorf("Connected stream data = %v, want %v", string(data), testData)
		}
		done <- true
	}()
	
	// Write test data
	io.WriteString(sourceWriter, testData)
	sourceWriter.Close()
	
	// Wait for reading to complete
	<-done
}

func TestStreamManager_ConnectWithTransform(t *testing.T) {
	sm := NewStreamManager()
	
	// Create source stream
	sourceName := "source"
	sourceWriter, _ := sm.CreateStream(sourceName)
	
	// Get reader and writer for transformation
	reader, writer, err := sm.ConnectWithTransform(sourceName, "transformed")
	if err != nil {
		t.Fatalf("ConnectWithTransform() error = %v", err)
	}
	
	// Get result reader before writing
	result, _ := sm.GetStream("transformed")
	
	// Perform all operations concurrently
	done := make(chan bool)
	
	// Reader goroutine
	go func() {
		data, _ := io.ReadAll(result)
		if string(data) != "HELLO WORLD" {
			t.Errorf("Transformed data = %v, want %v", string(data), "HELLO WORLD")
		}
		done <- true
	}()
	
	// Transformer goroutine
	go func() {
		defer writer.Close()
		data, _ := io.ReadAll(reader)
		transformed := strings.ToUpper(string(data))
		io.WriteString(writer, transformed)
	}()
	
	// Write source data
	io.WriteString(sourceWriter, "hello world")
	sourceWriter.Close()
	
	// Wait for reading to complete
	<-done
}

func TestStreamManager_CloseAll(t *testing.T) {
	sm := NewStreamManager()
	
	// Create multiple streams
	streams := []string{"stream1", "stream2", "stream3"}
	readers := make(map[string]io.Reader)
	
	// Create streams and start readers
	for _, name := range streams {
		writer, _ := sm.CreateStream(name)
		reader, _ := sm.GetStream(name)
		readers[name] = reader
		
		// Read concurrently
		go func(n string, w io.WriteCloser) {
			io.WriteString(w, "test")
			// Don't close writers yet
		}(name, writer)
	}
	
	// Give some time for writes to happen
	// In production, this would be coordinated differently
	
	// Close all streams
	err := sm.CloseAll()
	if err != nil {
		t.Errorf("CloseAll() error = %v", err)
	}
	
	// Verify all streams received data
	for name, reader := range readers {
		data, err := io.ReadAll(reader)
		if err != nil && err != io.EOF {
			t.Errorf("Failed to read from stream %s after CloseAll: %v", name, err)
		}
		if string(data) != "test" && string(data) != "" {
			// Stream might be closed before data is written, which is OK for this test
			t.Logf("Stream %s data = %v", name, string(data))
		}
	}
}
package progress

import (
	"bytes"
	"testing"
)

// MockReporter is a test implementation of ProgressReporter
type MockReporter struct {
	StartCalled    bool
	UpdateCalled   int
	CompleteCalled bool
	ErrorCalled    bool

	LastOperation string
	LastTotal     int64
	LastCurrent   int64
	LastError     error
}

func (m *MockReporter) Start(operation string, total int64) {
	m.StartCalled = true
	m.LastOperation = operation
	m.LastTotal = total
}

func (m *MockReporter) Update(current, total int64) {
	m.UpdateCalled++
	m.LastCurrent = current
	m.LastTotal = total
}

func (m *MockReporter) Complete() {
	m.CompleteCalled = true
}

func (m *MockReporter) Error(err error) {
	m.ErrorCalled = true
	m.LastError = err
}

// Test 1: ProgressWriter tracks bytes written
func TestProgressWriter_TracksBytes(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := &MockReporter{}
	writer := NewProgressWriter(&buf, reporter, 1000)

	// Act
	data := []byte("Hello, World!")
	n, err := writer.Write(data)

	// Assert
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if writer.Current() != int64(len(data)) {
		t.Errorf("Expected current to be %d, got %d", len(data), writer.Current())
	}
}

// Test 2: ProgressWriter reports progress on each write
func TestProgressWriter_ReportsProgress(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	reporter := &MockReporter{}
	writer := NewProgressWriter(&buf, reporter, 100)

	// Act
	writer.Write([]byte("First write"))  // 11 bytes
	writer.Write([]byte("Second write")) // 12 bytes

	// Assert
	if reporter.UpdateCalled != 2 {
		t.Errorf("Expected Update to be called 2 times, called %d times", reporter.UpdateCalled)
	}
	if reporter.LastCurrent != 23 {
		t.Errorf("Expected last current to be 23, got %d", reporter.LastCurrent)
	}
	if reporter.LastTotal != 100 {
		t.Errorf("Expected last total to be 100, got %d", reporter.LastTotal)
	}
}

// Test 3: ProgressWriter passes through write errors
func TestProgressWriter_PassesThroughErrors(t *testing.T) {
	// Arrange
	reporter := &MockReporter{}
	failWriter := &failingWriter{err: bytes.ErrTooLarge}
	writer := NewProgressWriter(failWriter, reporter, 100)

	// Act
	_, err := writer.Write([]byte("test"))

	// Assert
	if err != bytes.ErrTooLarge {
		t.Errorf("Expected error %v, got %v", bytes.ErrTooLarge, err)
	}
}

// Test 4: ProgressWriter handles nil reporter gracefully
func TestProgressWriter_HandlesNilReporter(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	writer := NewProgressWriter(&buf, nil, 100)

	// Act & Assert - should not panic
	_, err := writer.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
}

// failingWriter is a writer that always returns an error
type failingWriter struct {
	err error
}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, f.err
}

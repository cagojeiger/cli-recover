package valueobject

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// StreamType represents the type of stream output
type StreamType int

const (
	StreamTypeStream StreamType = iota
	StreamTypeFile
	StreamTypeVariable
)

// String returns the string representation of StreamType
func (st StreamType) String() string {
	switch st {
	case StreamTypeStream:
		return "stream"
	case StreamTypeFile:
		return "file"
	case StreamTypeVariable:
		return "variable"
	default:
		return "unknown"
	}
}

// StreamName is a value object representing a valid stream name
type StreamName struct {
	value string
}

// NewStreamName creates a new StreamName with validation
func NewStreamName(name string) (StreamName, error) {
	if name == "" {
		return StreamName{}, errors.New("stream name cannot be empty")
	}
	
	// Allow alphanumeric, hyphens, underscores, and dots (for filenames)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !validName.MatchString(name) {
		return StreamName{}, errors.New("stream name contains invalid characters")
	}
	
	return StreamName{value: name}, nil
}

// String returns the string value of StreamName
func (sn StreamName) String() string {
	return sn.value
}

// StreamReference represents a reference to a stream with its type
type StreamReference struct {
	Type StreamType
	Name StreamName
}

// ParseStreamReference parses a stream reference string
func ParseStreamReference(ref string) (*StreamReference, error) {
	if ref == "" {
		return nil, errors.New("stream reference cannot be empty")
	}
	
	// Check for prefixes
	if strings.HasPrefix(ref, "file:") {
		name, err := NewStreamName(strings.TrimPrefix(ref, "file:"))
		if err != nil {
			return nil, fmt.Errorf("invalid file reference: %w", err)
		}
		return &StreamReference{
			Type: StreamTypeFile,
			Name: name,
		}, nil
	}
	
	if strings.HasPrefix(ref, "var:") {
		name, err := NewStreamName(strings.TrimPrefix(ref, "var:"))
		if err != nil {
			return nil, fmt.Errorf("invalid variable reference: %w", err)
		}
		return &StreamReference{
			Type: StreamTypeVariable,
			Name: name,
		}, nil
	}
	
	// Check for invalid prefix
	if strings.Contains(ref, ":") {
		return nil, errors.New("invalid stream reference prefix")
	}
	
	// Default to stream type
	name, err := NewStreamName(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid stream reference: %w", err)
	}
	
	return &StreamReference{
		Type: StreamTypeStream,
		Name: name,
	}, nil
}
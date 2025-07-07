package log

import (
	"time"
)

// Repository defines the interface for log storage
type Repository interface {
	// Save saves a log entry
	Save(log *Log) error
	
	// Get retrieves a log by ID
	Get(id string) (*Log, error)
	
	// List returns all logs with optional filters
	List(filter ListFilter) ([]*Log, error)
	
	// Update updates an existing log
	Update(log *Log) error
	
	// Delete removes a log entry (and optionally its file)
	Delete(id string) error
	
	// GetLatest returns the most recent log matching the filter
	GetLatest(filter ListFilter) (*Log, error)
}

// ListFilter provides filtering options for listing logs
type ListFilter struct {
	Type      Type
	Provider  string
	Status    Status
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
}
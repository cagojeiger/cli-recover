package pipeline

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile parses a YAML file and returns a Pipeline
func ParseFile(filepath string) (*Pipeline, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseBytes(data)
}

// ParseBytes parses YAML content from bytes
func ParseBytes(data []byte) (*Pipeline, error) {
	var pipeline Pipeline

	err := yaml.Unmarshal(data, &pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the pipeline
	if err := pipeline.Validate(); err != nil {
		return nil, fmt.Errorf("pipeline validation failed: %w", err)
	}

	return &pipeline, nil
}
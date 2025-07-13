package persistence

import (
	"errors"
	"fmt"
	"os"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"gopkg.in/yaml.v3"
)

// StepConfig represents a step in the YAML configuration
type StepConfig struct {
	Name   string `yaml:"name"`
	Run    string `yaml:"run"`
	Input  string `yaml:"input,omitempty"`
	Output string `yaml:"output,omitempty"`
}

// PipelineConfig represents the YAML configuration structure
type PipelineConfig struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description,omitempty"`
	Steps       []StepConfig `yaml:"steps,omitempty"`
}

// YAMLParser handles parsing of pipeline YAML files
type YAMLParser struct{}

// NewYAMLParser creates a new YAML parser instance
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// ParseFile parses a YAML file and returns a PipelineConfig
func (p *YAMLParser) ParseFile(filepath string) (*PipelineConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	return p.ParseBytes(data)
}

// ParseBytes parses YAML content from bytes
func (p *YAMLParser) ParseBytes(data []byte) (*PipelineConfig, error) {
	var config PipelineConfig
	
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Basic validation
	if config.Name == "" {
		return nil, errors.New("pipeline name is required")
	}
	
	return &config, nil
}

// ToPipeline converts a PipelineConfig to a domain Pipeline entity
func (p *YAMLParser) ToPipeline(config *PipelineConfig) (*entity.Pipeline, error) {
	if config.Name == "" {
		return nil, errors.New("pipeline name cannot be empty")
	}
	
	pipeline, err := entity.NewPipeline(config.Name, config.Description)
	if err != nil {
		return nil, err
	}
	
	// Convert steps
	for _, stepConfig := range config.Steps {
		if stepConfig.Name == "" {
			return nil, errors.New("step name cannot be empty")
		}
		if stepConfig.Run == "" {
			return nil, errors.New("step command cannot be empty")
		}
		
		step, err := entity.NewStep(stepConfig.Name, stepConfig.Run)
		if err != nil {
			return nil, fmt.Errorf("failed to create step '%s': %w", stepConfig.Name, err)
		}
		
		if stepConfig.Input != "" {
			step.SetInput(stepConfig.Input)
		}
		
		if stepConfig.Output != "" {
			step.SetOutput(stepConfig.Output)
		}
		
		pipeline.AddStep(step)
	}
	
	return pipeline, nil
}
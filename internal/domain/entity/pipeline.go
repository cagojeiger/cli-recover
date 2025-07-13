package entity

import (
	"errors"
	"fmt"
)

// Pipeline represents a series of steps to be executed
type Pipeline struct {
	Name        string
	Description string
	Steps       []*Step
}

// NewPipeline creates a new Pipeline instance
func NewPipeline(name, description string) (*Pipeline, error) {
	if name == "" {
		return nil, errors.New("pipeline name cannot be empty")
	}
	
	return &Pipeline{
		Name:        name,
		Description: description,
		Steps:       make([]*Step, 0),
	}, nil
}

// AddStep adds a step to the pipeline
func (p *Pipeline) AddStep(step *Step) {
	p.Steps = append(p.Steps, step)
}

// Validate validates the pipeline structure
func (p *Pipeline) Validate() error {
	if len(p.Steps) == 0 {
		return errors.New("pipeline must have at least one step")
	}
	
	// Check for duplicate step names
	stepNames := make(map[string]bool)
	for _, step := range p.Steps {
		if stepNames[step.Name] {
			return fmt.Errorf("duplicate step name: %s", step.Name)
		}
		stepNames[step.Name] = true
	}
	
	// Build output map
	outputs := make(map[string]bool)
	for _, step := range p.Steps {
		if step.Output != "" {
			outputs[step.Output] = true
		}
	}
	
	// Check for orphaned inputs
	for _, step := range p.Steps {
		if step.Input != "" && !outputs[step.Input] {
			return fmt.Errorf("step '%s' references undefined input '%s'", step.Name, step.Input)
		}
	}
	
	return nil
}
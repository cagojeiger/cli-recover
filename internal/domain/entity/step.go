package entity

import "errors"

// Step represents a single command in a pipeline
type Step struct {
	Name   string
	Run    string
	Input  string
	Output string
}

// NewStep creates a new Step instance
func NewStep(name, run string) (*Step, error) {
	if name == "" {
		return nil, errors.New("step name cannot be empty")
	}
	
	if run == "" {
		return nil, errors.New("step command cannot be empty")
	}
	
	return &Step{
		Name: name,
		Run:  run,
	}, nil
}

// SetInput sets the input stream name for this step
func (s *Step) SetInput(input string) {
	s.Input = input
}

// SetOutput sets the output stream name for this step
func (s *Step) SetOutput(output string) {
	s.Output = output
}

// Validate validates the step
func (s *Step) Validate() error {
	// Individual step validation is minimal
	// Most validation happens at the pipeline level
	return nil
}
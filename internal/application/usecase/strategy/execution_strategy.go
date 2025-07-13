package strategy

import (
	"github.com/cagojeiger/cli-recover/internal/domain/entity"
)

// ExecutionStrategy defines the interface for different pipeline execution strategies
type ExecutionStrategy interface {
	// Execute runs the pipeline using the specific strategy
	Execute(pipeline *entity.Pipeline) error
}

// DetermineStrategy decides which execution strategy to use based on pipeline characteristics
func DetermineStrategy(pipeline *entity.Pipeline) ExecutionStrategy {
	// Check if pipeline is simple linear and doesn't require progress
	if isSimpleLinear(pipeline) && !requiresProgress(pipeline) {
		return &ShellPipeStrategy{}
	}
	
	// For complex pipelines or those requiring progress, use Go streams
	return &GoStreamStrategy{}
}

// ShellPipeStrategy executes pipeline using Unix pipes
type ShellPipeStrategy struct{}

// Execute implements ExecutionStrategy for shell pipes
func (s *ShellPipeStrategy) Execute(pipeline *entity.Pipeline) error {
	// TODO: Implement shell pipe execution
	return nil
}

// GoStreamStrategy executes pipeline using Go io streams
type GoStreamStrategy struct{}

// Execute implements ExecutionStrategy for Go streams
func (g *GoStreamStrategy) Execute(pipeline *entity.Pipeline) error {
	// TODO: Implement Go stream execution
	return nil
}
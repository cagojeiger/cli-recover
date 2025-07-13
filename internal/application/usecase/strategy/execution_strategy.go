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
	// For now, return nil to make tests compile
	// We'll implement the actual logic after creating the strategies
	return nil
}
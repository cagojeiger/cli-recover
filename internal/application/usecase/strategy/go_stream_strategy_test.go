package strategy

import (
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestGoStreamStrategy_ImplementsInterface(t *testing.T) {
	// Test that GoStreamStrategy implements ExecutionStrategy
	var _ ExecutionStrategy = (*GoStreamStrategy)(nil)
}

func TestGoStreamStrategy_Execute(t *testing.T) {
	// For now, just test that we can create and call Execute
	strategy := &GoStreamStrategy{}
	
	pipeline := &entity.Pipeline{
		Name: "test",
		Steps: []*entity.Step{
			{Name: "echo", Run: "echo hello"},
		},
	}
	
	// This will return nil for now (stub implementation)
	err := strategy.Execute(pipeline)
	assert.NoError(t, err)
}
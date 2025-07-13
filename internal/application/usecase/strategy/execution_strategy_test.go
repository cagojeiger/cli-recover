package strategy

import (
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

// Test that ExecutionStrategy interface is properly defined
func TestExecutionStrategyInterface(t *testing.T) {
	// Test that interface is properly defined by checking nil interface
	var strategy ExecutionStrategy
	assert.Nil(t, strategy) // Uninitialized interface should be nil
	
	// Test that we can assign implementations
	strategy = &ShellPipeStrategy{}
	assert.NotNil(t, strategy)
	
	strategy = &GoStreamStrategy{}
	assert.NotNil(t, strategy)
}

// Mock implementation for testing
type mockStrategy struct {
	executeCalled bool
	executeError  error
}

func (m *mockStrategy) Execute(pipeline *entity.Pipeline) error {
	m.executeCalled = true
	return m.executeError
}

func TestMockStrategyImplementsInterface(t *testing.T) {
	// Verify mock implements the interface
	var _ ExecutionStrategy = (*mockStrategy)(nil)
	
	// Test mock behavior
	mock := &mockStrategy{}
	pipeline := &entity.Pipeline{
		Name: "test",
		Steps: []*entity.Step{
			{Name: "step1", Run: "echo test"},
		},
	}
	
	err := mock.Execute(pipeline)
	assert.NoError(t, err)
	assert.True(t, mock.executeCalled)
}

// Test strategy determination
func TestDetermineStrategy(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *entity.Pipeline
		wantType string
	}{
		{
			name: "simple linear pipeline should use shell pipe",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "echo hello", Output: "text"},
					{Name: "step2", Run: "tr upper", Input: "text"},
				},
			},
			wantType: "ShellPipeStrategy",
		},
		{
			name: "complex pipeline with branches should use go streams",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "echo hello", Output: "text"},
					{Name: "step2", Run: "tee", Input: "text", Output: "copy1"},
					{Name: "step3", Run: "tr upper", Input: "text", Output: "copy2"},
				},
			},
			wantType: "GoStreamStrategy",
		},
		{
			name: "pipeline requiring progress should use go streams",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "step1", Run: "tar cf - /data", Output: "archive"},
					{Name: "step2", Run: "gzip", Input: "archive"},
				},
			},
			wantType: "GoStreamStrategy", // When progress is needed
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := DetermineStrategy(tt.pipeline)
			assert.NotNil(t, strategy)
			
			// Check the actual type
			switch tt.wantType {
			case "ShellPipeStrategy":
				_, ok := strategy.(*ShellPipeStrategy)
				assert.True(t, ok, "Expected ShellPipeStrategy but got %T", strategy)
			case "GoStreamStrategy":
				_, ok := strategy.(*GoStreamStrategy)
				assert.True(t, ok, "Expected GoStreamStrategy but got %T", strategy)
			}
		})
	}
}
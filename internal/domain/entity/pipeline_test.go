package entity

import (
	"testing"
)

func TestPipeline_NewPipeline(t *testing.T) {
	tests := []struct {
		name        string
		pipelineName string
		description string
		wantErr     bool
	}{
		{
			name:        "valid pipeline",
			pipelineName: "test-pipeline",
			description: "Test pipeline description",
			wantErr:     false,
		},
		{
			name:        "empty name should fail",
			pipelineName: "",
			description: "Test description",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := NewPipeline(tt.pipelineName, tt.description)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPipeline() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("NewPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if pipeline.Name != tt.pipelineName {
				t.Errorf("Pipeline.Name = %v, want %v", pipeline.Name, tt.pipelineName)
			}
			
			if pipeline.Description != tt.description {
				t.Errorf("Pipeline.Description = %v, want %v", pipeline.Description, tt.description)
			}
		})
	}
}

func TestPipeline_AddStep(t *testing.T) {
	pipeline, _ := NewPipeline("test-pipeline", "Test description")
	
	step := &Step{
		Name:   "test-step",
		Run:    "echo hello",
		Output: "greeting",
	}
	
	pipeline.AddStep(step)
	
	if len(pipeline.Steps) != 1 {
		t.Errorf("Pipeline.Steps length = %v, want %v", len(pipeline.Steps), 1)
	}
	
	if pipeline.Steps[0] != step {
		t.Errorf("Pipeline.Steps[0] = %v, want %v", pipeline.Steps[0], step)
	}
}

func TestPipeline_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Pipeline
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pipeline with connected steps",
			setup: func() *Pipeline {
				p, _ := NewPipeline("test", "test")
				p.AddStep(&Step{Name: "step1", Run: "echo test", Output: "data"})
				p.AddStep(&Step{Name: "step2", Run: "cat", Input: "data"})
				return p
			},
			wantErr: false,
		},
		{
			name: "pipeline with no steps should fail",
			setup: func() *Pipeline {
				p, _ := NewPipeline("test", "test")
				return p
			},
			wantErr: true,
			errMsg:  "pipeline must have at least one step",
		},
		{
			name: "pipeline with orphaned input should fail",
			setup: func() *Pipeline {
				p, _ := NewPipeline("test", "test")
				p.AddStep(&Step{Name: "step1", Run: "cat", Input: "nonexistent"})
				return p
			},
			wantErr: true,
			errMsg:  "step 'step1' references undefined input 'nonexistent'",
		},
		{
			name: "pipeline with duplicate step names should fail",
			setup: func() *Pipeline {
				p, _ := NewPipeline("test", "test")
				p.AddStep(&Step{Name: "step1", Run: "echo test", Output: "data"})
				p.AddStep(&Step{Name: "step1", Run: "cat", Input: "data"})
				return p
			},
			wantErr: true,
			errMsg:  "duplicate step name: step1",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := tt.setup()
			err := pipeline.Validate()
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Pipeline.Validate() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Pipeline.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Pipeline.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
package entity

import (
	"testing"
)

func TestStep_NewStep(t *testing.T) {
	tests := []struct {
		name    string
		stepName string
		run     string
		wantErr bool
		errMsg  string
	}{
		{
			name:     "valid step",
			stepName: "test-step",
			run:      "echo hello",
			wantErr:  false,
		},
		{
			name:     "empty name should fail",
			stepName: "",
			run:      "echo hello",
			wantErr:  true,
			errMsg:   "step name cannot be empty",
		},
		{
			name:     "empty command should fail",
			stepName: "test-step",
			run:      "",
			wantErr:  true,
			errMsg:   "step command cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, err := NewStep(tt.stepName, tt.run)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewStep() error = nil, wantErr %v", tt.wantErr)
				} else if err.Error() != tt.errMsg {
					t.Errorf("NewStep() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if step.Name != tt.stepName {
				t.Errorf("Step.Name = %v, want %v", step.Name, tt.stepName)
			}

			if step.Run != tt.run {
				t.Errorf("Step.Run = %v, want %v", step.Run, tt.run)
			}
		})
	}
}

func TestStep_SetInput(t *testing.T) {
	step, _ := NewStep("test-step", "cat")
	
	input := "test-input"
	step.SetInput(input)
	
	if step.Input != input {
		t.Errorf("Step.Input = %v, want %v", step.Input, input)
	}
}

func TestStep_SetOutput(t *testing.T) {
	step, _ := NewStep("test-step", "echo hello")
	
	output := "test-output"
	step.SetOutput(output)
	
	if step.Output != output {
		t.Errorf("Step.Output = %v, want %v", step.Output, output)
	}
}

func TestStep_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Step
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid step with output",
			setup: func() *Step {
				s, _ := NewStep("step1", "echo hello")
				s.SetOutput("greeting")
				return s
			},
			wantErr: false,
		},
		{
			name: "valid step with input and output",
			setup: func() *Step {
				s, _ := NewStep("step1", "cat")
				s.SetInput("input-stream")
				s.SetOutput("output-stream")
				return s
			},
			wantErr: false,
		},
		{
			name: "first step should not have input",
			setup: func() *Step {
				s, _ := NewStep("step1", "echo hello")
				s.SetInput("should-not-exist")
				s.SetOutput("greeting")
				return s
			},
			wantErr: false, // This is OK - validation happens at pipeline level
		},
		{
			name: "last step can have no output",
			setup: func() *Step {
				s, _ := NewStep("step1", "cat > file.txt")
				s.SetInput("data")
				return s
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := tt.setup()
			err := step.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Step.Validate() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Step.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Step.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
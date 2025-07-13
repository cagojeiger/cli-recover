package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYAMLParser_ParseFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		validate func(t *testing.T, p *PipelineConfig)
	}{
		{
			name: "simple pipeline",
			content: `name: test-pipeline
description: Test pipeline
steps:
  - name: generate
    run: echo "Hello World"
    output: greeting
  - name: transform
    run: tr '[:lower:]' '[:upper:]'
    input: greeting
    output: result
`,
			wantErr: false,
			validate: func(t *testing.T, p *PipelineConfig) {
				if p.Name != "test-pipeline" {
					t.Errorf("Pipeline.Name = %v, want %v", p.Name, "test-pipeline")
				}
				if len(p.Steps) != 2 {
					t.Errorf("len(Pipeline.Steps) = %v, want %v", len(p.Steps), 2)
				}
				if p.Steps[0].Name != "generate" {
					t.Errorf("Steps[0].Name = %v, want %v", p.Steps[0].Name, "generate")
				}
				if p.Steps[1].Input != "greeting" {
					t.Errorf("Steps[1].Input = %v, want %v", p.Steps[1].Input, "greeting")
				}
			},
		},
		{
			name: "pipeline with no steps",
			content: `name: empty-pipeline
description: Empty pipeline
`,
			wantErr: false,
			validate: func(t *testing.T, p *PipelineConfig) {
				if len(p.Steps) != 0 {
					t.Errorf("len(Pipeline.Steps) = %v, want %v", len(p.Steps), 0)
				}
			},
		},
		{
			name: "invalid yaml",
			content: `name: broken
steps
  - this is not valid yaml
`,
			wantErr: true,
		},
		{
			name: "missing name",
			content: `description: No name pipeline
steps:
  - name: step1
    run: echo test
`,
			wantErr: true,
		},
	}

	parser := NewYAMLParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "pipeline.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Parse file
			config, err := parser.ParseFile(tmpFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFile() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestYAMLParser_ParseBytes(t *testing.T) {
	content := []byte(`name: byte-pipeline
steps:
  - name: step1
    run: echo test
    output: data
`)

	parser := NewYAMLParser()
	config, err := parser.ParseBytes(content)

	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if config.Name != "byte-pipeline" {
		t.Errorf("Pipeline.Name = %v, want %v", config.Name, "byte-pipeline")
	}
}

func TestYAMLParser_ToPipeline(t *testing.T) {
	tests := []struct {
		name    string
		config  *PipelineConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &PipelineConfig{
				Name: "test-pipeline",
				Description: "Test",
				Steps: []StepConfig{
					{Name: "step1", Run: "echo test", Output: "data"},
					{Name: "step2", Run: "cat", Input: "data"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: &PipelineConfig{
				Name: "",
				Steps: []StepConfig{
					{Name: "step1", Run: "echo test"},
				},
			},
			wantErr: true,
			errMsg: "pipeline name cannot be empty",
		},
		{
			name: "step with empty name",
			config: &PipelineConfig{
				Name: "test",
				Steps: []StepConfig{
					{Name: "", Run: "echo test"},
				},
			},
			wantErr: true,
			errMsg: "step name cannot be empty",
		},
		{
			name: "step with empty command",
			config: &PipelineConfig{
				Name: "test",
				Steps: []StepConfig{
					{Name: "step1", Run: ""},
				},
			},
			wantErr: true,
			errMsg: "step command cannot be empty",
		},
	}

	parser := NewYAMLParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := parser.ToPipeline(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ToPipeline() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ToPipeline() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ToPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if pipeline.Name != tt.config.Name {
				t.Errorf("Pipeline.Name = %v, want %v", pipeline.Name, tt.config.Name)
			}

			if len(pipeline.Steps) != len(tt.config.Steps) {
				t.Errorf("len(Pipeline.Steps) = %v, want %v", len(pipeline.Steps), len(tt.config.Steps))
			}
		})
	}
}
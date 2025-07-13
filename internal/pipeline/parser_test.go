package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    *Pipeline
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pipeline",
			yaml: `name: test-pipeline
description: A test pipeline
steps:
  - name: step1
    run: echo hello
    output: data
  - name: step2
    run: cat
    input: data`,
			want: &Pipeline{
				Name:        "test-pipeline",
				Description: "A test pipeline",
				Steps: []Step{
					{Name: "step1", Run: "echo hello", Output: "data"},
					{Name: "step2", Run: "cat", Input: "data"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid yaml",
			yaml: `name: test
steps
  - invalid`,
			wantErr: true,
			errMsg:  "failed to parse YAML",
		},
		{
			name: "missing name",
			yaml: `steps:
  - name: step1
    run: echo hello`,
			wantErr: true,
			errMsg:  "pipeline name cannot be empty",
		},
		{
			name: "empty pipeline",
			yaml: `name: test
steps: []`,
			wantErr: true,
			errMsg:  "pipeline must have at least one step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBytes([]byte(tt.yaml))
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a valid pipeline file
	validYAML := `name: test-pipeline
steps:
  - name: echo
    run: echo hello`
	validFile := filepath.Join(tmpDir, "valid.yaml")
	require.NoError(t, os.WriteFile(validFile, []byte(validYAML), 0644))

	// Create an invalid pipeline file
	invalidYAML := `invalid yaml content`
	invalidFile := filepath.Join(tmpDir, "invalid.yaml")
	require.NoError(t, os.WriteFile(invalidFile, []byte(invalidYAML), 0644))

	tests := []struct {
		name    string
		file    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid file",
			file:    validFile,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file:    filepath.Join(tmpDir, "missing.yaml"),
			wantErr: true,
			errMsg:  "failed to read file",
		},
		{
			name:    "invalid yaml file",
			file:    invalidFile,
			wantErr: true,
			errMsg:  "failed to parse YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile(tt.file)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "test-pipeline", got.Name)
			}
		})
	}
}
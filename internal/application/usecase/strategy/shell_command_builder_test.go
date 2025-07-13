package strategy

import (
	"testing"

	"github.com/cagojeiger/cli-recover/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestBuildShellCommand(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *entity.Pipeline
		want     string
		wantErr  bool
	}{
		{
			name: "simple two-step pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "echo", Run: "echo hello", Output: "text"},
					{Name: "upper", Run: "tr '[:lower:]' '[:upper:]'", Input: "text"},
				},
			},
			want: `echo hello | tr '[:lower:]' '[:upper:]'`,
		},
		{
			name: "three-step pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "tar", Run: "tar cf - /data", Output: "archive"},
					{Name: "gzip", Run: "gzip -9", Input: "archive", Output: "compressed"},
					{Name: "save", Run: "cat > backup.tar.gz", Input: "compressed"},
				},
			},
			want: `tar cf - /data | gzip -9 | cat > backup.tar.gz`,
		},
		{
			name: "single step pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "list", Run: "ls -la"},
				},
			},
			want: `ls -la`,
		},
		{
			name: "empty pipeline",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{},
			},
			wantErr: true,
		},
		{
			name: "non-linear pipeline should error",
			pipeline: &entity.Pipeline{
				Steps: []*entity.Step{
					{Name: "source", Run: "echo data", Output: "data"},
					{Name: "branch1", Run: "grep foo", Input: "data"},
					{Name: "branch2", Run: "grep bar", Input: "data"},
				},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildShellCommand(tt.pipeline)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuildShellCommandWithLogging(t *testing.T) {
	tests := []struct {
		name     string
		pipeline *entity.Pipeline
		logDir   string
		want     string
	}{
		{
			name: "pipeline with logging",
			pipeline: &entity.Pipeline{
				Name: "test-pipeline",
				Steps: []*entity.Step{
					{Name: "echo", Run: "echo hello", Output: "text"},
					{Name: "upper", Run: "tr upper", Input: "text"},
				},
			},
			logDir: "/tmp/logs",
			want: `#!/bin/bash
set -o pipefail
LOGDIR="/tmp/logs"
mkdir -p "$LOGDIR"

# Pipeline: test-pipeline
(echo hello 2>"$LOGDIR/echo.err" | tee "$LOGDIR/echo.out") | \
(tr upper 2>"$LOGDIR/upper.err" | tee "$LOGDIR/upper.out")

# Save exit code
EXIT_CODE=$?
echo "Pipeline exit code: $EXIT_CODE" > "$LOGDIR/pipeline.status"
exit $EXIT_CODE`,
		},
		{
			name: "single step with logging",
			pipeline: &entity.Pipeline{
				Name: "single",
				Steps: []*entity.Step{
					{Name: "list", Run: "ls -la"},
				},
			},
			logDir: "/tmp/logs",
			want: `#!/bin/bash
set -o pipefail
LOGDIR="/tmp/logs"
mkdir -p "$LOGDIR"

# Pipeline: single
ls -la 2>"$LOGDIR/list.err" | tee "$LOGDIR/list.out"

# Save exit code
EXIT_CODE=$?
echo "Pipeline exit code: $EXIT_CODE" > "$LOGDIR/pipeline.status"
exit $EXIT_CODE`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildShellCommandWithLogging(tt.pipeline, tt.logDir)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
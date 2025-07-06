package runner

import (
	"errors"
	"os"
	"os/exec"
)

// ErrTest is a test error for mocking
var ErrTest = errors.New("test error")

// Runner interface for command execution
type Runner interface {
	Run(cmd string, args ...string) ([]byte, error)
}

// ShellRunner executes real commands
type ShellRunner struct{}

// Run executes actual shell command
func (r *ShellRunner) Run(cmd string, args ...string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

// NewRunner creates appropriate runner based on environment
func NewRunner() Runner {
	if os.Getenv("USE_GOLDEN") == "true" {
		return &GoldenRunner{dir: "../../testdata"}
	}
	return &ShellRunner{}
}
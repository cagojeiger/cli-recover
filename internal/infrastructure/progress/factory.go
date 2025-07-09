package progress

import (
	"io"
	"os"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/progress"
	"golang.org/x/term"
)

// NewAutoReporter creates the appropriate progress reporter based on the environment
func NewAutoReporter(w io.Writer) progress.ProgressReporter {
	// Check if we're in a CI environment
	if isCI() {
		return NewCIReporter(w, 10*time.Second)
	}

	// Check if output is a terminal
	if f, ok := w.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		return NewTerminalReporter(w)
	}

	// Default to pipe reporter for non-terminal output
	return NewPipeReporter(w)
}

// isCI detects if we're running in a CI/CD environment
func isCI() bool {
	// Check common CI environment variables
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"JENKINS_HOME",
		"GITLAB_CI",
		"GITHUB_ACTIONS",
		"CIRCLECI",
		"TRAVIS",
		"BUILDKITE",
		"DRONE",
	}

	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}

	return false
}

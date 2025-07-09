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
	return NewAutoReporterWithDelay(w, true)
}

// NewAutoReporterWithDelay creates the appropriate progress reporter with optional 3-second delay
func NewAutoReporterWithDelay(w io.Writer, useDelay bool) progress.ProgressReporter {
	var reporter progress.ProgressReporter

	// Check if we're in a CI environment
	if isCI() {
		reporter = NewCIReporter(w, 10*time.Second)
	} else if f, ok := w.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		// Check if output is a terminal
		reporter = NewTerminalReporter(w)
	} else {
		// Default to pipe reporter for non-terminal output
		reporter = NewPipeReporter(w)
	}

	// Apply 3-second delay rule if requested
	if useDelay {
		reporter = NewDelayedReporter(reporter)
	}

	return reporter
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

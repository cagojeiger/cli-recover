package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoldenRunner reads from test files  
type GoldenRunner struct {
	dir string
}

// Run reads golden file instead of executing command
func (r *GoldenRunner) Run(cmd string, args ...string) ([]byte, error) {
	filename := sanitizeFilename(cmd, args)
	path := filepath.Join(r.dir, cmd, filename)
	
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no golden file for: %s %s", cmd, strings.Join(args, " "))
	}
	
	return data, nil
}

// NewGoldenRunner creates a new golden file runner
func NewGoldenRunner(dir string) *GoldenRunner {
	return &GoldenRunner{dir: dir}
}

// sanitizeFilename converts command and args to filename
func sanitizeFilename(cmd string, args []string) string {
	// Remove flags that start with -
	cleanArgs := make([]string, 0)
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			// Skip the flag, but keep its value
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				cleanArgs = append(cleanArgs, args[i][1:]) // Remove leading dash
				cleanArgs = append(cleanArgs, args[i+1])
				i++ // Skip next arg as we already processed it
			} else {
				cleanArgs = append(cleanArgs, args[i][1:]) // Remove leading dash
			}
		} else {
			cleanArgs = append(cleanArgs, args[i])
		}
	}
	
	filename := strings.Join(cleanArgs, "-")
	filename = strings.ReplaceAll(filename, "/", "")
	return filename + ".golden"
}
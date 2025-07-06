# Project Context

## Goals and Constraints

### Primary Goals
- Kubernetes pod filesystem backup tool with interactive TUI
- Simplified kubectl commands for backup operations
- Directory browsing and selective backup functionality
- Educational command comparison (kubectl vs cli-restore)

### Core Features
- Interactive TUI for namespace/pod/directory selection
- Comprehensive backup options (compression, exclusions, advanced settings)
- Command generation with tar options
- Golden file testing for reliable CI/CD

### Technical Constraints
- CLAUDE.md compliance: files <500 lines, functions <50 lines
- Test coverage >90%
- Go standard project layout with internal packages
- Bubble Tea framework for TUI
- No external Kubernetes dependencies for testing

### User Experience Goals
- Intuitive navigation with keyboard shortcuts
- Visual directory browsing with icons
- Side-by-side command comparison
- Tab-based option configuration

### Success Criteria
- ✅ Interactive directory selection instead of hardcoded paths
- ✅ Backup options configuration before command generation
- ✅ Clean separation of concerns (kubernetes/runner/tui packages)
- ✅ Comprehensive test coverage with golden files
- ✅ All CLAUDE.md rules compliance
# System Architecture

## Package Structure

### CLI Entry Point
- `cmd/cli-restore/main.go`: Cobra CLI with TUI mode and backup subcommand
- Single binary deployment with embedded TUI

### Internal Packages

#### `internal/kubernetes/`
- **Purpose**: Kubernetes API interactions and backup logic
- **Files**:
  - `types.go`: Data structures (Pod, DirectoryEntry, BackupOptions)
  - `client.go`: K8s API calls (GetNamespaces, GetPods, GetDirectoryContents)
  - `backup.go`: Backup command generation with tar options
- **Dependencies**: internal/runner for command execution

#### `internal/runner/`
- **Purpose**: Command execution abstraction for testing
- **Files**:
  - `runner.go`: Interface and ShellRunner for production
  - `golden.go`: GoldenRunner for tests with mock data
- **Pattern**: Strategy pattern for test/production environments

#### `internal/tui/`
- **Purpose**: Terminal User Interface components
- **Files**:
  - `model.go`: Bubble Tea model and core state
  - `view.go`: Main view renderer with version display
  - `screens.go`: Basic screens (main, namespace, pod, directory)
  - `handlers.go`: Keyboard input handling and navigation
  - `options.go`: Backup options UI with tab navigation
- **Dependencies**: internal/kubernetes, internal/runner

## Data Flow

### TUI Navigation Flow
```
Main Menu → Namespace Selection → Pod Selection → 
Directory Browsing → Backup Options → Command Comparison
```

### Command Execution Flow
```
User Input → Model Update → Screen Render → 
K8s API Call → Golden/Shell Runner → Display Results
```

## Testing Strategy

### Golden File Testing
- Mock kubectl responses in `testdata/kubectl/`
- Filename pattern: `{command}-{args}.golden`
- Environment variable `USE_GOLDEN=true` for test mode

### TUI Testing
- Bubble Tea teatest framework for interaction simulation
- Terminal size simulation for responsive design
- Complete user journey testing
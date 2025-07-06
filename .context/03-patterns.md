# Coding Conventions

## Code Organization

### File Naming
- **Descriptive names**: `backup.go`, `client.go`, `handlers.go`
- **Test files**: `*_test.go` suffix
- **No underscores**: Use camelCase for multi-word names

### Package Structure
- **Single responsibility**: Each package has one clear purpose
- **Minimal dependencies**: Packages depend only on what they need
- **Internal first**: Use internal packages for non-exported logic

## Go Conventions

### Naming
- **Exported types**: PascalCase (Pod, BackupOptions)
- **Unexported types**: camelCase (goldenRunner, shellRunner)
- **Functions**: PascalCase for exported, camelCase for internal
- **Constants**: PascalCase or ALL_CAPS for package-level

### Error Handling
- **Wrap errors**: Use `fmt.Errorf("context: %w", err)`
- **Return early**: Check errors immediately and return
- **Meaningful messages**: Include context about what failed

### Testing
- **Table-driven tests**: Use struct slices for multiple test cases
- **Descriptive test names**: Test{Function}_{Scenario}
- **Setup/teardown**: Use environment variables for test configuration

## TUI Patterns

### Model-View-Update
- **Model**: Contains all application state
- **Update**: Handles messages and updates model
- **View**: Renders current state to string

### Screen Management
- **Enum screens**: Use iota for screen constants
- **State transitions**: Clear navigation between screens
- **Back navigation**: Consistent 'b' key for going back

### Key Handling
- **Consistent bindings**: Same keys for similar actions
- **Vi-style navigation**: j/k for up/down movement
- **Enter for selection**: Standard selection key

## Testing Patterns

### Golden File Testing
- **Filename convention**: `{command}-{args}.golden`
- **Environment switching**: `USE_GOLDEN` env var
- **Sanitized filenames**: Remove special characters safely

### TUI Testing
- **Fixed terminal size**: Consistent test environment
- **Wait patterns**: Use teatest.WaitFor for async operations
- **Complete journeys**: Test full user workflows

## Documentation

### Comments
- **Package comments**: Describe package purpose
- **Exported functions**: Document parameters and return values
- **Complex logic**: Explain non-obvious implementations

### README Style
- **Korean for users**: End-user documentation in Korean
- **English for developers**: Technical docs in English
- **Code examples**: Include usage examples
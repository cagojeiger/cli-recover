# Long-term Patterns

## Development Patterns
- TDD (Test-Driven Development) cycle strictly followed
- Occam's Razor applied to all design decisions
- Small, focused functions (< 50 lines)
- High test coverage (> 90%) maintained

## Code Patterns
- Unified monitoring for all pipeline execution
- Builder pattern for command construction
- Interface-based design for extensibility
- Concurrent I/O handling with goroutines

## Architecture Patterns
- Simplified clean architecture
- Configuration-based approach (no flags)
- Automatic monitoring (no opt-in/opt-out)
- Linear pipeline focus (non-linear deferred)
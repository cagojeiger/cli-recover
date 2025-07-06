# Key Learnings

## Development Insights

### Refactoring Large Files
**Learning**: Breaking 900+ line files requires careful dependency analysis
**Challenge**: Circular import issues when splitting packages
**Solution**: Clear dependency hierarchy (tui → kubernetes → runner)
**Takeaway**: Plan package boundaries before splitting

### Go Module Path Issues
**Learning**: Internal imports must use full module path
**Mistake**: Using relative paths in internal packages  
**Fix**: Updated all imports to `github.com/cagojeiger/cli-restore/internal/*`
**Lesson**: Module path consistency is critical

### Test Data Management
**Learning**: Golden files need careful path management
**Issue**: testdata moved from cmd/cli-restore/ to project root
**Impact**: All relative paths in tests needed updating
**Solution**: Use `../../testdata` from test files
**Takeaway**: Consider test data location early in project structure

### TUI Testing Complexity
**Learning**: Testing interactive UIs requires patience and precision
**Challenge**: Timing issues with async screen updates
**Solution**: teatest.WaitFor with appropriate timeouts
**Best Practice**: Test complete user journeys, not just individual screens

## Technical Insights

### CLAUDE.md Rule Application
**Learning**: 500-line file limit drives better architecture
**Benefit**: Forced separation of concerns improved code quality
**Challenge**: Some complex functions still approach 50-line limit
**Action**: Need further refactoring of large functions

### Test Coverage Metrics
**Learning**: Go coverage reports can be misleading
**Issue**: 0% coverage reported due to main() function only being counted
**Reality**: Internal packages have good test coverage
**Solution**: Run coverage on specific packages, not just main

### Golden File Naming
**Learning**: Filename sanitization is tricky
**Pattern**: Replace spaces, slashes, and special characters
**Implementation**: `sanitizeFilename()` function handles edge cases
**Maintenance**: Keep golden files synchronized with actual kubectl output

## User Experience Insights

### Progressive Disclosure
**Learning**: Too many options at once overwhelm users
**Solution**: Tab-based category navigation in backup options
**Result**: Cleaner, more focused interface

### Command Comparison Value
**Learning**: Showing both kubectl and cli-restore commands is educational
**Benefit**: Users understand what the tool does behind the scenes
**Implementation**: Side-by-side display with wrapping for long commands

### Directory Browsing UX
**Learning**: Visual cues (📁/📄 icons) significantly improve navigation
**Implementation**: Unicode icons for directory vs file distinction
**Result**: More intuitive file system exploration

## Process Learnings

### Planning Before Implementation
**Learning**: CLAUDE.md Rule 02 prevents scope creep
**Practice**: Always create plan before coding
**Benefit**: Clear requirements understanding reduces rework

### Incremental Testing
**Learning**: Test each major change immediately
**Practice**: Run tests after each package split
**Benefit**: Easier to identify and fix issues early
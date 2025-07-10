# ⚠️ EXPERIMENTAL - Provider Isolation Implementation

## WARNING: This is experimental code!
This directory contains experimental code for provider isolation architecture.
DO NOT use in production. This entire directory may be deleted at any time.

## Status
- 🧪 **Phase**: Experimenting
- 📅 **Started**: 2025-01-10
- 🎯 **Goal**: Validate provider isolation approach

## Philosophy
> "Duplication is cheaper than the wrong abstraction"  
> — Sandi Metz

> "Isolation with minimal coordination"  
> — Our approach

## What's Different Here?

### Current (Production)
```
internal/
├── domain/
│   └── backup/provider.go      # Common interface
└── infrastructure/
    └── filesystem/             # Implements common interface
```

### Experimental (This Directory)
```
experimental/providers/
└── filesystem_v2/
    ├── backup.go               # Standalone implementation
    ├── restore.go              # No shared interfaces
    └── tui.go                  # Provider-specific TUI
```

## Key Principles
1. **Complete Isolation**: No dependencies on other providers
2. **Intentional Duplication**: Copy code rather than share
3. **Provider-Specific**: Optimized for filesystem only
4. **Self-Contained**: Everything needed in one place

## How to Test
```bash
# Enable experimental provider
export USE_EXPERIMENTAL=true

# Run backup with experimental code
./cli-recover backup filesystem <pod> <path>

# Disable experimental provider
unset USE_EXPERIMENTAL
```

## Success Criteria
- [ ] Reduced complexity compared to current implementation
- [ ] Complete provider isolation achieved
- [ ] Tests run independently
- [ ] No impact on production code

## Failure Criteria
- Increased complexity
- Difficult to maintain
- Performance degradation
- Team confusion

## Next Steps
1. If successful: Gradual migration plan
2. If failed: Delete this directory
3. If partial: Extract good ideas only

---

**Remember**: This is an experiment. It's okay to fail. The goal is learning.
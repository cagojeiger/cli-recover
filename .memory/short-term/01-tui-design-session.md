# TUI Design Session Summary

## Session Date: 2025-07-06

## Key Design Decisions

### 1. Command Pattern
- Established pattern: `cli-restore [action] [target] [options]`
- Actions: backup, restore, verify, schedule, history
- Targets: pod, configmap, secret, pvc, mongodb, postgres, mysql, redis, elastic, minio, s3

### 2. Layout Standardization
```
┌─ Header (2 lines) ─────┐
│ Title & Status         │
│ Navigation Breadcrumb  │
├─ Main Content ─────────┤
│ List-based Interface   │
├─ Command Preview ──────┤
│ $ Generated command    │
├─ Footer ──────────────┤
│ Contextual shortcuts   │
└────────────────────────┘
```

### 3. UI Principles
- No box/card UI - pure list navigation
- Consistent j/k vim-style movement
- Real-time command preview
- Execute within TUI

### 4. Pod File Browser Design
- Directory tree navigation
- Pattern-based selection
- Smart path suggestions
- Real-time size calculation
- Autocomplete support

### 5. Extensibility
- Easy to add new backup types
- Each target has its own view module
- Shared layout system
- Independent view architecture
# CLAUDE.md

## Metadata
- PURPOSE: Define collaboration patterns between AI assistant and user
- VERSION: 2.0
- SCOPE: All projects in this repository
- LANGUAGE: English (optimized for AI parsing)

## Core Principles

### PRINCIPLE: Working Agreement
This document defines project-agnostic collaboration patterns that persist across all our work.

## Rules (Priority Order)

### RULE_00: Context Engineering Management
```
PRIORITY: P0 (Critical - Always Execute)
TRIGGER: 
  - Project initialization
  - Task direction change
  - User request for status
ACTION:
  - MAINTAIN directories: [.context/, .memory/, .planning/, .checkpoint/]
  - CREATE files with pattern: 00-*.md
  - ENFORCE limits: 500 lines per file
  - USE format: Bullet points only
PURPOSE: Synchronize mental models between user and AI
EVALUATION: Files exist and are current
```

### RULE_01: Occam's Razor Enforcement
```
PRIORITY: P0 (Critical)
TRIGGER: 
  - New feature request
  - Solution proposal
  - Architecture decision
ACTION:
  - EVALUATE complexity: 0-100 scale
  - REJECT if complexity > 70
  - PROPOSE simpler alternative
COMPLEXITY_SCALE:
  - [0-30]: Simple, clear solution ✅
  - [31-50]: Acceptable complexity ⚠️
  - [51-70]: Complex but allowable ⚠️⚠️
  - [71-100]: Over-engineered, must simplify ❌
EVALUATION_CRITERIA:
  - Lines of code
  - Number of dependencies
  - Abstraction layers
  - Maintenance burden
  - Readability score
PURPOSE: Prevent unnecessary complexity
```

### RULE_02: Planning Before Implementation
```
PRIORITY: P0 (Critical)
TRIGGER: User requests implementation
ACTION:
  - PAUSE before coding
  - CREATE detailed plan
  - VERIFY understanding with user
  - ONLY THEN proceed to implementation
PURPOSE: Prevent misunderstood requirements
EVALUATION: Plan exists before first line of code
```

### RULE_03: Documentation Standards
```
PRIORITY: P1 (High)
TRIGGER: Creating user-facing documentation
ACTION:
  - WRITE in Korean for all user docs
  - FOLLOW project's existing style
PURPOSE: Maintain consistency for end users
```

### RULE_04: Code Quality Metrics
```
PRIORITY: P1 (High)
TRIGGER: Code submission
REQUIREMENTS:
  - Test coverage > 90%
  - File size < 500 lines
  - Functions < 50 lines
ACTION_IF_VIOLATED:
  - REFACTOR large files
  - ADD missing tests
  - SPLIT complex functions
PURPOSE: Maintain high code quality
```

### RULE_05: Commit Convention
```
PRIORITY: P2 (Medium)
TRIGGER: Creating git commit
ACTION:
  - USE format: type(scope): description
  - TYPES: [feat, fix, docs, style, refactor, test, chore]
  - ADD emoji suffix: 🤖 Generated with Claude Code
  - CO-AUTHOR: Claude <noreply@anthropic.com>
PURPOSE: Clear commit history
```

### RULE_06: CI/CD Verification
```
PRIORITY: P1 (High)
TRIGGER: After push with open PR
ACTION:
  - RUN: gh pr checks <PR_NUMBER>
  - IF failed: gh run view <RUN_ID> --log-failed
  - FIX all failures before marking complete
PURPOSE: Ensure code quality in pipeline
```

## Directory Structure

### .context/
```
PURPOSE: Project runtime context
CONTAINS:
  - 00-project.md: Goals and constraints
  - 01-architecture.md: System design
  - 02-tech-stack.md: Technology decisions
  - 03-patterns.md: Coding conventions
```

### .memory/
```
PURPOSE: AI memory system
STRUCTURE:
  - short-term/: Current session state
    - 00-current-task.md
    - 01-working-context.md
  - long-term/: Persistent knowledge
    - 00-decisions.md
    - 01-learnings.md
    - 02-patterns.md
```

### .planning/
```
PURPOSE: Task and execution planning
CONTAINS:
  - 00-roadmap.md: Overall project plan
  - 01-current-sprint.md: Active work
  - 02-backlog.md: Future tasks
```

### .checkpoint/
```
PURPOSE: Project state snapshots
PATTERN: 00-descriptor.md (e.g., 01-mvp.md)
USE_CASE: Major milestones, rollback points
```

## Override Conditions

User can override any rule by:
1. Explicit instruction
2. Modifying this file
3. Project-specific requirements

## Evaluation

AI should self-evaluate adherence to these rules:
- Before starting tasks
- After completing tasks
- When uncertainty arises
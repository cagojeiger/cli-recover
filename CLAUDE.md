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
  - MAINTAIN directories: [.meta/context/, .meta/memory/, .meta/planning/, .meta/checkpoint/]
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

### RULE_02: Planning Before Implementation (TDD-Enhanced)
```
PRIORITY: P0 (Critical)
TRIGGER: User requests implementation
ACTION:
  - PAUSE before coding
  - CREATE detailed plan
  - IDENTIFY first failing test to write
  - VERIFY understanding with user
  - FOLLOW TDD cycle: Red → Green → Refactor
  - ONLY THEN proceed to implementation
PURPOSE: Prevent misunderstood requirements and ensure test-driven development
EVALUATION: 
  - Plan exists before first line of code
  - First commit is a failing test
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

### RULE_05: Commit Convention (TDD-Aware)
```
PRIORITY: P2 (Medium)
TRIGGER: Creating git commit
ACTION:
  - USE format: type(scope): description
  - TYPES: [feat, fix, docs, style, refactor, test, chore]
  - SPECIAL for TDD:
    - "test: " for adding tests (Red phase)
    - "feat: " for implementation (Green phase)
    - "refactor: " for structure improvements (Refactor phase)
  - ADD emoji suffix: 🤖 Generated with Claude Code
  - CO-AUTHOR: Claude <noreply@anthropic.com>
PURPOSE: Clear commit history showing TDD progression
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

### RULE_07: Architecture Analysis with Tree
```
PRIORITY: P1 (High)
TRIGGER:
  - Architecture compliance verification
  - Code structure analysis request
  - Hexagonal architecture validation
  - After major refactoring
  - Project structure documentation
ACTION:
  - USE: tree command with appropriate depth
  - APPLY ignore patterns: [node_modules, __pycache__, .git, vendor, dist, build]
  - DEPTH: 3-4 levels for overview, unlimited for detailed analysis
  - FOCUS on: internal/, cmd/, pkg/ directories for Go projects
  - VERIFY: Dependency directions (Domain → Application → Infrastructure)
COMMAND_EXAMPLES:
  - Overview: tree -L 3 -I 'node_modules|__pycache__|.git'
  - Hexagonal check: tree internal/ -L 4
  - Full analysis: tree -a -I '.git' > project-structure.txt
PURPOSE: Visual understanding of project structure and architecture compliance
EVALUATION: 
  - Clear separation of concerns visible
  - No circular dependencies
  - Hexagonal architecture layers properly organized
HEXAGONAL_ARCHITECTURE_CHECK:
  - Domain layer: No external dependencies
  - Application layer: Only depends on Domain
  - Infrastructure layer: Implements Domain interfaces
  - Adapters properly separated from core business logic
```

### RULE_08: TDD & Tidy First Development
```
PRIORITY: P0 (Critical)
TRIGGER:
  - Writing new functionality
  - Modifying existing behavior
  - Code refactoring needed
ACTION:
  - RED: Write failing test first
  - GREEN: Implement minimum code to pass
  - REFACTOR: Improve structure only after tests pass
  - SEPARATE changes:
    - Structural: Renaming, extracting, moving (no behavior change)
    - Behavioral: Adding/modifying functionality
  - COMMIT separately: Never mix structural and behavioral changes
TDD_CYCLE:
  1. Write one failing test
  2. Run test to see it fail
  3. Write minimal code to pass
  4. Run all tests
  5. Refactor if needed
  6. Repeat
COMMIT_RULES:
  - Only when ALL tests pass
  - One logical change per commit
  - Message format: "test: add test for X" or "feat: implement X" or "refactor: extract Y"
PURPOSE: Maintain high code quality through disciplined development
EVALUATION:
  - Test coverage increases with each feature
  - No mixed commits in history
  - All tests passing before merge
```

## Directory Structure

### .meta/context/
```
PURPOSE: Project runtime context
CONTAINS (minimum):
  - 00-project.md: Goals and constraints
  - 01-architecture.md: System design
  - 02-tech-stack.md: Technology decisions
  - 03-patterns.md: Coding conventions
NOTE: Additional files may be added as needed following 00-*.md pattern
```

### .meta/memory/
```
PURPOSE: AI memory system
STRUCTURE:
  - short-term/: Current session state (minimum)
    - 00-current-task.md
    - 01-working-context.md
    - Additional files as needed (00-*.md pattern)
  - long-term/: Persistent knowledge (minimum)
    - 00-decisions.md
    - 01-learnings.md
    - 02-patterns.md
    - Additional files as needed (00-*.md pattern)
```

### .meta/planning/
```
PURPOSE: Task and execution planning
CONTAINS (minimum):
  - 00-roadmap.md: Overall project plan
  - 01-current-sprint.md: Active work
  - 02-backlog.md: Future tasks
NOTE: Additional files may be added as needed following 00-*.md pattern
```

### .meta/checkpoint/
```
PURPOSE: Project state snapshots
PATTERN: 00-descriptor.md (e.g., 01-mvp.md)
USE_CASE: Major milestones, rollback points
NOTE: Multiple checkpoints expected as project progresses
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
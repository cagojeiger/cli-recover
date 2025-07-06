# Next Task: Backup Command Implementation

## Objective
- Implement `cli-restore backup <pod> <path>` command
- Kubernetes Pod file extraction with tar

## Prerequisites
- Wait for v0.1.0 PR merge
- Create new feature branch

## Planned Structure Expansion
From:
```
cmd/cli-restore/main.go  # Everything here
```

To:
```
internal/
├── commands/
│   ├── root.go
│   ├── version.go
│   └── backup.go
├── k8s/
│   └── client.go
└── archive/
    └── tar.go
```

## Key Dependencies to Add
- k8s.io/client-go
- k8s.io/cli-runtime

## Reference
- See `.planning/01-future-features.md` for detailed design
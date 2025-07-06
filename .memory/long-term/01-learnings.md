# Learnings

## 2025-07-06: GitHub Actions for Go Projects
- **Learning**: Actions syntax for Go builds
- **Key Points**:
  * Use `actions/setup-go@v5` for Go setup
  * Extract version from tag: `VERSION=${GITHUB_REF#refs/tags/}`
  * Must set `permissions: contents: write` for releases
  * Upload each binary separately as release asset

## 2025-07-06: Cross-Platform Build Process
- **Learning**: GOOS/GOARCH combinations
- **Platforms Tested**:
  * `GOOS=darwin GOARCH=amd64` - Intel Mac
  * `GOOS=darwin GOARCH=arm64` - Apple Silicon
  * `GOOS=linux GOARCH=amd64` - Linux x64
  * `GOOS=linux GOARCH=arm64` - Linux ARM
- **Build Flags**: `-ldflags "-s -w"` reduces binary size

## 2025-07-06: PR Workflow with gh CLI
- **Learning**: Creating PR from command line
- **Commands**:
  ```bash
  git checkout -b feat/branch-name
  git push -u origin feat/branch-name
  gh pr create --title "..." --body "..."
  ```
- **Best Practice**: Include test plan in PR body
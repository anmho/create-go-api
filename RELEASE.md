# Release Guide

This document describes how to release new versions of `create-go-api`.

## Prerequisites

1. Install [goreleaser](https://goreleaser.com/install/):
   ```bash
   brew install goreleaser
   # or
   go install github.com/goreleaser/goreleaser@latest
   ```

2. Set up GitHub token (for releases):
   ```bash
   export GITHUB_TOKEN=your_token_here
   ```

3. Ensure you have push access to:
   - `github.com/anmho/create-go-api` (main repository)

## Release Process

### 1. Update Version

Update the version in `internal/version/version.go` if needed (though it's typically set via build flags).

### 2. Create Git Tag

Create and push a new git tag:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

### 3. Run Release

```bash
make release
```

This will:
- Build binaries for Linux, macOS, and Windows (amd64 and arm64)
- Create a GitHub release with release notes
- Upload checksums

### 4. Verify Release

1. Check GitHub releases: https://github.com/anmho/create-go-api/releases
2. Test Go install:
   ```bash
   go install github.com/anmho/create-go-api@latest
   ```

## Testing Releases

Before creating a real release, test with a snapshot:

```bash
make release-snapshot
```

This creates a snapshot release without creating a git tag or GitHub release.

## Troubleshooting

### goreleaser fails with authentication error

Make sure `GITHUB_TOKEN` is set and has the correct permissions:
- `repo` scope for private repos
- `public_repo` scope for public repos


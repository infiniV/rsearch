# CI/CD Quick Start Guide

Quick reference for common CI/CD tasks in rsearch.

## Quick Commands

### Local Development

```bash
# Run all checks before committing
make test && golangci-lint run && make build

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run linter
golangci-lint run

# Build binary
go build -o bin/rsearch cmd/rsearch/main.go

# Generate documentation
go run ./cmd/gendocs
```

### Creating a Release

```bash
# Using release script (recommended)
./scripts/release.sh v1.2.3

# Preview without creating (dry run)
./scripts/release.sh v1.2.3 --dry-run

# Manual release
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### Working with Docker

```bash
# Build locally
docker build -t rsearch:dev .

# Build with version info
docker build -t rsearch:v1.2.3 \
  --build-arg VERSION=v1.2.3 \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  .

# Run locally
docker run -p 8080:8080 rsearch:dev

# Pull from registry
docker pull ghcr.io/infiniv/rsearch:latest
docker pull ghcr.io/infiniv/rsearch:v1.2.3
```

## Workflow Status

### Check Workflow Runs

- All workflows: https://github.com/infiniv/rsearch/actions
- CI workflow: https://github.com/infiniv/rsearch/actions/workflows/ci.yaml
- Release workflow: https://github.com/infiniv/rsearch/actions/workflows/release.yaml
- CodeQL workflow: https://github.com/infiniv/rsearch/actions/workflows/codeql.yaml

### Adding Status Badges to README

```markdown
[![CI](https://github.com/infiniv/rsearch/actions/workflows/ci.yaml/badge.svg)](https://github.com/infiniv/rsearch/actions/workflows/ci.yaml)
[![Release](https://github.com/infiniv/rsearch/actions/workflows/release.yaml/badge.svg)](https://github.com/infiniv/rsearch/actions/workflows/release.yaml)
[![CodeQL](https://github.com/infiniv/rsearch/actions/workflows/codeql.yaml/badge.svg)](https://github.com/infiniv/rsearch/actions/workflows/codeql.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/infiniv/rsearch)](https://goreportcard.com/report/github.com/infiniv/rsearch)
```

## Common Scenarios

### Fix Failing Tests

```bash
# Run specific test
go test -run TestName ./internal/package/...

# Run with verbose output
go test -v ./internal/package/...

# Run with race detection
go test -race ./...
```

### Fix Linting Issues

```bash
# Run linter
golangci-lint run

# Auto-fix issues (where possible)
golangci-lint run --fix

# Run specific linters
golangci-lint run --enable=errcheck,gosimple
```

### Update Dependencies

```bash
# Update all dependencies
go get -u ./...
go mod tidy

# Update specific dependency
go get -u github.com/user/package@latest

# Verify dependencies
go mod verify
```

### Fix Documentation

```bash
# Regenerate documentation
go run ./cmd/gendocs

# Check what changed
git diff docs/

# Commit if needed
git add docs/
git commit -m "docs: update generated documentation"
```

### Troubleshoot Release

```bash
# Check if tag exists
git tag -l "v1.2.3"

# Delete local tag
git tag -d v1.2.3

# Delete remote tag
git push origin :refs/tags/v1.2.3

# View tag details
git show v1.2.3

# List all tags
git tag -l
```

### Debug Docker Build

```bash
# Build with verbose output
docker build --progress=plain -t rsearch:test .

# Build specific stage
docker build --target=builder -t rsearch:builder .

# Check image
docker images rsearch:test
docker inspect rsearch:test

# Run with shell
docker run -it rsearch:test /bin/sh
```

## Environment Variables

### CI Environment

Automatically set by GitHub Actions:
- `GITHUB_SHA`: Commit hash
- `GITHUB_REF`: Branch or tag reference
- `GITHUB_REPOSITORY`: Repository name (owner/repo)
- `GITHUB_ACTOR`: User who triggered workflow

### Application Environment

For rsearch configuration:
```bash
RSEARCH_SERVER_PORT=8080
RSEARCH_LOGGING_LEVEL=debug
RSEARCH_LOGGING_FORMAT=console
RSEARCH_METRICS_ENABLED=true
```

## File Locations

### Configuration Files
- `.github/workflows/ci.yaml` - CI workflow
- `.github/workflows/release.yaml` - Release workflow
- `.github/workflows/codeql.yaml` - Security analysis
- `.golangci.yaml` - Linter configuration
- `.github/dependabot.yml` - Dependency updates
- `.github/CODEOWNERS` - Code ownership

### Scripts
- `scripts/release.sh` - Release automation
- `Makefile` - Build automation

### Documentation
- `docs/CI_CD.md` - Full CI/CD documentation
- `docs/CI_CD_QUICK_START.md` - This file
- `README.md` - Project overview

## Getting Help

### CI/CD Issues

1. Check workflow logs in GitHub Actions
2. Review the CI/CD documentation: `docs/CI_CD.md`
3. Search existing issues: https://github.com/infiniv/rsearch/issues
4. Create new issue using bug report template

### Release Issues

1. Run release script with `--dry-run` first
2. Check git status: `git status`
3. Verify remote: `git remote -v`
4. Review release workflow logs

### Build Issues

1. Ensure Go version matches `go.mod`
2. Update dependencies: `go mod tidy`
3. Clear build cache: `go clean -cache`
4. Check for platform-specific issues

## Best Practices Checklist

### Before Committing

- [ ] Run tests: `go test ./...`
- [ ] Run linter: `golangci-lint run`
- [ ] Update docs if needed: `go run ./cmd/gendocs`
- [ ] Write clear commit message
- [ ] Reference issue numbers

### Before Creating PR

- [ ] Rebase on latest main: `git rebase origin/main`
- [ ] All tests pass locally
- [ ] No linting issues
- [ ] Documentation updated
- [ ] Added tests for new features
- [ ] Filled out PR template

### Before Releasing

- [ ] All tests pass
- [ ] Documentation up-to-date
- [ ] CHANGELOG updated (if used)
- [ ] Version number follows SemVer
- [ ] Breaking changes documented
- [ ] Migration guide provided (if needed)

## Useful Git Commands

```bash
# View commit history
git log --oneline --graph --all

# View changes since last tag
git log --pretty=format:"%h %s" $(git describe --tags --abbrev=0)..HEAD

# View diff between tags
git diff v1.0.0..v1.1.0

# Cherry-pick commit
git cherry-pick <commit-hash>

# Revert commit
git revert <commit-hash>

# Interactive rebase
git rebase -i HEAD~3
```

## Monitoring

### Check CI Status

```bash
# Using GitHub CLI (gh)
gh run list --workflow=ci.yaml
gh run view <run-id>
gh run watch

# View recent failures
gh run list --workflow=ci.yaml --status=failure
```

### Check Release Status

```bash
# List releases
gh release list

# View specific release
gh release view v1.2.3

# Download release assets
gh release download v1.2.3
```

### Check Container Images

```bash
# List images
docker images ghcr.io/infiniv/rsearch

# Pull latest
docker pull ghcr.io/infiniv/rsearch:latest

# Check image details
docker inspect ghcr.io/infiniv/rsearch:v1.2.3
```

## Tips and Tricks

### Speed Up Local Tests

```bash
# Run tests in parallel
go test -parallel 4 ./...

# Run only short tests
go test -short ./...

# Cache test results
go test -count=1 ./...  # Disable cache
go clean -testcache     # Clear cache
```

### Speed Up CI

- Commit `go.sum` for reproducible builds
- Use Go modules caching in workflows
- Run independent jobs in parallel
- Use matrix builds efficiently

### Debug Workflow Issues

```bash
# Run workflow locally (requires act)
act -j test

# Validate workflow syntax
gh workflow view ci.yaml

# Manually trigger workflow
gh workflow run ci.yaml
```

## Resources

- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [golangci-lint Docs](https://golangci-lint.run/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

# CI/CD Documentation

This document describes the continuous integration and continuous deployment setup for rsearch.

## Overview

The rsearch project uses GitHub Actions for CI/CD automation. The setup includes:

- **Continuous Integration (CI)**: Automated testing, linting, and building on every push and pull request
- **Continuous Deployment (CD)**: Automated releases with cross-compiled binaries and Docker images
- **Dependency Management**: Automated dependency updates via Dependabot
- **Code Quality**: Automated linting with golangci-lint

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yaml`)

Triggers on:
- Push to `main` or `master` branch
- Pull requests to `main` or `master` branch

#### Jobs

**Lint Job**
- Runs `golangci-lint` to check code quality
- Uses configuration from `.golangci.yaml`
- Enforces coding standards and best practices

**Test Job**
- Runs on Go versions: 1.21, 1.22, 1.23
- Executes all tests with race detection
- Generates coverage reports
- Uploads coverage to Codecov (optional)
- Matrix strategy ensures compatibility across Go versions

**Build Job**
- Cross-compiles binaries for multiple platforms:
  - Linux (amd64, arm64)
  - macOS/Darwin (amd64, arm64)
  - Windows (amd64)
- Produces static binaries with version information
- Uploads binaries as artifacts for download

**Integration Test Job**
- Runs integration tests (if present)
- Executes after lint, test, and build complete

**Validate Documentation Job**
- Regenerates documentation using `cmd/gendocs`
- Ensures documentation is up-to-date
- Fails if documentation needs updating

**All Checks Job**
- Aggregates results from all jobs
- Provides single status check for branch protection

### 2. Release Workflow (`.github/workflows/release.yaml`)

Triggers on:
- Push of tags matching pattern `v*` (e.g., `v1.0.0`, `v1.2.3-beta.1`)

#### Jobs

**Build Binaries Job**
- Cross-compiles release binaries for all platforms
- Embeds version information via ldflags:
  - `main.version`: Git tag (e.g., `v1.0.0`)
  - `main.commit`: Git commit hash
  - `main.date`: Build timestamp
- Creates compressed archives (`.tar.gz` for Unix, `.zip` for Windows)
- Uploads artifacts for release

**Build Docker Job**
- Builds multi-platform Docker images (linux/amd64, linux/arm64)
- Pushes to GitHub Container Registry (ghcr.io)
- Tags images with:
  - Semantic version (e.g., `1.0.0`)
  - Major.minor version (e.g., `1.0`)
  - Major version (e.g., `1`)
  - `latest` (for default branch)
- Uses Docker Buildx for multi-platform support
- Leverages GitHub Actions cache for faster builds

**Create Release Job**
- Generates changelog from commit history
- Creates GitHub release with:
  - Release notes
  - Changelog since last tag
  - Installation instructions
  - Binary download links
- Attaches binary archives as release assets
- Generates release notes automatically

**Verify Release Job**
- Verifies Docker image was pushed successfully
- Validates release was created
- Ensures artifacts are accessible

## Configuration Files

### `.golangci.yaml`

Linter configuration with enabled checks:
- **Error checking**: `errcheck`, `errorlint`, `nilerr`
- **Simplification**: `gosimple`, `unconvert`, `unparam`
- **Analysis**: `staticcheck`, `govet`, `revive`
- **Security**: `gosec`
- **Code quality**: `gocyclo`, `dupl`, `goconst`
- **Formatting**: `gofmt`, `goimports`, `whitespace`
- **Best practices**: `ineffassign`, `unused`, `exportloopref`

Special rules:
- Tests excluded from certain checks (complexity, duplication)
- Generated files excluded
- Configurable complexity threshold (15)

### `.github/dependabot.yml`

Automated dependency updates for:
- **Go modules**: Weekly updates on Monday mornings
- **GitHub Actions**: Weekly updates on Monday mornings
- **Docker**: Weekly updates on Monday mornings

Grouped updates:
- Go standard library packages
- Testing libraries
- Observability libraries
- All GitHub Actions

### `.github/CODEOWNERS`

Defines code ownership for automatic review requests:
- Default owner: `@infiniv`
- Specific ownership for core packages, docs, CI/CD, and tests

## Release Process

### Automated Release (Recommended)

Use the release script for guided release creation:

```bash
# Create a release (performs checks)
./scripts/release.sh v1.2.3

# Dry run (preview without creating)
./scripts/release.sh v1.2.3 --dry-run

# Force release (skip checks)
./scripts/release.sh v1.2.3 --force
```

The script:
1. Validates version format (semantic versioning)
2. Checks git status (clean working directory)
3. Verifies remote is up-to-date
4. Ensures tag doesn't exist
5. Runs tests and build
6. Creates annotated git tag with changelog
7. Pushes tag to trigger release workflow

### Manual Release

```bash
# Ensure you're on main/master and up-to-date
git checkout main
git pull origin main

# Run tests
go test ./...

# Create and push tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### Version Format

Follow semantic versioning (SemVer):
- **Format**: `v<major>.<minor>.<patch>[-prerelease][+buildmetadata]`
- **Examples**:
  - `v1.0.0` - Major release
  - `v1.2.3` - Patch release
  - `v2.0.0-beta.1` - Pre-release
  - `v1.0.0+20240101` - Build metadata

### After Release

1. **Monitor GitHub Actions**: Check workflow runs at https://github.com/infiniv/rsearch/actions
2. **Verify Release**: Check release page at https://github.com/infiniv/rsearch/releases
3. **Test Artifacts**:
   - Download and test binary: `./rsearch-linux-amd64 --version`
   - Pull and test Docker image: `docker pull ghcr.io/infiniv/rsearch:v1.2.3`
4. **Update Documentation**: Update README, CHANGELOG, or other docs if needed
5. **Announce**: Notify users through appropriate channels

## Docker Images

### Pulling Images

```bash
# Specific version
docker pull ghcr.io/infiniv/rsearch:v1.2.3
docker pull ghcr.io/infiniv/rsearch:1.2
docker pull ghcr.io/infiniv/rsearch:1

# Latest
docker pull ghcr.io/infiniv/rsearch:latest
```

### Running Containers

```bash
# Basic usage
docker run -p 8080:8080 ghcr.io/infiniv/rsearch:latest

# With configuration
docker run -p 8080:8080 \
  -e RSEARCH_SERVER_PORT=8080 \
  -e RSEARCH_LOGGING_LEVEL=debug \
  ghcr.io/infiniv/rsearch:latest

# With config file
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  ghcr.io/infiniv/rsearch:latest --config /app/config.yaml
```

### Multi-Platform Support

Images are built for:
- `linux/amd64` - x86-64 architecture
- `linux/arm64` - ARM64 architecture (Apple Silicon, ARM servers)

Docker automatically pulls the correct image for your platform.

## Binary Downloads

Binaries are available as GitHub Release assets:

### Platforms

- **Linux**: `rsearch-linux-amd64.tar.gz`, `rsearch-linux-arm64.tar.gz`
- **macOS**: `rsearch-darwin-amd64.tar.gz`, `rsearch-darwin-arm64.tar.gz`
- **Windows**: `rsearch-windows-amd64.zip`

### Installation

```bash
# Linux/macOS
wget https://github.com/infiniv/rsearch/releases/download/v1.2.3/rsearch-linux-amd64.tar.gz
tar -xzf rsearch-linux-amd64.tar.gz
chmod +x rsearch-linux-amd64
sudo mv rsearch-linux-amd64 /usr/local/bin/rsearch

# Verify
rsearch --version
```

## Development Workflow

### Pre-commit Checks

Before committing, run:

```bash
# Run tests
make test

# Run linter
golangci-lint run

# Build
make build

# All checks
make test && golangci-lint run && make build
```

### Pull Request Process

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes and commit
3. Push branch: `git push origin feature/my-feature`
4. Open Pull Request on GitHub
5. CI automatically runs:
   - Linting
   - Tests on multiple Go versions
   - Cross-platform builds
   - Documentation validation
6. Address any CI failures
7. Get code review from CODEOWNERS
8. Merge when approved and CI passes

### Local Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/parser/...

# Run specific test
go test -run TestParsePhraseQuery ./internal/parser/...

# Run integration tests
go test -v -tags=integration ./tests/integration/...
```

## Troubleshooting

### CI Failures

**Lint Failures**
- Run `golangci-lint run` locally
- Fix reported issues
- Check `.golangci.yaml` for enabled linters

**Test Failures**
- Run `go test -v ./...` locally
- Check test output for specific failures
- Run failing test individually: `go test -run TestName ./package/...`

**Build Failures**
- Verify code compiles: `go build ./cmd/rsearch`
- Check for platform-specific issues
- Review build logs in GitHub Actions

**Documentation Out of Date**
- Run `go run ./cmd/gendocs`
- Commit generated documentation changes
- Ensure test cases match documentation

### Release Failures

**Tag Already Exists**
- Delete local tag: `git tag -d v1.2.3`
- Delete remote tag: `git push origin :refs/tags/v1.2.3`
- Create tag again

**Docker Build Fails**
- Check Dockerfile syntax
- Verify build context includes all necessary files
- Test locally: `docker build -t rsearch:test .`

**Release Workflow Fails**
- Check workflow logs in GitHub Actions
- Re-run workflow from GitHub UI
- Verify GitHub token has necessary permissions

### Dependabot Issues

**Merge Conflicts**
- Update your branch with latest main
- Resolve conflicts manually
- Dependabot will rebase automatically

**Failed Tests**
- Review Dependabot PR for breaking changes
- Update code to work with new dependency version
- Or close PR and pin dependency version if needed

## Security

### Secrets Management

Required secrets (configured in GitHub repository settings):
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- Optional: `CODECOV_TOKEN` for coverage uploads

### Container Security

- Runs as non-root user (`rsearch:1000`)
- Based on minimal Alpine Linux image
- Only includes necessary runtime dependencies
- Regular updates via Dependabot

### Supply Chain Security

- Go module verification: `go mod verify`
- Reproducible builds with `-trimpath`
- Static binaries (CGO_ENABLED=0)
- Minimal runtime dependencies

## Monitoring

### CI Status

- View all workflows: https://github.com/infiniv/rsearch/actions
- Branch status badges (can be added to README)
- Commit status checks in pull requests

### Release Status

- Releases page: https://github.com/infiniv/rsearch/releases
- Container registry: https://github.com/infiniv/rsearch/pkgs/container/rsearch
- Download statistics in release assets

### Coverage Tracking

- Codecov dashboard (if configured)
- Coverage reports in CI artifacts
- Coverage trends over time

## Best Practices

### Commits

- Write clear, descriptive commit messages
- Reference issue numbers when applicable
- Keep commits focused and atomic
- Use conventional commit format (optional)

### Versioning

- Follow semantic versioning strictly
- Major version for breaking changes
- Minor version for new features
- Patch version for bug fixes
- Use pre-release tags for testing

### Documentation

- Update documentation with code changes
- Keep CHANGELOG up-to-date (optional)
- Regenerate docs: `go run ./cmd/gendocs`
- Include examples for new features

### Testing

- Write tests for all new features
- Maintain high test coverage
- Test edge cases and error conditions
- Use table-driven tests for multiple scenarios

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [Dependabot Documentation](https://docs.github.com/en/code-security/dependabot)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)

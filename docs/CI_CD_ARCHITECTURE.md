# CI/CD Architecture

Visual representation of the CI/CD pipeline for rsearch.

## Pipeline Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                          rsearch CI/CD Pipeline                      │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────┐
│ Developer    │
│ Commits Code │
└──────┬───────┘
       │
       ├─────────────────┐
       │                 │
       ▼                 ▼
┌────────────┐    ┌─────────────┐
│ Push to    │    │ Create      │
│ main/master│    │ Pull Request│
└─────┬──────┘    └──────┬──────┘
      │                  │
      └────────┬─────────┘
               │
               ▼
      ┌────────────────┐
      │  CI Workflow   │
      └────────┬───────┘
               │
     ┌─────────┼─────────┬─────────┬───────────┐
     │         │         │         │           │
     ▼         ▼         ▼         ▼           ▼
┌────────┐ ┌──────┐ ┌───────┐ ┌──────────┐ ┌─────────┐
│  Lint  │ │ Test │ │ Build │ │Integration│ │Validate │
│        │ │      │ │       │ │   Tests   │ │  Docs   │
└────────┘ └──────┘ └───────┘ └──────────┘ └─────────┘
                                                 │
                                                 ▼
                                         ┌───────────────┐
                                         │  All Checks   │
                                         │    Passed     │
                                         └───────┬───────┘
                                                 │
                                                 ▼
                                         ┌───────────────┐
                                         │  Merge Ready  │
                                         └───────────────┘

┌──────────────┐
│  Maintainer  │
│  Creates Tag │
│  (v1.2.3)    │
└──────┬───────┘
       │
       ▼
┌─────────────────┐
│ Release Workflow│
└────────┬────────┘
         │
    ┌────┼────┬───────────┐
    │    │    │           │
    ▼    ▼    ▼           ▼
┌────┐ ┌────┐ ┌──────┐ ┌───────┐
│Build│ │Build│ │Create│ │Verify│
│Bins │ │Docker│ │Release│ │      │
└────┘ └────┘ └──────┘ └───────┘
    │    │      │
    └────┼──────┘
         │
         ▼
   ┌──────────┐
   │ Published│
   │ Release  │
   └──────────┘
```

## Workflow Details

### CI Workflow (Continuous Integration)

```
Trigger: Push | Pull Request to main/master
│
├─ Lint Job
│  ├─ Checkout code
│  ├─ Setup Go (version from go.mod)
│  ├─ Run golangci-lint
│  └─ Report results
│
├─ Test Job (Matrix: Go 1.21, 1.22, 1.23)
│  ├─ Checkout code
│  ├─ Setup Go
│  ├─ Download dependencies
│  ├─ Run tests with race detection
│  ├─ Generate coverage report
│  └─ Upload to Codecov (optional)
│
├─ Build Job (Matrix: 6 platforms)
│  ├─ Platforms:
│  │  ├─ linux/amd64
│  │  ├─ linux/arm64
│  │  ├─ darwin/amd64
│  │  ├─ darwin/arm64
│  │  ├─ windows/amd64
│  │  └─ windows/arm64
│  ├─ Checkout code
│  ├─ Setup Go
│  ├─ Cross-compile binary
│  └─ Upload artifacts
│
├─ Integration Test Job
│  ├─ Wait for: lint, test, build
│  ├─ Run integration tests (if present)
│  └─ Report results
│
├─ Validate Docs Job
│  ├─ Regenerate documentation
│  ├─ Check for differences
│  └─ Fail if out of date
│
└─ All Checks Job
   ├─ Wait for: all jobs
   ├─ Check all passed
   └─ Set status (pass/fail)
```

### Release Workflow

```
Trigger: Git tag matching v*
│
├─ Build Binaries Job (Matrix: 6 platforms)
│  ├─ Checkout code (full history)
│  ├─ Setup Go
│  ├─ Extract version from tag
│  ├─ Build with version info:
│  │  ├─ -X main.version=$VERSION
│  │  ├─ -X main.commit=$COMMIT
│  │  └─ -X main.date=$DATE
│  ├─ Create archives (.tar.gz/.zip)
│  └─ Upload artifacts
│
├─ Build Docker Job
│  ├─ Checkout code
│  ├─ Setup Docker Buildx
│  ├─ Login to ghcr.io
│  ├─ Extract metadata & tags
│  ├─ Build multi-arch image:
│  │  ├─ linux/amd64
│  │  └─ linux/arm64
│  └─ Push to registry
│
├─ Create Release Job
│  ├─ Wait for: build-binaries, build-docker
│  ├─ Download artifacts
│  ├─ Generate changelog
│  ├─ Create GitHub release:
│  │  ├─ Release notes
│  │  ├─ Changelog
│  │  ├─ Installation guide
│  │  └─ Binary assets
│  └─ Publish release
│
└─ Verify Release Job
   ├─ Wait for: create-release
   ├─ Pull Docker image
   ├─ Test Docker image
   └─ Verify GitHub release exists
```

### CodeQL Security Workflow

```
Trigger: Push | Pull Request | Weekly Schedule (Monday 2 AM UTC)
│
└─ Analyze Job
   ├─ Checkout code
   ├─ Initialize CodeQL
   ├─ Setup Go
   ├─ Build application
   ├─ Run security analysis:
   │  ├─ Security vulnerabilities
   │  ├─ Code quality issues
   │  └─ Best practice violations
   └─ Upload results to GitHub Security
```

## Data Flow

### Source Code to Production

```
┌──────────────┐
│   Developer  │
│  Local Work  │
└──────┬───────┘
       │ git push
       ▼
┌──────────────┐
│   GitHub     │
│  Repository  │
└──────┬───────┘
       │ webhook
       ▼
┌──────────────┐
│ GitHub       │
│ Actions      │
│ (CI/CD)      │
└──────┬───────┘
       │
       ├─ Lint & Test
       ├─ Build Binaries
       ├─ Build Docker
       └─ Create Release
       │
       ▼
┌──────────────┐     ┌──────────────┐
│   GitHub     │     │   GitHub     │
│  Releases    │     │  Container   │
│  (Binaries)  │     │  Registry    │
└──────┬───────┘     └──────┬───────┘
       │                    │
       │ download           │ docker pull
       ▼                    ▼
┌──────────────┐     ┌──────────────┐
│    Users     │     │    Users     │
│  (Binaries)  │     │   (Docker)   │
└──────────────┘     └──────────────┘
```

### Artifact Flow

```
┌─────────────┐
│   Source    │
│    Code     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Build     │
│   Process   │
└──────┬──────┘
       │
       ├─────────────────┬─────────────────┐
       │                 │                 │
       ▼                 ▼                 ▼
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Binary  │     │  Docker  │     │ Coverage │
│ Archives │     │  Image   │     │  Report  │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     ▼                ▼                ▼
┌──────────┐     ┌──────────┐     ┌──────────┐
│ GitHub   │     │  ghcr.io │     │ Codecov  │
│ Release  │     │ Registry │     │  (opt)   │
└──────────┘     └──────────┘     └──────────┘
```

## Component Interactions

### Dependency Management

```
┌──────────────┐
│  Dependabot  │
│   (Weekly)   │
└──────┬───────┘
       │
       ├─ Scan: Go modules
       ├─ Scan: GitHub Actions
       └─ Scan: Docker images
       │
       ▼
┌──────────────┐
│    Create    │
│ Pull Request │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  CI Workflow │
│     Runs     │
└──────┬───────┘
       │
       ├─ Tests pass?
       ├─ Builds pass?
       └─ No conflicts?
       │
       ▼
┌──────────────┐
│    Review    │
│   & Merge    │
└──────────────┘
```

### Code Quality Gate

```
┌──────────────┐
│    Commit    │
│    Pushed    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  golangci-   │
│    lint      │
└──────┬───────┘
       │
       ├─ errcheck
       ├─ gosimple
       ├─ govet
       ├─ staticcheck
       ├─ gosec
       ├─ revive
       └─ 20+ more
       │
       ├─ PASS ────────┐
       │               ▼
       │         ┌──────────┐
       │         │ Continue │
       │         │ Pipeline │
       │         └──────────┘
       │
       └─ FAIL ────────┐
                       ▼
                 ┌──────────┐
                 │   Stop   │
                 │  Report  │
                 └──────────┘
```

## Security Architecture

### Security Scanning

```
┌────────────────────────────────────────┐
│        Security Scanning Layers        │
└────────────────────────────────────────┘

Layer 1: Code Analysis
├─ CodeQL (GitHub)
│  ├─ Security vulnerabilities
│  ├─ Code quality issues
│  └─ Best practices
│
Layer 2: Dependency Scanning
├─ Dependabot
│  ├─ Go module vulnerabilities
│  ├─ GitHub Actions versions
│  └─ Docker base images
│
Layer 3: Linting Security
├─ gosec (via golangci-lint)
│  ├─ Security issues in code
│  ├─ Unsafe patterns
│  └─ Cryptographic weaknesses
│
Layer 4: Container Security
└─ Docker Best Practices
   ├─ Non-root user
   ├─ Minimal base image
   └─ No secrets in layers
```

### Authentication & Authorization

```
┌──────────────┐
│   GitHub     │
│   Actions    │
└──────┬───────┘
       │
       ├─ Uses: GITHUB_TOKEN (automatic)
       │  ├─ Read: repository
       │  ├─ Write: packages (Docker)
       │  └─ Write: releases
       │
       └─ Optional: CODECOV_TOKEN
          └─ Upload: coverage reports
```

## Monitoring & Observability

### CI/CD Observability

```
┌────────────────────────────────────────┐
│          Monitoring Points             │
└────────────────────────────────────────┘

Workflow Level
├─ GitHub Actions Dashboard
│  ├─ Workflow runs
│  ├─ Job status
│  └─ Run duration
│
Job Level
├─ Individual job logs
│  ├─ Step output
│  ├─ Error messages
│  └─ Artifact downloads
│
Code Level
├─ Test results
│  ├─ Pass/fail status
│  ├─ Coverage percentage
│  └─ Race conditions
│
├─ Lint results
│  ├─ Issue count
│  ├─ Severity levels
│  └─ File locations
│
Quality Level
├─ CodeQL alerts
│  ├─ Security issues
│  ├─ Quality issues
│  └─ Trend over time
│
Dependency Level
└─ Dependabot alerts
   ├─ Vulnerability CVEs
   ├─ Update status
   └─ PR status
```

## Performance Characteristics

### Build Times (Approximate)

```
CI Workflow (per job):
├─ Lint:         ~2-3 minutes
├─ Test:         ~3-5 minutes (per Go version)
├─ Build:        ~2-3 minutes (per platform)
├─ Integration:  ~5-10 minutes (if present)
└─ Validate:     ~1-2 minutes

Total CI Time:   ~5-8 minutes (parallel execution)

Release Workflow:
├─ Build Bins:   ~2-3 minutes (per platform)
├─ Build Docker: ~5-8 minutes (multi-arch)
├─ Create Release: ~2-3 minutes
└─ Verify:       ~1-2 minutes

Total Release:   ~10-15 minutes (parallel execution)

CodeQL Workflow:
└─ Analyze:      ~5-10 minutes

Weekly Dependabot:
└─ Scan & PR:    ~2-5 minutes
```

### Caching Strategy

```
┌────────────────────────────────────────┐
│           Cache Hierarchy              │
└────────────────────────────────────────┘

Level 1: Go Module Cache
├─ Keyed by: go.sum hash
├─ Speeds up: Dependency downloads
└─ Savings: ~1-2 minutes per job

Level 2: Docker Layer Cache
├─ Keyed by: Dockerfile + context
├─ Speeds up: Docker builds
└─ Savings: ~3-5 minutes per build

Level 3: Action Cache
├─ setup-go: Go installation
├─ actions/cache: Custom caching
└─ Savings: ~30-60 seconds per job
```

## Scalability

### Parallel Execution

```
CI Workflow Parallelism:
├─ Lint:         1 instance
├─ Test:         3 instances (Go 1.21, 1.22, 1.23)
├─ Build:        6 instances (all platforms)
└─ Total:        10 concurrent jobs

Max GitHub Actions Concurrent Jobs:
├─ Free tier:    20 concurrent jobs
├─ Pro tier:     60 concurrent jobs
└─ Enterprise:   180 concurrent jobs

Resource Usage:
├─ Per job:      ~2 CPU, 7 GB RAM
└─ Total peak:   ~20 CPU, 70 GB RAM
```

## Disaster Recovery

### Rollback Procedures

```
Release Rollback:
├─ Delete bad release:
│  └─ gh release delete v1.2.3
│
├─ Delete bad tag:
│  ├─ git tag -d v1.2.3
│  └─ git push origin :refs/tags/v1.2.3
│
├─ Delete Docker images:
│  └─ Via GitHub UI (Packages)
│
└─ Re-release:
   ├─ Fix issues
   └─ Create new tag v1.2.4
```

## Future Enhancements

Potential additions to the pipeline:

```
Short Term:
├─ Container vulnerability scanning (Trivy/Snyk)
├─ Performance benchmarking
├─ E2E tests with PostgreSQL
└─ Automated changelog generation

Medium Term:
├─ Staging environment deployment
├─ Canary deployments
├─ Load testing
└─ API documentation generation

Long Term:
├─ Multi-region deployments
├─ Blue-green deployments
├─ Automated rollback on metrics
└─ Integration with monitoring (DataDog/New Relic)
```

## References

- GitHub Actions: https://docs.github.com/en/actions
- golangci-lint: https://golangci-lint.run/
- CodeQL: https://codeql.github.com/
- Dependabot: https://docs.github.com/en/code-security/dependabot
- Docker Buildx: https://docs.docker.com/buildx/

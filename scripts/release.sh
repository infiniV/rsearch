#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

usage() {
    cat <<EOF
Usage: $0 <version> [options]

Create a new release for rsearch.

Arguments:
    version         Version to release (e.g., v1.0.0, v1.2.3)

Options:
    -h, --help      Show this help message
    -n, --dry-run   Perform a dry run without creating tags or pushing
    -f, --force     Force creation even if checks fail

Examples:
    $0 v1.0.0
    $0 v1.2.3 --dry-run
    $0 v2.0.0-beta.1

EOF
    exit 1
}

validate_version() {
    local version="$1"

    # Check if version matches semantic versioning pattern
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        print_error "Invalid version format: $version"
        print_info "Version must follow semantic versioning: v<major>.<minor>.<patch>[-prerelease][+buildmetadata]"
        print_info "Examples: v1.0.0, v1.2.3, v2.0.0-beta.1, v1.0.0+20240101"
        exit 1
    fi

    print_success "Version format is valid: $version"
}

check_git_status() {
    cd "$PROJECT_ROOT"

    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not a git repository"
        exit 1
    fi

    # Check if working directory is clean
    if [[ -n $(git status --porcelain) ]]; then
        print_error "Working directory is not clean. Please commit or stash changes."
        git status --short
        exit 1
    fi

    # Check if we're on main/master branch
    local current_branch
    current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
        print_warning "Current branch is '$current_branch', not 'main' or 'master'"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    print_success "Git status is clean"
}

check_tag_exists() {
    local version="$1"

    if git rev-parse "$version" >/dev/null 2>&1; then
        print_error "Tag '$version' already exists"
        exit 1
    fi

    print_success "Tag '$version' does not exist"
}

check_remote() {
    cd "$PROJECT_ROOT"

    # Check if remote 'origin' exists
    if ! git remote get-url origin > /dev/null 2>&1; then
        print_error "Git remote 'origin' not found"
        exit 1
    fi

    # Fetch latest from remote
    print_info "Fetching latest from remote..."
    git fetch origin --tags

    # Check if local is up to date with remote
    local local_commit
    local remote_commit
    local_commit=$(git rev-parse HEAD)
    remote_commit=$(git rev-parse origin/$(git rev-parse --abbrev-ref HEAD) 2>/dev/null || echo "")

    if [[ -n "$remote_commit" && "$local_commit" != "$remote_commit" ]]; then
        print_error "Local branch is not up to date with remote"
        print_info "Run 'git pull' to update"
        exit 1
    fi

    print_success "Repository is up to date with remote"
}

run_tests() {
    cd "$PROJECT_ROOT"

    print_info "Running tests..."
    if ! go test -v ./...; then
        print_error "Tests failed"
        exit 1
    fi

    print_success "All tests passed"
}

run_build() {
    cd "$PROJECT_ROOT"

    print_info "Running build..."
    if ! go build -v ./cmd/rsearch; then
        print_error "Build failed"
        exit 1
    fi

    print_success "Build successful"
}

create_tag() {
    local version="$1"
    local dry_run="$2"

    cd "$PROJECT_ROOT"

    # Generate changelog
    local prev_tag
    prev_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

    local changelog_header="Release $version"
    local changelog_body

    if [[ -z "$prev_tag" ]]; then
        changelog_body="Initial release of rsearch"
    else
        print_info "Generating changelog since $prev_tag..."
        changelog_body=$(git log --pretty=format:"- %s (%h)" "$prev_tag"..HEAD)
    fi

    local tag_message
    tag_message=$(cat <<EOF
$changelog_header

$changelog_body

Released: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
EOF
)

    if [[ "$dry_run" == "true" ]]; then
        print_info "DRY RUN: Would create tag '$version' with message:"
        echo "$tag_message"
        return
    fi

    # Create annotated tag
    print_info "Creating tag '$version'..."
    git tag -a "$version" -m "$tag_message"

    print_success "Tag '$version' created successfully"
}

push_tag() {
    local version="$1"
    local dry_run="$2"

    cd "$PROJECT_ROOT"

    if [[ "$dry_run" == "true" ]]; then
        print_info "DRY RUN: Would push tag '$version' to origin"
        return
    fi

    print_info "Pushing tag '$version' to origin..."
    git push origin "$version"

    print_success "Tag '$version' pushed successfully"
}

print_next_steps() {
    local version="$1"

    cat <<EOF

${GREEN}Release process completed successfully!${NC}

${BLUE}Next steps:${NC}

1. Monitor the GitHub Actions workflow:
   https://github.com/infiniv/rsearch/actions

2. The following will be created automatically:
   - Release binaries for multiple platforms
   - Docker image: ghcr.io/infiniv/rsearch:$version
   - GitHub release with changelog

3. After the release workflow completes:
   - Verify the release at: https://github.com/infiniv/rsearch/releases/tag/$version
   - Test the Docker image: docker pull ghcr.io/infiniv/rsearch:$version
   - Update documentation if needed

4. Announce the release:
   - Update project README if needed
   - Notify users through appropriate channels

${YELLOW}Troubleshooting:${NC}
If the release workflow fails, you can:
- Check the Actions logs for errors
- Re-run the workflow from the GitHub UI
- Delete and recreate the tag if needed: git tag -d $version && git push origin :refs/tags/$version

EOF
}

# Main script
main() {
    local version=""
    local dry_run="false"
    local force="false"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                ;;
            -n|--dry-run)
                dry_run="true"
                shift
                ;;
            -f|--force)
                force="true"
                shift
                ;;
            -*)
                print_error "Unknown option: $1"
                usage
                ;;
            *)
                if [[ -z "$version" ]]; then
                    version="$1"
                else
                    print_error "Too many arguments"
                    usage
                fi
                shift
                ;;
        esac
    done

    # Check if version is provided
    if [[ -z "$version" ]]; then
        print_error "Version argument is required"
        usage
    fi

    # Banner
    cat <<EOF
${BLUE}========================================${NC}
${BLUE}    rsearch Release Script${NC}
${BLUE}========================================${NC}

Version: ${GREEN}$version${NC}
Dry Run: ${YELLOW}$dry_run${NC}
Force:   ${YELLOW}$force${NC}

EOF

    # Validate version format
    validate_version "$version"

    # Perform checks
    if [[ "$force" == "false" ]]; then
        check_git_status
        check_remote
        check_tag_exists "$version"
        run_tests
        run_build
    else
        print_warning "Force mode enabled, skipping some checks"
    fi

    # Confirm release
    if [[ "$dry_run" == "false" ]]; then
        echo
        print_warning "About to create and push tag '$version'"
        read -p "Continue? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Release cancelled"
            exit 0
        fi
    fi

    # Create and push tag
    create_tag "$version" "$dry_run"
    push_tag "$version" "$dry_run"

    # Print next steps
    if [[ "$dry_run" == "false" ]]; then
        print_next_steps "$version"
    else
        print_info "DRY RUN COMPLETE - No changes were made"
    fi
}

main "$@"

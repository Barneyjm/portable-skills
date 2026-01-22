#!/bin/bash
#
# Setup script for portable-skills development environment
#
# Usage:
#   ./scripts/setup-hooks.sh          # Install git hooks
#   ./scripts/setup-hooks.sh check    # Run all checks without installing
#   ./scripts/setup-hooks.sh status   # Show current hook status
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'

REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || {
    echo -e "${RED}Error: Not in a git repository${NC}"
    exit 1
}

cd "$REPO_ROOT"

# Print header
header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Install hooks
install_hooks() {
    header "Installing Git Hooks"

    # Configure git to use our hooks directory
    git config core.hooksPath .githooks

    echo -e "${GREEN}✓${NC} Git hooks installed successfully!"
    echo -e ""
    echo -e "The following hooks are now active:"
    echo -e "  ${BOLD}pre-commit${NC} - Runs before each commit to ensure:"
    echo -e "    • Go code is properly formatted (gofmt)"
    echo -e "    • No static analysis issues (go vet)"
    echo -e "    • Linting passes (golangci-lint)"
    echo -e "    • All unit tests pass"
    echo -e ""
    echo -e "${YELLOW}Note:${NC} To bypass hooks in emergencies: ${BOLD}git commit --no-verify${NC}"
    echo -e ""

    # Check for required tools
    check_tools
}

# Check for required development tools
check_tools() {
    header "Checking Required Tools"

    local missing=0

    # Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | cut -d' ' -f3)
        echo -e "${GREEN}✓${NC} Go installed: $GO_VERSION"
    else
        echo -e "${RED}✗${NC} Go not installed"
        echo -e "  Install from: https://go.dev/dl/"
        missing=1
    fi

    # golangci-lint
    if command -v golangci-lint &> /dev/null; then
        LINT_VERSION=$(golangci-lint --version 2>&1 | head -1)
        echo -e "${GREEN}✓${NC} golangci-lint installed: $LINT_VERSION"
    else
        echo -e "${YELLOW}⚠${NC} golangci-lint not installed (recommended)"
        echo -e "  Install with: ${BOLD}go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
    fi

    # gofmt (comes with Go)
    if command -v gofmt &> /dev/null; then
        echo -e "${GREEN}✓${NC} gofmt installed (bundled with Go)"
    fi

    echo ""
    if [ $missing -eq 1 ]; then
        echo -e "${YELLOW}Some tools are missing. Install them for the best experience.${NC}"
    else
        echo -e "${GREEN}All required tools are installed!${NC}"
    fi
}

# Show hook status
show_status() {
    header "Git Hooks Status"

    local hooks_path=$(git config --get core.hooksPath 2>/dev/null || echo ".git/hooks")

    echo -e "Hooks directory: ${BOLD}$hooks_path${NC}"
    echo ""

    if [ "$hooks_path" = ".githooks" ]; then
        echo -e "${GREEN}✓${NC} Custom hooks are active (.githooks directory)"
    else
        echo -e "${YELLOW}⚠${NC} Custom hooks not installed"
        echo -e "  Run: ${BOLD}./scripts/setup-hooks.sh${NC} to install"
    fi

    echo ""
    echo -e "Available hooks in .githooks/:"
    for hook in .githooks/*; do
        if [ -f "$hook" ]; then
            hook_name=$(basename "$hook")
            if [ -x "$hook" ]; then
                echo -e "  ${GREEN}✓${NC} $hook_name (executable)"
            else
                echo -e "  ${YELLOW}⚠${NC} $hook_name (not executable)"
            fi
        fi
    done
}

# Run all checks manually
run_checks() {
    header "Running All Pre-Commit Checks"

    echo -e "This runs the same checks as the pre-commit hook.\n"

    # Run the pre-commit hook directly
    if [ -x ".githooks/pre-commit" ]; then
        ./.githooks/pre-commit
    else
        echo -e "${RED}Error: Pre-commit hook not found or not executable${NC}"
        exit 1
    fi
}

# Main
case "${1:-install}" in
    install|"")
        install_hooks
        ;;
    check|run)
        run_checks
        ;;
    status)
        show_status
        ;;
    tools)
        check_tools
        ;;
    *)
        echo -e "${BOLD}Usage:${NC} $0 [command]"
        echo ""
        echo -e "${BOLD}Commands:${NC}"
        echo -e "  install  Install git hooks (default)"
        echo -e "  check    Run all pre-commit checks manually"
        echo -e "  status   Show current hook configuration"
        echo -e "  tools    Check for required development tools"
        exit 1
        ;;
esac

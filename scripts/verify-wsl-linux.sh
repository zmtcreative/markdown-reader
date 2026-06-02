#!/bin/bash

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Verify WSL/Linux build prerequisites and run repo validation steps"
    echo ""
    echo "Options:"
    echo "  --skip-go-tests         Skip 'go test ./...'"
    echo "  --skip-frontend         Skip frontend dependency install and build"
    echo "  --skip-wails-build      Skip 'wails build -tags webkit2_41'"
    echo "  -h, --help              Display this help message"
    echo ""
    echo "Examples:"
    echo "  $0"
    echo "  $0 --skip-wails-build"
}

skip_go_tests=0
skip_frontend=0
skip_wails_build=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-go-tests)
            skip_go_tests=1
            shift
            ;;
        --skip-frontend)
            skip_frontend=1
            shift
            ;;
        --skip-wails-build)
            skip_wails_build=1
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Error: Unknown option $1" >&2
            usage
            exit 1
            ;;
    esac
done

script_full_path="$(readlink -f "${BASH_SOURCE[0]}")"
script_root="$(dirname "$script_full_path")"

if [[ "$script_root" == */scripts ]]; then
    tmp_project_root="$(dirname "$script_root")"
else
    tmp_project_root="$script_root"
fi

if [[ -f "$tmp_project_root/wails.json" ]]; then
    project_root="$tmp_project_root"
else
    echo "Error: Could not find wails.json in the expected project root: $tmp_project_root" >&2
    exit 1
fi

frontend_dir="$project_root/frontend"
frontend_dist_dir="$frontend_dir/dist"

required_apt_packages=(
    build-essential
    pkg-config
    libgtk-3-dev
    libwebkit2gtk-4.1-dev
)

required_commands=(
    go
    npm
    wails
    pkg-config
)

print_section() {
    echo ""
    echo "== $1 =="
}

fail() {
    echo "Error: $1" >&2
    exit 1
}

check_command() {
    local command_name="$1"

    if ! command -v "$command_name" >/dev/null 2>&1; then
        fail "Required command not found: $command_name"
    fi
}

check_dpkg_package() {
    local package_name="$1"

    if ! dpkg -s "$package_name" >/dev/null 2>&1; then
        fail "Required package not installed: $package_name"
    fi
}

run_checked() {
    echo "> $*"
    "$@" || fail "Command failed: $*"
}

original_dir="$PWD"
cd "$project_root" || fail "Could not change to project root directory: $project_root"
trap 'cd "$original_dir" >/dev/null 2>&1 || true' EXIT

print_section "Environment"
uname -a

if grep -qi microsoft /proc/version 2>/dev/null; then
    echo "WSL environment detected"
else
    echo "Warning: This does not appear to be WSL. Continuing anyway."
fi

if [[ "$project_root" == /mnt/* ]]; then
    echo "Warning: Repo is running from a mounted Windows path: $project_root"
    echo "Warning: For better file watching and dev-server performance, prefer a Linux path such as ~/src/markdown-reader-dev"
fi

print_section "Required Commands"
for command_name in "${required_commands[@]}"; do
    check_command "$command_name"
    echo "Found command: $command_name -> $(command -v "$command_name")"
done

echo "Go version: $(go version)"
echo "Node version: $(node --version 2>/dev/null || echo 'node not found')"
echo "npm version: $(npm --version)"
echo "Wails version: $(wails version)"

print_section "Required Ubuntu Packages"
for package_name in "${required_apt_packages[@]}"; do
    check_dpkg_package "$package_name"
    echo "Installed package: $package_name"
done

print_section "pkg-config Checks"
echo "gtk+-3.0: $(pkg-config --modversion gtk+-3.0)"
echo "webkit2gtk-4.1: $(pkg-config --modversion webkit2gtk-4.1)"

if [[ -f "$frontend_dir/package-lock.json" ]]; then
    echo "frontend/package-lock.json found; npm ci will be used"
else
    echo "frontend/package-lock.json not found; npm install will be used"
fi

if [[ $skip_frontend -eq 0 ]]; then
    print_section "Frontend Install"
    cd "$frontend_dir" || fail "Could not change to frontend directory: $frontend_dir"

    if [[ -f package-lock.json ]]; then
        run_checked npm ci
    else
        run_checked npm install
    fi

    print_section "Frontend Build"
    run_checked npm run build
    cd "$project_root" || fail "Could not change back to project root directory: $project_root"
else
    print_section "Frontend Install/Build"
    echo "Skipped"

    if [[ ! -d "$frontend_dist_dir" ]]; then
        fail "Frontend build output not found at $frontend_dist_dir. Run without --skip-frontend or build frontend assets first."
    fi
fi

if [[ $skip_go_tests -eq 0 ]]; then
    print_section "Go Tests"
    run_checked go test ./...
else
    print_section "Go Tests"
    echo "Skipped"
fi

if [[ $skip_wails_build -eq 0 ]]; then
    print_section "Wails Build"
    run_checked wails build -tags webkit2_41
else
    print_section "Wails Build"
    echo "Skipped"
fi

print_section "Verification Complete"
echo "WSL/Linux verification completed successfully"
#!/usr/bin/env bash

set -uo pipefail

usage() {
    cat <<'EOF'
Usage: run-all-tests.sh [OPTIONS]

Run the Go and frontend test suites in WSL/Ubuntu.

Options:
  -q, --silent          Suppress command output and summary. Exit 0 only when all test suites pass.
  --run-all-tests       Run the full frontend Playwright suite instead of the default fast headless slice.
  --show-frontend-report
                        Open the Playwright HTML report after frontend tests complete.
  -h, --help            Display this help message.

Compatibility aliases:
  -Silent
  -RunAllTests
  -ShowFrontendReport

With no options, the script runs:
  - go test ./...
    - frontend fast Playwright tests in headless mode when Windows/WSL interop is available
EOF
}

SCRIPT_FULL_PATH="$(readlink -f "${BASH_SOURCE[0]}")"
SCRIPT_ROOT="$(dirname "$SCRIPT_FULL_PATH")"

if [[ "$SCRIPT_ROOT" == */scripts ]]; then
    TMP_PROJECT_ROOT="$(dirname "$SCRIPT_ROOT")"
else
    TMP_PROJECT_ROOT="$SCRIPT_ROOT"
fi

if [[ -f "$TMP_PROJECT_ROOT/wails.json" ]]; then
    PROJECT_ROOT="$TMP_PROJECT_ROOT"
else
    echo "Could not find wails.json in the expected project root: $TMP_PROJECT_ROOT" >&2
    exit 1
fi

FRONTEND_DIR="$PROJECT_ROOT/frontend"
FRONTEND_PS_SCRIPT="$SCRIPT_ROOT/Run-FrontendTests.ps1"
SILENT=0
RUN_ALL_TESTS=0
SHOW_FRONTEND_REPORT=0
FRONTEND_TESTS_SKIPPED=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        -q|--silent|-Silent)
            SILENT=1
            ;;
        --run-all-tests|-RunAllTests)
            RUN_ALL_TESTS=1
            ;;
        --show-frontend-report|-ShowFrontendReport)
            SHOW_FRONTEND_REPORT=1
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Error: Unknown option $1" >&2
            usage >&2
            exit 1
            ;;
    esac
    shift
done

write_status() {
    local message="$1"
    if [[ "$SILENT" -eq 0 ]]; then
        printf '%s\n' "$message"
    fi
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

format_duration() {
    local total_seconds="$1"
    local hours minutes seconds

    hours=$((total_seconds / 3600))
    minutes=$(((total_seconds % 3600) / 60))
    seconds=$((total_seconds % 60))

    printf '%02d:%02d:%02d' "$hours" "$minutes" "$seconds"
}

run_command() {
    "$@"
    return $?
}

run_go_tests() {
    if [[ "$SILENT" -eq 1 ]]; then
        run_command go test ./... >/dev/null 2>&1
    else
        run_command go test ./...
    fi
}

run_frontend_tests() {
    local selected_script requested_runtime_mode install_command previous_runtime_mode had_previous_runtime_mode test_exit_code
    local power_shell_exe frontend_ps_script_windows_path

    if [[ ! -d "$FRONTEND_DIR" ]]; then
        echo "Frontend directory not found: $FRONTEND_DIR" >&2
        return 1
    fi

    if [[ ! -f "$FRONTEND_PS_SCRIPT" ]]; then
        echo "Frontend test script not found: $FRONTEND_PS_SCRIPT" >&2
        return 1
    fi

    if [[ "$RUN_ALL_TESTS" -eq 1 ]]; then
        selected_script="all"
        requested_runtime_mode="auto"
    else
        selected_script="fast"
        requested_runtime_mode="headless"
    fi

    if command_exists pwsh.exe && command_exists wslpath; then
        power_shell_exe="pwsh.exe"
        frontend_ps_script_windows_path="$(wslpath -w "$FRONTEND_PS_SCRIPT")"

        if [[ "$SILENT" -eq 0 ]]; then
            write_status "Running frontend tests via Run-FrontendTests.ps1 through pwsh.exe..."
        fi

        if [[ "$SHOW_FRONTEND_REPORT" -eq 1 ]]; then
            if [[ "$SILENT" -eq 1 ]]; then
                CI=true run_command "$power_shell_exe" -NoProfile -File "$frontend_ps_script_windows_path" -TestSuite "$selected_script" -RuntimeMode "$requested_runtime_mode" -ShowReport >/dev/null 2>&1
            else
                CI=true run_command "$power_shell_exe" -NoProfile -File "$frontend_ps_script_windows_path" -TestSuite "$selected_script" -RuntimeMode "$requested_runtime_mode" -ShowReport
            fi
        else
            if [[ "$SILENT" -eq 1 ]]; then
                CI=true run_command "$power_shell_exe" -NoProfile -File "$frontend_ps_script_windows_path" -TestSuite "$selected_script" -RuntimeMode "$requested_runtime_mode" >/dev/null 2>&1
            else
                CI=true run_command "$power_shell_exe" -NoProfile -File "$frontend_ps_script_windows_path" -TestSuite "$selected_script" -RuntimeMode "$requested_runtime_mode"
            fi
        fi

        return $?
    fi

    FRONTEND_TESTS_SKIPPED=1
    if [[ "$SILENT" -eq 0 ]]; then
        write_status "Skipping frontend Playwright tests: Windows/WSL interop is not available."
        write_status "Development and frontend E2E testing are optimized for Windows and WSL/Ubuntu in this repository."
    fi
    return 0
}

invoke_test_command() {
    local name="$1"
    local command_name="$2"
    local started_at ended_at duration exit_code

    write_status "Running $name..."
    started_at=$(date +%s)

    "$command_name"
    exit_code=$?

    ended_at=$(date +%s)
    duration=$((ended_at - started_at))

    if [[ "$exit_code" -eq 0 ]]; then
        write_status "$name passed in $(format_duration "$duration")"
    else
        write_status "$name failed with exit code $exit_code after $(format_duration "$duration")"
    fi

    LAST_TEST_NAME="$name"
    LAST_TEST_EXIT_CODE="$exit_code"
    LAST_TEST_DURATION="$duration"
}

if ! command_exists go; then
    echo "The 'go' command is not available in PATH." >&2
    exit 1
fi

if [[ ! -d "$FRONTEND_DIR" ]]; then
    echo "Could not find frontend directory: $FRONTEND_DIR" >&2
    exit 1
fi

if [[ ! -f "$FRONTEND_PS_SCRIPT" ]]; then
    echo "Could not find frontend test script: $FRONTEND_PS_SCRIPT" >&2
    exit 1
fi

if command_exists pwsh.exe && command_exists wslpath; then
    :
elif [[ "$SILENT" -eq 0 ]]; then
    write_status "Windows/WSL interop is unavailable; frontend Playwright tests will be skipped."
fi

LAST_TEST_NAME=""
LAST_TEST_EXIT_CODE=1
LAST_TEST_DURATION=0
OVERALL_STARTED_AT=$(date +%s)

pushd "$PROJECT_ROOT" >/dev/null || exit 1

invoke_test_command "Go tests" run_go_tests
GO_NAME="$LAST_TEST_NAME"
GO_EXIT_CODE="$LAST_TEST_EXIT_CODE"
GO_DURATION="$LAST_TEST_DURATION"

invoke_test_command "Frontend tests" run_frontend_tests
FRONTEND_NAME="$LAST_TEST_NAME"
FRONTEND_EXIT_CODE="$LAST_TEST_EXIT_CODE"
FRONTEND_DURATION="$LAST_TEST_DURATION"

popd >/dev/null || true

ALL_PASSED=1
if [[ "$GO_EXIT_CODE" -ne 0 || "$FRONTEND_EXIT_CODE" -ne 0 ]]; then
    ALL_PASSED=0
fi

OVERALL_DURATION=$(( $(date +%s) - OVERALL_STARTED_AT ))

if [[ "$SILENT" -eq 0 ]]; then
    printf '\nTest summary\n'
    printf '============\n'

    if [[ "$GO_EXIT_CODE" -eq 0 ]]; then
        printf '[PASS] %s (exit %s, duration %s)\n' "$GO_NAME" "$GO_EXIT_CODE" "$(format_duration "$GO_DURATION")"
    else
        printf '[FAIL] %s (exit %s, duration %s)\n' "$GO_NAME" "$GO_EXIT_CODE" "$(format_duration "$GO_DURATION")"
    fi

    if [[ "$FRONTEND_TESTS_SKIPPED" -eq 1 ]]; then
        printf '[SKIP] %s (exit %s, duration %s)\n' "$FRONTEND_NAME" "$FRONTEND_EXIT_CODE" "$(format_duration "$FRONTEND_DURATION")"
    elif [[ "$FRONTEND_EXIT_CODE" -eq 0 ]]; then
        printf '[PASS] %s (exit %s, duration %s)\n' "$FRONTEND_NAME" "$FRONTEND_EXIT_CODE" "$(format_duration "$FRONTEND_DURATION")"
    else
        printf '[FAIL] %s (exit %s, duration %s)\n' "$FRONTEND_NAME" "$FRONTEND_EXIT_CODE" "$(format_duration "$FRONTEND_DURATION")"
    fi

    printf '\n'
    if [[ "$ALL_PASSED" -eq 1 ]]; then
        printf 'All tests passed (total duration %s)\n' "$(format_duration "$OVERALL_DURATION")"
    else
        printf 'One or more test suites failed (total duration %s)\n' "$(format_duration "$OVERALL_DURATION")"
    fi
fi

if [[ "$ALL_PASSED" -eq 1 ]]; then
    exit 0
fi

exit 1
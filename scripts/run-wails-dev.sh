#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 [OPTIONS] [FILE_PATH]"
    echo "Run Wails development server with optional file argument"
    echo ""
    echo "Options:"
    echo "  -f, --file FILE_PATH    Specify a markdown file to open"
    echo "  -h, --help              Display this help message"
    echo ""
    echo "Examples:"
    echo "  $0                              # Use default sample file"
    echo "  $0 /path/to/file.md            # Open specific file"
    echo "  $0 -f /path/to/file.md         # Open specific file (using flag)"
    echo "  $0 --file /path/to/file.md     # Open specific file (using long flag)"
}

# Initialize variables
file_path=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--file)
            file_path="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        -*)
            echo "Error: Unknown option $1" >&2
            usage
            exit 1
            ;;
        *)
            # Positional argument - treat as file path
            if [[ -z "$file_path" ]]; then
                file_path="$1"
            else
                echo "Error: Multiple file paths specified" >&2
                usage
                exit 1
            fi
            shift
            ;;
    esac
done

# Set up script and project paths
script_full_path="$(readlink -f "${BASH_SOURCE[0]}")"
script_root="$(dirname "$script_full_path")"
script_name="$(basename "$script_full_path")"

# Determine project root by looking for scripts directory
if [[ "$script_root" == */scripts ]]; then
    tmp_project_root="$(dirname "$script_root")"
else
    tmp_project_root="$script_root"
fi

# Verify we found the correct project root by checking for wails.json
if [[ -f "$tmp_project_root/wails.json" ]]; then
    project_root="$tmp_project_root"
else
    echo "Error: Could not find wails.json in the expected project root: $tmp_project_root" >&2
    exit 1
fi

# Function to run Wails dev
invoke_wails_dev() {
    # Save current directory and change to project root
    local original_dir="$PWD"
    cd "$project_root" || {
        echo "Error: Could not change to project root directory: $project_root" >&2
        exit 1
    }

    local sample_file="${project_root}/docs/sample.md"

    # If file path is provided and exists, use it instead of default
    if [[ -n "$file_path" ]]; then
        if [[ -f "$file_path" ]]; then
            sample_file="$file_path"
        else
            echo "Warning: Specified file does not exist: $file_path" >&2
            echo "Using default sample file instead: $sample_file" >&2
        fi
    fi

    # Run wails dev with the sample file as an app argument
    echo "Starting Wails dev server with file: $sample_file"
    wails dev -tags webkit2_41 -appargs "--file=\"${sample_file}\""

    # Alternative command (commented out, equivalent to PowerShell comment):
    # wails dev -loglevel Trace -appargs "--nohtml"

    # Restore original directory
    cd "$original_dir" || {
        echo "Warning: Could not return to original directory: $original_dir" >&2
    }
}

# Run the main function
invoke_wails_dev

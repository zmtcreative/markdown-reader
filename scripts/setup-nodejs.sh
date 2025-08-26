#!/bin/bash
# This script sets up Node.js and npm for the project

# Install curl
if command -v curl >/dev/null 2>&1; then
    echo "curl is already installed"
else
    sudo apt install curl
fi

# Install nvm
if ! command -v nvm >/dev/null 2>&1; then
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash
    . "$HOME/.nvm/nvm.sh"
else
    echo "nvm is already installed"
fi

if command -v nvm >/dev/null 2>&1; then
    nvm install --lts
    nvm install node
    nvm use --lts
fi

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

invoke_setup_nodejs() {
    # Save current directory and change to project root
    local original_dir="$PWD"
    cd "$project_root" || {
        echo "Error: Could not change to project root directory: $project_root" >&2
        exit 1
    }

    cd "frontend" || {
        echo "Error: Could not change to frontend directory: $project_root/frontend" >&2
        exit 1
    }

    # Run the Node.js setup commands
    nvm use --lts
    npm install

    # Restore original directory
    cd "$original_dir" || {
        echo "Warning: Could not return to original directory: $original_dir" >&2
    }
}

invoke_setup_nodejs

#!/bin/bash
# Convenience script to run uatiari from any directory

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Get the virtualenv path (try poetry first, fallback to .venv)
if command -v poetry &> /dev/null; then
    VENV_PATH=$(cd "$SCRIPT_DIR" && poetry env info --path 2>/dev/null)
fi

if [ -z "$VENV_PATH" ] && [ -d "$SCRIPT_DIR/.venv" ]; then
    VENV_PATH="$SCRIPT_DIR/.venv"
fi


if [ -z "$VENV_PATH" ] || [ ! -f "$VENV_PATH/bin/activate" ]; then
    echo "Error: Poetry environment not found. Please run 'poetry install' in $SCRIPT_DIR"
    exit 1
fi

# Activate virtual environment
source "$VENV_PATH/bin/activate"

# Run CLI using module path
# We preserve the current working directory so git operations work on the target repo
PYTHONPATH="$SCRIPT_DIR/src:$PYTHONPATH" python -m uatiari.cli "$@"

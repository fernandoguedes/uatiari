#!/bin/bash
# Convenience script to run uatiari from any directory

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Get the virtualenv path (running poetry from project root)
VENV_PATH=$(cd "$SCRIPT_DIR" && poetry env info --path 2>/dev/null)

if [ -z "$VENV_PATH" ] || [ ! -f "$VENV_PATH/bin/activate" ]; then
    echo "Error: Poetry environment not found. Please run 'poetry install' in $SCRIPT_DIR"
    exit 1
fi

# Activate virtual environment
source "$VENV_PATH/bin/activate"

# Run CLI using module path, adding src to PYTHONPATH
# We preserve the current working directory so git operations work on the target repo
PYTHONPATH="$SCRIPT_DIR/src:$PYTHONPATH" python -m uatiari.cli "$@"

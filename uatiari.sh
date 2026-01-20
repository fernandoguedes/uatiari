#!/bin/bash
# Convenience script to run uatiari from any directory

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Activate virtual environment and run CLI using absolute paths
# Keep current directory so git operations work
source "$SCRIPT_DIR/.venv/bin/activate"
PYTHONPATH="$SCRIPT_DIR:$PYTHONPATH" python "$SCRIPT_DIR/src/cli.py" "$@"

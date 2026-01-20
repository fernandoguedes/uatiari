"""Git operations for extracting diff and file information."""

import subprocess
from typing import Optional


class GitError(Exception):
    """Custom exception for git-related errors."""
    pass


def validate_branch_exists(branch: str) -> bool:
    """
    Check if a git branch exists locally.
    
    Args:
        branch: Name of the branch to check
        
    Returns:
        True if branch exists, False otherwise
        
    Raises:
        GitError: If not in a git repository
    """
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--verify", branch],
            capture_output=True,
            text=True,
            check=False
        )
        return result.returncode == 0
    except FileNotFoundError:
        raise GitError("git command not found. Please install git.")
    except Exception as e:
        raise GitError(f"Failed to validate branch: {e}")


def get_diff(branch: str, base: str = "main") -> str:
    """
    Get the git diff between two branches.
    
    Args:
        branch: The feature branch to compare
        base: The base branch to compare against (default: "main")
        
    Returns:
        The git diff output as a string
        
    Raises:
        GitError: If git operation fails or branches don't exist
    """
    # Validate we're in a git repository
    try:
        subprocess.run(
            ["git", "rev-parse", "--git-dir"],
            capture_output=True,
            check=True
        )
    except subprocess.CalledProcessError:
        raise GitError("Not in a git repository. Please run from within a git repo.")
    except FileNotFoundError:
        raise GitError("git command not found. Please install git.")
    
    # Validate both branches exist
    if not validate_branch_exists(base):
        raise GitError(f"Base branch '{base}' does not exist.")
    
    if not validate_branch_exists(branch):
        raise GitError(f"Branch '{branch}' does not exist.")
    
    # Get the diff
    try:
        result = subprocess.run(
            ["git", "diff", f"{base}...{branch}"],
            capture_output=True,
            text=True,
            check=True
        )
        
        if not result.stdout.strip():
            raise GitError(f"No differences found between '{base}' and '{branch}'.")
        
        return result.stdout
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to get diff: {e.stderr}")


def get_changed_files(branch: str, base: str = "main") -> list[str]:
    """
    Get list of files changed between two branches.
    
    Args:
        branch: The feature branch to compare
        base: The base branch to compare against (default: "main")
        
    Returns:
        List of file paths that were changed
        
    Raises:
        GitError: If git operation fails
    """
    # Validate we're in a git repository
    try:
        subprocess.run(
            ["git", "rev-parse", "--git-dir"],
            capture_output=True,
            check=True
        )
    except subprocess.CalledProcessError:
        raise GitError("Not in a git repository.")
    except FileNotFoundError:
        raise GitError("git command not found. Please install git.")
    
    # Validate both branches exist
    if not validate_branch_exists(base):
        raise GitError(f"Base branch '{base}' does not exist.")
    
    if not validate_branch_exists(branch):
        raise GitError(f"Branch '{branch}' does not exist.")
    
    # Get changed files
    try:
        result = subprocess.run(
            ["git", "diff", "--name-only", f"{base}...{branch}"],
            capture_output=True,
            text=True,
            check=True
        )
        
        files = [f.strip() for f in result.stdout.split("\n") if f.strip()]
        return files
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to get changed files: {e.stderr}")

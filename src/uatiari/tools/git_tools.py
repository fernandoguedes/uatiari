"""Git operations for extracting diff and file information."""
import subprocess


class GitError(Exception):
    """Custom exception for git-related errors."""

    pass


def _run_git_command(
    args: list[str], check: bool = True
) -> subprocess.CompletedProcess:
    """
    Helper to run git commands with consistent error handling.

    Args:
        args: Git command arguments (e.g., ["diff", "main...feature"])
        check: Whether to raise on non-zero exit code

    Returns:
        CompletedProcess object

    Raises:
        GitError: If git is not found or command fails
    """
    try:
        return subprocess.run(
            ["git"] + args, capture_output=True, text=True, check=check
        )
    except FileNotFoundError:
        raise GitError("git command not found. Please install git.")
    except subprocess.CalledProcessError as e:
        error_msg = (e.stderr or e.stdout or "").strip() or "Unknown error"
        raise GitError(f"Git command failed: {error_msg}")


def _check_git_repository() -> None:
    """
    Verify we're inside a git repository.

    Raises:
        GitError: If not in a git repository
    """
    try:
        _run_git_command(["rev-parse", "--git-dir"])
    except GitError:
        raise GitError("Not in a git repository. Please run from within a git repo.")


def validate_branch_exists(branch: str) -> bool:
    """
    Check if a git branch exists locally.

    Args:
        branch: Name of the branch to check

    Returns:
        True if branch exists, False otherwise

    Raises:
        GitError: If not in a git repository or git not found
    """
    try:
        result = _run_git_command(["rev-parse", "--verify", branch], check=False)
        return result.returncode == 0
    except GitError:
        raise


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
    # Validate environment
    _check_git_repository()

    # Validate both branches exist
    if not validate_branch_exists(base):
        raise GitError(f"Base branch '{base}' does not exist.")

    if not validate_branch_exists(branch):
        raise GitError(f"Branch '{branch}' does not exist.")

    # Get the diff
    result = _run_git_command(["diff", f"{base}...{branch}"])

    if not result.stdout.strip():
        raise GitError(f"No differences found between '{base}' and '{branch}'.")

    return result.stdout


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
    # Validate environment
    _check_git_repository()

    # Validate both branches exist
    if not validate_branch_exists(base):
        raise GitError(f"Base branch '{base}' does not exist.")

    if not validate_branch_exists(branch):
        raise GitError(f"Branch '{branch}' does not exist.")

    # Get changed files
    result = _run_git_command(["diff", "--name-only", f"{base}...{branch}"])

    files = [f.strip() for f in result.stdout.split("\n") if f.strip()]

    if not files:
        raise GitError(f"No files changed between '{base}' and '{branch}'.")

    return files


def get_diff_stats(branch: str, base: str = "main") -> dict[str, tuple[int, int]]:
    """
    Get statistics of added/deleted lines per file.

    Args:
        branch: The feature branch
        base: The base branch (default: "main")

    Returns:
        Dictionary mapping filename to (added_lines, deleted_lines)

    Raises:
        GitError: If git operation fails
    """
    # Validate environment
    _check_git_repository()

    # Validate both branches exist
    if not validate_branch_exists(base):
        raise GitError(f"Base branch '{base}' does not exist.")

    if not validate_branch_exists(branch):
        raise GitError(f"Branch '{branch}' does not exist.")

    # Get numstat
    result = _run_git_command(["diff", "--numstat", f"{base}...{branch}"])

    stats = {}
    for line in result.stdout.splitlines():
        if not line.strip():
            continue
        parts = line.split("\t")
        if len(parts) >= 3:
            try:
                added = int(parts[0]) if parts[0] != "-" else 0
                deleted = int(parts[1]) if parts[1] != "-" else 0
                filename = parts[2]
                stats[filename] = (added, deleted)
            except ValueError:
                continue

    return stats


def get_repository_root() -> str:
    """
    Get the root directory of the git repository.

    Returns:
        Absolute path to repository root

    Raises:
        GitError: If not in a git repository
    """
    _check_git_repository()
    result = _run_git_command(["rev-parse", "--show-toplevel"])
    return result.stdout.strip()


def get_current_branch() -> str:
    """
    Get the name of the current git branch.

    Returns:
        Current branch name

    Raises:
        GitError: If not in a git repository or in detached HEAD state
    """
    _check_git_repository()
    result = _run_git_command(["rev-parse", "--abbrev-ref", "HEAD"])
    branch = result.stdout.strip()

    if branch == "HEAD":
        raise GitError("Currently in detached HEAD state. Please checkout a branch.")

    return branch


def get_commit_count(branch: str, base: str = "main") -> int:
    """
    Get the number of commits in branch that are not in base.

    Args:
        branch: The feature branch
        base: The base branch (default: "main")

    Returns:
        Number of commits ahead

    Raises:
        GitError: If git operation fails
    """
    _check_git_repository()

    result = _run_git_command(["rev-list", "--count", f"{base}..{branch}"], check=False)

    if result.returncode != 0:
        return 0

    try:
        return int(result.stdout.strip())
    except ValueError:
        return 0


def list_repository_files() -> list[str]:
    """
    List all tracked files in the repository.

    Returns:
        List of file paths relative to repo root

    Raises:
        GitError: If git operation fails
    """
    _check_git_repository()
    result = _run_git_command(["ls-tree", "-r", "--name-only", "HEAD"])
    return [f.strip() for f in result.stdout.split("\n") if f.strip()]

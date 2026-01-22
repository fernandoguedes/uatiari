"""Tests for git statistics."""

from unittest.mock import patch, MagicMock
from uatiari.tools.git_tools import get_diff_stats


@patch("uatiari.tools.git_tools._run_git_command")
@patch("uatiari.tools.git_tools._check_git_repository")
@patch("uatiari.tools.git_tools.validate_branch_exists")
def test_get_diff_stats_parsing(mock_validate, mock_check, mock_run):
    """Test parsing of git diff --numstat output."""
    mock_validate.return_value = True

    # Mock numstat output
    # added \t deleted \t filename
    mock_output = "10\t5\tsrc/main.py\n-\t-\tbinary.file\n20\t0\ttests/test_main.py\n"

    mock_process = MagicMock()
    mock_process.stdout = mock_output
    mock_process.returncode = 0
    mock_run.return_value = mock_process

    stats = get_diff_stats("feature", "main")

    assert stats["src/main.py"] == (10, 5)
    assert stats["tests/test_main.py"] == (20, 0)
    # Binary files are included with 0 lines changed (git numstat returns '-' for them)
    assert stats["binary.file"] == (0, 0)


@patch("uatiari.tools.git_tools._run_git_command")
@patch("uatiari.tools.git_tools._check_git_repository")
@patch("uatiari.tools.git_tools.validate_branch_exists")
def test_get_diff_stats_empty(mock_validate, mock_check, mock_run):
    """Test empty stats."""
    mock_validate.return_value = True

    mock_process = MagicMock()
    mock_process.stdout = ""
    mock_run.return_value = mock_process

    stats = get_diff_stats("feature", "main")
    assert stats == {}

"""Tests for git_tools module."""

import pytest
from unittest.mock import patch, MagicMock
import subprocess

from src.tools.git_tools import (
    validate_branch_exists,
    get_diff,
    get_changed_files,
    GitError,
)


class TestValidateBranchExists:
    """Tests for validate_branch_exists function."""

    @patch("subprocess.run")
    def test_branch_exists(self, mock_run):
        """Test that function returns True when branch exists."""
        mock_run.return_value = MagicMock(returncode=0)

        result = validate_branch_exists("main")

        assert result is True
        mock_run.assert_called_once_with(
            ["git", "rev-parse", "--verify", "main"],
            capture_output=True,
            text=True,
            check=False,
        )

    @patch("subprocess.run")
    def test_branch_does_not_exist(self, mock_run):
        """Test that function returns False when branch doesn't exist."""
        mock_run.return_value = MagicMock(returncode=1)

        result = validate_branch_exists("nonexistent")

        assert result is False

    @patch("subprocess.run")
    def test_git_not_installed(self, mock_run):
        """Test that GitError is raised when git is not installed."""
        mock_run.side_effect = FileNotFoundError()

        with pytest.raises(GitError, match="git command not found"):
            validate_branch_exists("main")


class TestGetDiff:
    """Tests for get_diff function."""

    @patch("src.tools.git_tools.validate_branch_exists")
    @patch("subprocess.run")
    def test_successful_diff(self, mock_run, mock_validate):
        """Test successful diff retrieval."""
        mock_validate.return_value = True
        mock_run.side_effect = [
            MagicMock(returncode=0),  # git rev-parse check
            MagicMock(stdout="diff --git a/file.py b/file.py\n+new line", returncode=0),
        ]

        result = get_diff("feature", "main")

        assert "diff --git" in result
        assert "+new line" in result

    @patch("subprocess.run")
    def test_not_in_git_repo(self, mock_run):
        """Test error when not in a git repository."""
        mock_run.side_effect = subprocess.CalledProcessError(1, "git")

        with pytest.raises(GitError, match="Not in a git repository"):
            get_diff("feature", "main")

    @patch("src.tools.git_tools.validate_branch_exists")
    @patch("subprocess.run")
    def test_base_branch_not_found(self, mock_run, mock_validate):
        """Test error when base branch doesn't exist."""
        mock_run.return_value = MagicMock(returncode=0)  # git rev-parse succeeds
        mock_validate.side_effect = [False, True]  # base doesn't exist, feature does

        with pytest.raises(GitError, match="Base branch 'main' does not exist"):
            get_diff("feature", "main")

    @patch("src.tools.git_tools.validate_branch_exists")
    @patch("subprocess.run")
    def test_feature_branch_not_found(self, mock_run, mock_validate):
        """Test error when feature branch doesn't exist."""
        mock_run.return_value = MagicMock(returncode=0)
        mock_validate.side_effect = [True, False]  # base exists, feature doesn't

        with pytest.raises(GitError, match="Branch 'feature' does not exist"):
            get_diff("feature", "main")

    @patch("src.tools.git_tools.validate_branch_exists")
    @patch("subprocess.run")
    def test_no_differences(self, mock_run, mock_validate):
        """Test error when there are no differences between branches."""
        mock_validate.return_value = True
        mock_run.side_effect = [
            MagicMock(returncode=0),  # git rev-parse check
            MagicMock(stdout="", returncode=0),  # empty diff
        ]

        with pytest.raises(GitError, match="No differences found"):
            get_diff("feature", "main")


class TestGetChangedFiles:
    """Tests for get_changed_files function."""

    @patch("src.tools.git_tools.validate_branch_exists")
    @patch("subprocess.run")
    def test_successful_file_list(self, mock_run, mock_validate):
        """Test successful retrieval of changed files."""
        mock_validate.return_value = True
        mock_run.side_effect = [
            MagicMock(returncode=0),  # git rev-parse check
            MagicMock(stdout="file1.py\nfile2.py\nfile3.py\n", returncode=0),
        ]

        result = get_changed_files("feature", "main")

        assert result == ["file1.py", "file2.py", "file3.py"]

    @patch("src.tools.git_tools.validate_branch_exists")
    @patch("subprocess.run")
    def test_empty_file_list(self, mock_run, mock_validate):
        """Test error when no files changed."""
        mock_validate.return_value = True
        mock_run.side_effect = [
            MagicMock(returncode=0),
            MagicMock(stdout="", returncode=0),
        ]

        with pytest.raises(GitError, match="No files changed"):
            get_changed_files("feature", "main")

    @patch("subprocess.run")
    def test_not_in_git_repo(self, mock_run):
        """Test error when not in a git repository."""
        mock_run.side_effect = subprocess.CalledProcessError(1, "git")

        with pytest.raises(GitError, match="Not in a git repository"):
            get_changed_files("feature", "main")

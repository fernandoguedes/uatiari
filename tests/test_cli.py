"""Tests for CLI module."""

from unittest.mock import patch

import pytest
from uatiari.cli import parse_args


class TestParseArgs:
    """Tests for parse_args function."""

    @patch("sys.argv", ["uatiari", "feature-branch"])
    def test_basic_args(self):
        """Test parsing basic branch argument."""
        args = parse_args()
        assert args["branch_name"] == "feature-branch"
        assert args["base_branch"] == "main"
        assert args["skill"] is None

    @patch("sys.argv", ["uatiari", "feature-branch", "--base=develop"])
    def test_base_branch(self):
        """Test parsing base branch argument."""
        args = parse_args()
        assert args["branch_name"] == "feature-branch"
        assert args["base_branch"] == "develop"

    @patch("sys.argv", ["uatiari", "feature-branch", "--skill=laravel"])
    def test_skill_arg(self):
        """Test parsing skill argument."""
        args = parse_args()
        assert args["skill"] == "laravel"

    @patch("sys.argv", ["uatiari", "feature-branch", "--base=dev", "--skill=react"])
    def test_all_args(self):
        """Test parsing all arguments."""
        args = parse_args()
        assert args["branch_name"] == "feature-branch"
        assert args["base_branch"] == "dev"
        assert args["skill"] == "react"

    @patch("sys.argv", ["uatiari", "--help"])
    def test_help(self):
        """Test help flag exits."""
        with pytest.raises(SystemExit) as exc:
            parse_args()
        assert exc.value.code == 0

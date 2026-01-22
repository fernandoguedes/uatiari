"""Tests for configuration management."""

import os
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from uatiari.config import load_configurations

# Mock data
MOCK_LOCAL_ENV = "LOCAL_ENV_PATH/.env"
MOCK_USER_CONFIG_ENV = "USER_CONFIG_PATH/uatiari/.env"
MOCK_USER_HOME_ENV = "USER_HOME_PATH/.uatiari.env"
MOCK_API_KEY = "mock_api_key_123"


@pytest.fixture
def mock_paths():
    """Mock pathlib.Path to control file system interactions."""
    with patch("uatiari.config.Path") as mock_path:
        # Setup standard paths
        mock_cwd = MagicMock()
        mock_cwd.__truediv__.return_value = Path(MOCK_LOCAL_ENV)

        mock_home = MagicMock()
        mock_user_config_dir = MagicMock()
        mock_user_config_env = MagicMock()
        mock_user_home_env = MagicMock()

        # Chain: Path.home() / ".config" / "uatiari" / ".env"
        mock_path.home.return_value = mock_home
        mock_home.__truediv__.side_effect = lambda x: {
            ".config": mock_user_config_dir,
            ".uatiari.env": mock_user_home_env,
        }.get(x, MagicMock())

        mock_user_config_dir.__truediv__.side_effect = lambda x: {
            "uatiari": mock_user_config_env
        }.get(
            x, MagicMock()
        )  # This is actually home/.config/uatiari

        # Correcting the chain based on config.py structure:
        # local_env = Path.cwd() / ".env"
        # user_config_dir = Path.home() / ".config" / "uatiari"
        # user_config_env = user_config_dir / ".env"
        # user_home_env = Path.home() / ".uatiari.env"

        mock_local_path = MagicMock()
        mock_path.cwd.return_value = mock_local_path
        mock_local_path.__truediv__.return_value = MagicMock(spec=Path)  # local_env

        # Reset side effects to simple mocks for easier control in tests
        # We will patch the specific instances in the tests
        yield mock_path


@pytest.fixture
def mock_env_vars():
    """Clear environment variables."""
    with patch.dict(os.environ, {}, clear=True):
        yield


class TestLoadConfigurations:
    """Tests for load_configurations function priority logic."""

    @patch("uatiari.config.load_dotenv")
    @patch("uatiari.config.dotenv_values")
    @patch("uatiari.config.Path")
    def test_priority_1_local_env(
        self, mock_path, mock_dotenv_values, mock_load_dotenv, mock_env_vars
    ):
        """Test that local .env takes highest priority."""
        # Setup Paths
        mock_local_env = MagicMock()
        mock_local_env.exists.return_value = True

        # Path.cwd() / ".env"
        mock_path.cwd.return_value.__truediv__.return_value = mock_local_env

        # Mock dotenv values to simulate key presence
        mock_dotenv_values.side_effect = (
            lambda p: {"GOOGLE_API_KEY": "local_key"} if p == mock_local_env else {}
        )

        # Run
        source = load_configurations()

        # Assert
        assert str(source) == str(mock_local_env)
        # Verify load_dotenv called for local
        mock_load_dotenv.assert_any_call(mock_local_env, override=True)

    @patch("uatiari.config.load_dotenv")
    @patch("uatiari.config.dotenv_values")
    @patch("uatiari.config.Path")
    def test_priority_2_user_config(
        self, mock_path, mock_dotenv_values, mock_load_dotenv, mock_env_vars
    ):
        """Test that user config takes priority if local is missing."""
        # Setup Paths
        mock_local_env = MagicMock()
        mock_local_env.exists.return_value = False

        mock_user_config_env = MagicMock()
        mock_user_config_env.exists.return_value = True

        mock_path.cwd.return_value.__truediv__.return_value = mock_local_env
        # home / .config / uatiari / .env
        mock_path.home.return_value.__truediv__.return_value.__truediv__.return_value.__truediv__.return_value = (
            mock_user_config_env
        )

        mock_dotenv_values.side_effect = (
            lambda p: {"GOOGLE_API_KEY": "user_key"}
            if p == mock_user_config_env
            else {}
        )

        source = load_configurations()

        # In the actual code `user_config_dir` is defined as: Path.home() / ".config" / "uatiari"
        # We need to ensure we mock strictly what config.py does or rely on less strict mocking.
        # Verify source string contains expected path part or mock str
        assert source == str(mock_user_config_env)

    @patch("uatiari.config.load_dotenv")
    @patch("uatiari.config.dotenv_values")
    @patch("uatiari.config.Path")
    def test_directory_creation(self, mock_path, mock_dotenv_values, mock_load_dotenv):
        """Test that config directory is created."""
        mock_user_config_dir = MagicMock()

        # Path.home() / ".config" / "uatiari"
        mock_path.home.return_value.__truediv__.return_value.__truediv__.return_value = (
            mock_user_config_dir
        )

        load_configurations()

        mock_user_config_dir.mkdir.assert_called_once_with(parents=True, exist_ok=True)

    @patch("uatiari.config.load_dotenv")
    @patch("uatiari.config.dotenv_values")
    @patch("uatiari.config.Path")
    def test_directory_creation_error_handled(
        self, mock_path, mock_dotenv_values, mock_load_dotenv
    ):
        """Test that directory creation errors are swallowed."""
        mock_user_config_dir = MagicMock()
        mock_user_config_dir.mkdir.side_effect = PermissionError("Boom")

        mock_path.home.return_value.__truediv__.return_value.__truediv__.return_value = (
            mock_user_config_dir
        )

        # Should not raise exception
        load_configurations()

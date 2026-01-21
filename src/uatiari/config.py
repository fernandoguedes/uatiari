"""Configuration management for uatiari."""

import os
import sys
from pathlib import Path
from dotenv import load_dotenv, dotenv_values
from rich.console import Console

# Use stderr for config logs to avoid interfering with potential stdout output
console = Console(stderr=True)


def load_configurations() -> str | None:
    """
    Load configuration from .env files in priority order.

    Priority (Highest to Lowest):
    1. ./.env (Local)
    2. ~/.config/uatiari/.env (User Config)
    3. ~/.uatiari.env (User Home Legacy)
    4. Environment Variable (shell)

    Returns:
        Source of the API key, or None if not found.
    """
    local_env = Path.cwd() / ".env"
    user_config_dir = Path.home() / ".config" / "uatiari"
    user_config_env = user_config_dir / ".env"
    user_home_env = Path.home() / ".uatiari.env"

    # Ensure config directory exists
    try:
        user_config_dir.mkdir(parents=True, exist_ok=True)
    except Exception:
        # Ignore errors if we can't create directory
        pass

    # Determine source of API Key for logging purposes
    # We check in priority order to find the first one that defines the key
    api_key_source = None

    if local_env.exists() and "GOOGLE_API_KEY" in dotenv_values(local_env):
        api_key_source = str(local_env)
    elif user_config_env.exists() and "GOOGLE_API_KEY" in dotenv_values(user_config_env):
        api_key_source = str(user_config_env)
    elif user_home_env.exists() and "GOOGLE_API_KEY" in dotenv_values(user_home_env):
        api_key_source = str(user_home_env)
    elif os.getenv("GOOGLE_API_KEY"):
        api_key_source = "Environment Variable"

    # Load env files in REVERSE priority order using override=True
    # This ensures higher priority files overwrite lower priority ones
    # and all of them overwrite the shell environment variables (if defined in file).

    # 1. User Home Legacy (Lowest file priority)
    if user_home_env.exists():
        load_dotenv(user_home_env, override=True)

    # 2. User Config (Medium file priority)
    if user_config_env.exists():
        load_dotenv(user_config_env, override=True)

    # 3. Local (Highest file priority)
    if local_env.exists():
        load_dotenv(local_env, override=True)

    return api_key_source


# Execute configuration loading
_api_key_source = load_configurations()

# Google Gemini API key
GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY")

# Only validate API key if not running tests
if "pytest" not in sys.modules:
    if not GOOGLE_API_KEY:
        console.print("\n[bold red]❌ Error: GOOGLE_API_KEY not found.[/bold red]")
        console.print("Please configure your API key in one of the following locations:")
        console.print(f"  1. {Path.cwd() / '.env'} (Project specific)")
        console.print(f"  2. {Path.home() / '.config/uatiari/.env'} (Global)")
        console.print(f"  3. {Path.home() / '.uatiari.env'} (Global legacy)")
        console.print("  4. Environment Variable 'GOOGLE_API_KEY'")
        sys.exit(1)
    else:
        if _api_key_source:
             console.print(f"[dim]Using API key from: {_api_key_source}[/dim]")

# LLM configuration
LLM_MODEL = "models/gemini-2.5-flash"
LLM_TEMPERATURE = 0.3

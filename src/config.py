"""Configuration management for uatiari."""

import os
import sys
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Google Gemini API key
GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY")

# Only validate API key if not running tests
if not GOOGLE_API_KEY and "pytest" not in sys.modules:
    raise ValueError(
        "GOOGLE_API_KEY not found in environment. "
        "Please create a .env file with your API key. "
        "See .env.example for template."
    )

# LLM configuration
LLM_MODEL = "models/gemini-2.5-flash"  # Try without models/ prefix
LLM_TEMPERATURE = 0.3

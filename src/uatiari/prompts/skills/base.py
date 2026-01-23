"""Base class for review skills."""

from abc import ABC, abstractmethod
from typing import Any, Dict, List


class Skill(ABC):
    """Abstract base class for framework/language specific skills."""

    @property
    @abstractmethod
    def name(self) -> str:
        """The name of the skill (e.g., 'laravel')."""
        pass

    @abstractmethod
    def detect(self, file_paths: List[str], changed_files: List[str]) -> bool:
        """
        Detect if this skill should be applied.

        Args:
            file_paths: List of all files in the repository (or a representative set).
                       Used to detect framework structure (e.g., composer.json, artisan).
            changed_files: List of files changed in the current PR/diff.

        Returns:
            True if the skill is relevant, False otherwise.
        """
        pass

    @abstractmethod
    def get_prompt_addon(self) -> str:
        """
        Get the prompt text to append to the base XP prompt.

        Returns:
            String containing specific instructions for the LLM.
        """
        pass

    @abstractmethod
    def get_metadata(self) -> Dict[str, Any]:
        """
        Get metadata about the skill for the output report.

        Returns:
            Dictionary with skill info.
        """
        pass

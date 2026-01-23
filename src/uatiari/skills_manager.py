"""Skill manager for handling framework-specific skills."""

from typing import Any, Dict, List, Optional

from uatiari.logger import print_step
from uatiari.prompts.skills.base import Skill
from uatiari.prompts.skills.laravel import LaravelSkill
from uatiari.prompts.xp_reviewer import XP_SYSTEM_PROMPT


class SkillManager:
    """Manages detection and application of skills."""

    def __init__(self):
        self.available_skills: List[Skill] = [LaravelSkill()]
        self.active_skills: List[Skill] = []

    def detect_skills(
        self,
        manual_skill: Optional[str],
        repo_files: List[str],
        changed_files: List[str],
    ) -> List[Skill]:
        """
        Detect active skills based on manual override or automatic detection.

        Args:
            manual_skill: Name of skill forced by user, if any.
            repo_files: List of files in repository.
            changed_files: List of files changed in current diff.

        Returns:
            List of active skills.
        """
        self.active_skills = []

        for skill in self.available_skills:
            is_active = False

            # 1. Manual override
            if manual_skill and manual_skill.lower() == skill.name:
                is_active = True
                print_step(f"Skill '{skill.name}' activated manually", "info")

            # 2. Automatic detection (if no manual override specified)
            elif not manual_skill and skill.detect(repo_files, changed_files):
                is_active = True
                print_step(f"Skill '{skill.name}' detected automatically", "info")

            if is_active:
                self.active_skills.append(skill)

        return self.active_skills

    def get_system_prompt(self) -> str:
        """
        Compose the full system prompt including active skills.

        Returns:
            Complete system prompt string.
        """
        system_prompt = XP_SYSTEM_PROMPT

        prompt_addons = [skill.get_prompt_addon() for skill in self.active_skills]
        if prompt_addons:
            system_prompt += "\n\n" + "\n\n".join(prompt_addons)

        return system_prompt

    def get_metadata(self, manual_skill: Optional[str]) -> Optional[Dict[str, Any]]:
        """
        Generate metadata for the review result.

        Args:
            manual_skill: The manual skill name if provided.

        Returns:
            Metadata dictionary or None if no skills active.
        """
        if not self.active_skills:
            return None

        # For now, we take the name of the first skill as the "main" framework/skill
        # Future improvement: handle multiple skills more explicitly in output
        primary_skill = self.active_skills[0].name

        return {
            "framework_detected": primary_skill,
            "skills_applied": [s.name for s in self.active_skills],
            "detection_method": "manual" if manual_skill else "automatic",
            "skill_details": [s.get_metadata() for s in self.active_skills],
        }

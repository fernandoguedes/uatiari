"""Tests for SkillManager."""

from uatiari.prompts.skills.laravel import LaravelSkill
from uatiari.skills_manager import SkillManager


class TestSkillManager:
    """Tests for SkillManager class."""

    def test_initialization(self):
        """Test that manager initializes with available skills."""
        manager = SkillManager()
        assert len(manager.available_skills) > 0
        assert isinstance(manager.available_skills[0], LaravelSkill)

    def test_detect_skills_manual(self):
        """Test detection with manual override."""
        manager = SkillManager()
        skills = manager.detect_skills("laravel", [], [])

        assert len(skills) == 1
        assert skills[0].name == "laravel"
        assert len(manager.active_skills) == 1

    def test_detect_skills_automatic(self):
        """Test automatic detection."""
        manager = SkillManager()
        # Mock file structure for Laravel
        skills = manager.detect_skills(
            None, ["composer.json", "artisan"], ["app/User.php"]
        )

        assert len(skills) == 1
        assert skills[0].name == "laravel"

    def test_detect_skills_none(self):
        """Test no skills detected."""
        manager = SkillManager()
        skills = manager.detect_skills(None, ["package.json"], ["src/index.js"])

        assert len(skills) == 0

    def test_get_system_prompt_empty(self):
        """Test prompt generation with no active skills."""
        manager = SkillManager()
        prompt = manager.get_system_prompt()
        assert "XP" in prompt  # Should contain base prompt
        assert "LARAVEL" not in prompt

    def test_get_system_prompt_with_skill(self):
        """Test prompt generation with active skill."""
        manager = SkillManager()
        manager.detect_skills("laravel", [], [])
        prompt = manager.get_system_prompt()
        assert "XP" in prompt
        assert "LARAVEL SPECIFIC CHECKS" in prompt

    def test_get_metadata(self):
        """Test metadata generation."""
        manager = SkillManager()
        manager.detect_skills("laravel", [], [])

        metadata = manager.get_metadata("laravel")
        assert metadata["framework_detected"] == "laravel"
        assert metadata["detection_method"] == "manual"
        assert "performance" in metadata["skill_details"][0]["focus_areas"]

    def test_get_metadata_none(self):
        """Test metadata generation when no skills active."""
        manager = SkillManager()
        metadata = manager.get_metadata(None)
        assert metadata is None

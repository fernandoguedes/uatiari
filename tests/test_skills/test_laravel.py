from uatiari.prompts.skills.laravel import LaravelSkill


def test_laravel_detection_true():
    skill = LaravelSkill()
    # Mock file paths
    repo_files = ["app/Models/User.php", "composer.json", "artisan"]
    changed_files = ["app/Http/Controllers/UserController.php"]

    assert skill.detect(repo_files, changed_files) is True


def test_laravel_detection_false_no_php():
    skill = LaravelSkill()
    repo_files = ["app/Models/User.php", "composer.json"]
    changed_files = ["README.md"]  # No PHP changes

    assert skill.detect(repo_files, changed_files) is False


def test_laravel_detection_false_no_markers():
    skill = LaravelSkill()
    repo_files = ["src/main.py", "requirements.txt"]
    changed_files = ["src/utils.php"]  # PHP changes, but not a Laravel repo

    assert skill.detect(repo_files, changed_files) is False


def test_laravel_detection_dirs():
    skill = LaravelSkill()
    repo_files = ["app/User.php", "config/app.php"]
    changed_files = ["routes/web.php"]

    assert skill.detect(repo_files, changed_files) is True


def test_laravel_prompt_addon():
    skill = LaravelSkill()
    addon = skill.get_prompt_addon()
    assert "N+1 Queries" in addon
    assert "Mass Assignment" in addon
    assert "SQL Injection" in addon


def test_laravel_metadata():
    skill = LaravelSkill()
    metadata = skill.get_metadata()
    assert metadata["name"] == "laravel"
    assert "performance" in metadata["focus_areas"]

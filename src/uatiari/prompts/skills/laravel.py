"""Laravel skill for XP code reviewer."""

from typing import Any, Dict, List

from .base import Skill


class LaravelSkill(Skill):
    """Encapsulates Laravel-specific review knowledge."""

    @property
    def name(self) -> str:
        return "laravel"

    def detect(self, file_paths: List[str], changed_files: List[str]) -> bool:
        """
        Detect Laravel project.

        Criteria:
        1. Presence of typical Laravel files/dirs in file_paths (repository context).
        2. Presence of .php files in changed_files (relevance to current review).
        """
        # Check if we are reviewing PHP files
        has_php_changes = any(f.endswith(".php") for f in changed_files)
        if not has_php_changes:
            return False

        # Check for Laravel markers in the repository
        # file_paths should be a list of file paths relative to root
        # Simple check: do we see artisan or composer.json in the root?
        # Or do we see standard directories?

        # Optimize: check for exact root matches first
        if "artisan" in file_paths or "composer.json" in file_paths:
            return True

        # Check for directory existence (assuming file_paths contains directories or full paths)
        # If file_paths comes from something like `git ls-tree -r`, it might only have files.
        # We'll check if any file path starts with typical directories.

        marker_dirs = {"app/", "routes/", "config/", "database/"}
        for path in file_paths:
            for marker in marker_dirs:
                if path.startswith(marker):
                    return True

        return False

    def get_prompt_addon(self) -> str:
        """Get Laravel-specific XP instructions."""
        return """
## LARAVEL SPECIFIC CHECKS

In addition to standard XP rules, enforce these Laravel best practices:

### 1. PERFORMANCE
- **N+1 Queries**: Look for loops executing queries or missing `with()` in Eloquent chains.
- **Eager Loading**: Flag if relationships are accessed without preloading.
- **Indexing**: Suggest indexes for columns used in `where`, `orderBy`, or `join`.
- **Query Optimization**: Flag `count()` on collections (use `count()` on query builder) or loading entire collections just to check existence (use `exists()`).

### 2. MYSQL & DATABASE
- **Parameter Binding**: Ensure all raw queries use bindings.
- **Migrations**: Check for `down()` methods and appropriate column types.
- **Design**: Flag composite primary keys if simple ID suffices (unless specific need).

### 3. SECURITY
- **Mass Assignment**: Verify `$fillable` or `$guarded` usage on Models.
- **SQL Injection**: Flag any raw SQL interpolation (e.g., `DB::raw("id = $id")`).
- **XSS**: Check for `{{!! !!}}` (unescaped output) in Blade templates used with user input.
- **CSRF**: Ensure forms have `@csrf`.

### VERDICT ADJUSTMENTS
- **BLOCK**: SQL Injection risks, missing CSRF, critical N+1 in loops.
- **WARN**: Missing eager loading, unindexed queries.
"""

    def get_metadata(self) -> Dict[str, Any]:
        return {"name": "laravel", "focus_areas": ["performance", "mysql", "security"]}

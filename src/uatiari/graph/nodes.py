"""LangGraph workflow nodes for code review process."""

import json

from langchain_google_genai import ChatGoogleGenerativeAI

from uatiari.config import GOOGLE_API_KEY, LLM_MODEL, LLM_TEMPERATURE
from uatiari.graph.state import ReviewState
from uatiari.logger import (
    ask_approval,
    print_error,
    print_review_plan,
    print_review_result,
    print_step,
)
from uatiari.prompts.skills.laravel import LaravelSkill
from uatiari.prompts.xp_reviewer import PLAN_GENERATION_PROMPT, XP_SYSTEM_PROMPT
from uatiari.tools.git_tools import (
    GitError,
    get_changed_files,
    get_diff,
    get_diff_stats,
    list_repository_files,
)


def fetch_git_context(state: ReviewState) -> ReviewState:
    """
    Fetch git diff and changed files for the specified branch.

    Args:
        state: Current workflow state

    Returns:
        Updated state with diff_content and changed_files
    """
    print_step("Fetching git context...", "loading")

    try:
        branch = state["branch_name"]
        base = state["base_branch"]

        # Get diff and changed files
        diff_content = get_diff(branch, base)
        changed_files = get_changed_files(branch, base)
        diff_stats = get_diff_stats(branch, base)

        print_step(f"Found {len(changed_files)} changed file(s)", "success")

        return {
            **state,
            "diff_content": diff_content,
            "changed_files": changed_files,
            "diff_stats": diff_stats,
            "error": None,
        }
    except GitError as e:
        print_error(f"Git error: {e}")
        return {**state, "error": str(e)}
    except Exception as e:
        print_error(f"Unexpected error: {e}")
        return {**state, "error": f"Unexpected error: {e}"}


def generate_plan(state: ReviewState) -> ReviewState:
    """
    Generate a review plan using LLM analysis of the diff.

    Args:
        state: Current workflow state

    Returns:
        Updated state with review_plan
    """
    print_step("Generating review plan...", "loading")

    try:
        # Initialize LLM
        llm = ChatGoogleGenerativeAI(
            model=LLM_MODEL, temperature=LLM_TEMPERATURE, google_api_key=GOOGLE_API_KEY
        )

        # Prepare prompt
        changed_files_str = "\n".join(f"  - {f}" for f in state["changed_files"])
        diff_preview = state["diff_content"][:500]

        prompt = PLAN_GENERATION_PROMPT.format(
            changed_files=changed_files_str, diff_preview=diff_preview
        )

        # Get plan from LLM
        response = llm.invoke(prompt)
        plan = response.content

        return {**state, "review_plan": plan, "error": None}
    except Exception as e:
        print_error(f"Failed to generate plan: {e}")
        return {**state, "error": f"Failed to generate plan: {e}"}


def await_approval(state: ReviewState) -> ReviewState:
    """
    Display review plan and wait for user approval.

    Args:
        state: Current workflow state

    Returns:
        Updated state with user_approved boolean
    """
    print_review_plan(state["review_plan"])

    # Get user input with styled prompt
    approved = ask_approval()

    return {**state, "user_approved": approved}


def execute_review(state: ReviewState) -> ReviewState:
    """
    Execute XP-based code review using LLM.

    Args:
        state: Current workflow state

    Returns:
        Updated state with review_result
    """
    print_step("Executing XP review...", "loading")

    try:
        # Initialize LLM
        llm = ChatGoogleGenerativeAI(
            model=LLM_MODEL, temperature=LLM_TEMPERATURE, google_api_key=GOOGLE_API_KEY
        )

        # Skills detection
        available_skills = [LaravelSkill()]
        active_skills = []
        prompt_addons = []

        # Get repository files for detection
        try:
            repo_files = list_repository_files()
        except Exception:
            # If we can't list files (e.g. error), assume empty list
            # Detection will rely on changed files or manual override
            repo_files = []

        manual_framework = state.get("manual_framework")

        for skill in available_skills:
            is_active = False
            # 1. Manual override
            if manual_framework and manual_framework.lower() == skill.name:
                is_active = True
                print_step(f"Skill '{skill.name}' activated manually", "info")

            # 2. Automatic detection (if no manual override specified)
            elif not manual_framework and skill.detect(
                repo_files, state["changed_files"]
            ):
                is_active = True
                print_step(f"Skill '{skill.name}' detected automatically", "info")

            if is_active:
                active_skills.append(skill)
                prompt_addons.append(skill.get_prompt_addon())

        # Compose system prompt
        system_prompt = XP_SYSTEM_PROMPT
        if prompt_addons:
            system_prompt += "\n\n" + "\n\n".join(prompt_addons)

        # Create messages for the review
        messages = [
            ("system", system_prompt),
            ("human", f"Review this git diff:\n\n{state['diff_content']}"),
        ]

        # Get review from LLM
        response = llm.invoke(messages)

        # Parse JSON response
        try:
            # Try to extract JSON from response (in case it's wrapped in markdown)
            content = response.content.strip()

            # Remove markdown code blocks if present
            if content.startswith("```"):
                # Find the actual JSON content
                lines = content.split("\n")
                json_lines = []
                in_json = False
                for line in lines:
                    if line.startswith("```"):
                        in_json = not in_json
                        continue
                    if in_json or (not line.startswith("```")):
                        json_lines.append(line)
                content = "\n".join(json_lines).strip()

            review_result = json.loads(content)
        except json.JSONDecodeError:
            print_step("Warning: LLM returned invalid JSON", "warning")
            review_result = {
                "error": "Invalid JSON response from LLM",
                "raw_response": response.content,
            }

        # Metadata injection
        if active_skills:
            metadata = {
                "framework_detected": active_skills[0].name,  # Simplification
                "skills_applied": [s.name for s in active_skills],
                "detection_method": "manual" if manual_framework else "automatic",
                "skill_details": [s.get_metadata() for s in active_skills],
            }
            review_result["metadata"] = metadata

        # Inject deterministic test analysis
        try:
            diff_stats = state.get("diff_stats", {})
            prod_lines = 0
            test_lines = 0

            for filename, (added, deleted) in diff_stats.items():
                if "test" in filename.lower():
                    test_lines += added
                else:
                    prod_lines += added

            ratio = 0.0
            if prod_lines > 0:
                ratio = test_lines / prod_lines
            elif test_lines > 0:
                # Infinite ratio (tests only), cap or handle gracefully
                ratio = 100.0

            if ratio >= 1.0:
                verdict = "EXCELLENT"
            elif ratio >= 0.5:
                verdict = "GOOD"
            elif ratio > 0:
                verdict = "ACCEPTABLE"
            elif prod_lines == 0 and test_lines == 0:
                verdict = "N/A"
            else:
                verdict = "MISSING"

            # Merge with LLM result, overriding the numbers
            existing_analysis = review_result.get("test_analysis", {})
            if isinstance(existing_analysis, dict):
                # Keep notes/missing_tests_for from LLM if present
                existing_analysis["production_lines"] = prod_lines
                existing_analysis["test_lines"] = test_lines
                existing_analysis["ratio"] = ratio
                existing_analysis["verdict"] = verdict
                review_result["test_analysis"] = existing_analysis

        except Exception as e:
            # Don't fail the whole review if stats fail
            print_error(f"Failed to calculate test stats: {e}")

        return {
            **state,
            "review_result": review_result,
            "active_skills": [s.name for s in active_skills],
            "error": None,
        }
    except Exception as e:
        print_error(f"Failed to execute review: {e}")
        return {**state, "error": f"Failed to execute review: {e}"}


def generate_report(state: ReviewState) -> ReviewState:
    """
    Format and display the final review report.

    Args:
        state: Current workflow state

    Returns:
        State unchanged (terminal node)
    """
    print_review_result(state["review_result"])

    return state

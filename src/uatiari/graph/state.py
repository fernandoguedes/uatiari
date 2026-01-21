"""State definition for LangGraph workflow."""

from typing import TypedDict, Optional


class ReviewState(TypedDict):
    """
    State that flows through the LangGraph workflow.

    This state is passed between nodes and accumulates information
    throughout the review process.
    """

    # Input parameters
    branch_name: str
    base_branch: str

    # Git context
    diff_content: str
    changed_files: list[str]

    # Review workflow
    review_plan: str
    user_approved: bool

    # Results
    review_result: dict

    # Error handling
    error: Optional[str]

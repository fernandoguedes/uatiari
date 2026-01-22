"""Tests for LangGraph workflow."""

from unittest.mock import patch

from uatiari.graph.state import ReviewState
from uatiari.graph.workflow import should_continue, create_workflow


class TestShouldContinue:
    """Tests for conditional edge function."""

    def test_approved_no_error(self):
        """Test that approved state continues to execute_review."""
        state: ReviewState = {
            "branch_name": "feature",
            "base_branch": "main",
            "diff_content": "",
            "changed_files": [],
            "review_plan": "",
            "user_approved": True,
            "review_result": {},
            "error": None,
        }

        result = should_continue(state)
        assert result == "execute_review"

    def test_not_approved(self):
        """Test that rejected state ends workflow."""
        state: ReviewState = {
            "branch_name": "feature",
            "base_branch": "main",
            "diff_content": "",
            "changed_files": [],
            "review_plan": "",
            "user_approved": False,
            "review_result": {},
            "error": None,
        }

        result = should_continue(state)
        assert result == "end"

    def test_error_present(self):
        """Test that error state ends workflow."""
        state: ReviewState = {
            "branch_name": "feature",
            "base_branch": "main",
            "diff_content": "",
            "changed_files": [],
            "review_plan": "",
            "user_approved": True,
            "review_result": {},
            "error": "Some error occurred",
        }

        result = should_continue(state)
        assert result == "end"


class TestCreateWorkflow:
    """Tests for workflow creation."""

    def test_workflow_creation(self):
        """Test that workflow is created successfully."""
        workflow = create_workflow()

        # Verify workflow is compiled and ready
        assert workflow is not None

    @patch("uatiari.graph.nodes.ask_approval")
    @patch("uatiari.graph.nodes.get_diff")
    @patch("uatiari.graph.nodes.get_changed_files")
    def test_workflow_execution_with_error(self, mock_files, mock_diff, mock_approval):
        """Test workflow handles git errors gracefully."""
        mock_diff.side_effect = Exception("Git error")
        mock_approval.return_value = False  # User rejects

        workflow = create_workflow()

        initial_state: ReviewState = {
            "branch_name": "feature",
            "base_branch": "main",
            "diff_content": "",
            "changed_files": [],
            "review_plan": "",
            "user_approved": False,
            "review_result": {},
            "error": None,
        }

        result = workflow.invoke(initial_state)

        # Should have error set
        assert result["error"] is not None

"""LangGraph workflow definition for code review process."""

from typing import Literal
from langgraph.graph import StateGraph, END

from src.graph.state import ReviewState
from src.graph.nodes import (
    fetch_git_context,
    generate_plan,
    await_approval,
    execute_review,
    generate_report
)


def should_continue(state: ReviewState) -> Literal["execute_review", "end"]:
    """
    Conditional edge function to determine if review should proceed.
    
    Args:
        state: Current workflow state
        
    Returns:
        "execute_review" if approved, "end" if rejected or error occurred
    """
    # Check for errors first
    if state.get("error"):
        return "end"
    
    # Check user approval
    if state.get("user_approved", False):
        return "execute_review"
    else:
        return "end"


def create_workflow() -> StateGraph:
    """
    Create and compile the LangGraph workflow for code review.
    
    Returns:
        Compiled StateGraph ready for invocation
    """
    # Create the graph
    workflow = StateGraph(ReviewState)
    
    # Add nodes
    workflow.add_node("fetch_git_context", fetch_git_context)
    workflow.add_node("generate_plan", generate_plan)
    workflow.add_node("await_approval", await_approval)
    workflow.add_node("execute_review", execute_review)
    workflow.add_node("generate_report", generate_report)
    
    # Define edges
    workflow.set_entry_point("fetch_git_context")
    
    # Linear flow until approval
    workflow.add_edge("fetch_git_context", "generate_plan")
    workflow.add_edge("generate_plan", "await_approval")
    
    # Conditional edge based on approval
    workflow.add_conditional_edges(
        "await_approval",
        should_continue,
        {
            "execute_review": "execute_review",
            "end": END
        }
    )
    
    # Continue to report after review
    workflow.add_edge("execute_review", "generate_report")
    workflow.add_edge("generate_report", END)
    
    # Compile and return
    return workflow.compile()

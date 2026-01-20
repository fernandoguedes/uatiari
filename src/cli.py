"""CLI entry point for uatiari code review agent."""

import sys
from src.graph.workflow import create_workflow
from src.logger import print_header, print_error, console


def parse_args() -> tuple[str, str]:
    """
    Parse command line arguments.
    
    Returns:
        Tuple of (branch_name, base_branch)
    """
    if len(sys.argv) < 2:
        console.print("\n[bold red]Usage:[/bold red] uatiari <branch-name> [--base=main]")
        console.print("\n[bold]Example:[/bold]")
        console.print("  [cyan]uatiari feature/user-authentication[/cyan]")
        console.print("  [cyan]uatiari feature/new-api --base=develop[/cyan]\n")
        sys.exit(1)
    
    branch = sys.argv[1]
    base = "main"
    
    # Parse optional --base flag
    for arg in sys.argv[2:]:
        if arg.startswith("--base="):
            base = arg.split("=", 1)[1]
    
    return branch, base


def main():
    """Main entry point for the CLI application."""
    # Parse arguments
    branch, base = parse_args()
    
    # Print header
    print_header(branch, base)
    
    # Create workflow
    try:
        workflow = create_workflow()
    except Exception as e:
        print_error(f"Failed to initialize workflow: {e}")
        sys.exit(1)
    
    # Initialize state
    initial_state: dict = {
        "branch_name": branch,
        "base_branch": base,
        "diff_content": "",
        "changed_files": [],
        "review_plan": "",
        "user_approved": False,
        "review_result": {},
        "error": None
    }
    
    # Run workflow
    try:
        result = workflow.invoke(initial_state)
        
        # Check for errors
        if result.get("error"):
            print_error(result['error'])
            sys.exit(1)
        
        # Success
        console.print("\n[bold green]✓ Review completed successfully![/bold green]\n")
        sys.exit(0)
        
    except KeyboardInterrupt:
        console.print("\n\n[bold red]✗ Interrupted by user.[/bold red]\n")
        sys.exit(1)
    except Exception as e:
        print_error(f"Unexpected error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()

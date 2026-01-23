"""CLI entry point for uatiari code review agent."""
import sys

from uatiari.graph.workflow import create_workflow
from uatiari.logger import console, print_error, print_header
from uatiari.updater import update_cli
from uatiari.version import __version__


def print_help():
    """Display help message."""
    help_text = """
[bold cyan]╭──────────────────────────────────────────────────────────────╮[/bold cyan]
[bold cyan]│[/bold cyan]                                                              [bold cyan]│[/bold cyan]
[bold cyan]│[/bold cyan]  [bold white]🎯 uatiari - XP Code Reviewer[/bold white]                               [bold cyan]│[/bold cyan]
[bold cyan]│[/bold cyan]  [dim]"to guide" in Nheengatu[/dim]                                     [bold cyan]│[/bold cyan]
[bold cyan]│[/bold cyan]                                                              [bold cyan]│[/bold cyan]
[bold cyan]╰──────────────────────────────────────────────────────────────╯[/bold cyan]

[bold]USAGE:[/bold]
  [cyan]uatiari[/cyan] <branch-name> [options]

[bold]ARGUMENTS:[/bold]
[bold]ARGUMENTS:[/bold]
  [yellow]<branch-name>[/yellow]    Branch to review against base (required)

[bold]COMMANDS:[/bold]
  [green]update[/green]           Update uatiari to the latest version

[bold]OPTIONS:[/bold]
  [green]--base=<branch>[/green]   Base branch for comparison (default: main)
  [green]--version[/green]         Show version information
  [green]--help, -h[/green]        Show this help message

[bold]EXAMPLES:[/bold]
  [dim]# Review feature branch against main[/dim]
  [cyan]uatiari feature/user-authentication[/cyan]

  [dim]# Review against develop branch[/dim]
  [cyan]uatiari feature/new-api --base=develop[/cyan]

  [dim]# Get help[/dim]
  [cyan]uatiari --help[/cyan]

[bold]WORKFLOW:[/bold]
  1. 📊 Fetches git diff between branches
  2. 📋 Generates review plan (files, XP checks, time estimate)
  3. ✋ Asks for human approval
  4. 🚀 Executes XP-based code review
  5. ✅ Outputs structured JSON report

[bold]XP PRINCIPLES ENFORCED:[/bold]
  • [green]TDD[/green] - Production code needs tests
  • [green]Simple Design[/green] - Flags unnecessary complexity
  • [green]Refactoring[/green] - Suggests small, safe improvements
  • [green]YAGNI[/green] - Identifies premature optimization

[dim]For more info: https://github.com/your-repo/uatiari[/dim]
"""
    console.print(help_text)


def parse_args() -> tuple[str, str]:
    """
    Parse command line arguments.

    Returns:
        Tuple of (branch_name, base_branch)
    """
    # Check for help flag
    if len(sys.argv) < 2 or "--help" in sys.argv or "-h" in sys.argv:
        print_help()
        sys.exit(0)

    if "--version" in sys.argv:
        console.print(f"uatiari version {__version__}")
        sys.exit(0)

    # Check for commands
    if sys.argv[1] == "update":
        update_cli()
        sys.exit(0)

    branch = sys.argv[1]
    base = "main"

    # Parse optional --base flag
    for arg in sys.argv[2:]:
        if arg.startswith("--base="):
            base = arg.split("=", 1)[1]
        elif arg.startswith("--"):
            console.print(f"\n[bold red]Error:[/bold red] Unknown option '{arg}'")
            console.print("[dim]Use --help for usage information[/dim]\n")
            sys.exit(1)

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
        "error": None,
    }

    # Run workflow
    try:
        result = workflow.invoke(initial_state)

        # Check for errors
        if result.get("error"):
            print_error(result["error"])
            sys.exit(1)

        # Success
        console.print("\n[bold green]✓ Review completed successfully![/bold green]\n")
        sys.exit(0)

    except KeyboardInterrupt:
        console.print("\n\n[bold yellow]⚠️  Review cancelled by user.[/bold yellow]\n")
        sys.exit(1)
    except Exception as e:
        print_error(f"Unexpected error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()

"""Modern logging utilities using rich library."""

from rich.console import Console
from rich.panel import Panel
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.table import Table
from rich.markdown import Markdown
from rich import print as rprint
from rich.syntax import Syntax
import json

# Create console instance
console = Console()


def print_header(branch: str, base: str):
    """Print application header with branding."""
    console.print()
    console.print(
        Panel.fit(
            f"[bold cyan]uatiari[/bold cyan] - XP Code Reviewer\n"
            f"[dim]Branch:[/dim] [yellow]{branch}[/yellow] [dim]→[/dim] [green]{base}[/green]",
            border_style="cyan",
            padding=(1, 2)
        )
    )
    console.print()


def print_step(message: str, step_type: str = "info"):
    """
    Print a step message with appropriate styling.
    
    Args:
        message: The message to display
        step_type: Type of step (info, success, error, warning)
    """
    icons = {
        "info": "⚙️",
        "success": "✅",
        "error": "❌",
        "warning": "⚠️",
        "loading": "⏳"
    }
    
    colors = {
        "info": "cyan",
        "success": "green",
        "error": "red",
        "warning": "yellow",
        "loading": "blue"
    }
    
    icon = icons.get(step_type, "•")
    color = colors.get(step_type, "white")
    
    console.print(f"{icon} [{color}]{message}[/{color}]")


def print_review_plan(plan: str):
    """Print the review plan in a formatted panel."""
    console.print()
    console.print(
        Panel(
            Markdown(plan),
            title="[bold cyan]📋 Review Plan[/bold cyan]",
            border_style="cyan",
            title_align="left",
            padding=(1, 2)
        )
    )
    console.print()


def print_review_result(result: dict):
    """Print the review result with rich formatting."""
    console.print()
    console.print(
        Panel.fit(
            "[bold green]✅ Review Complete[/bold green]",
            border_style="green"
        )
    )
    console.print()
    
    # Overall verdict
    verdict = result.get("overall", {})
    verdict_text = verdict.get("verdict", "UNKNOWN")
    verdict_reason = verdict.get("reason", "No reason provided")
    
    verdict_colors = {
        "APPROVE": "green",
        "REQUEST_CHANGES": "yellow",
        "REJECT": "red"
    }
    verdict_color = verdict_colors.get(verdict_text, "white")
    
    console.print(
        Panel(
            f"[bold {verdict_color}]{verdict_text}[/bold {verdict_color}]\n"
            f"[dim]{verdict_reason}[/dim]",
            title="[bold]Overall Verdict[/bold]",
            border_style=verdict_color
        )
    )
    console.print()
    
    # Blocking issues
    blocking = result.get("blocking_issues", [])
    if blocking:
        table = Table(title="🚫 Blocking Issues", border_style="red", show_header=True)
        table.add_column("File", style="cyan")
        table.add_column("Lines", style="yellow")
        table.add_column("Issue", style="red")
        table.add_column("Action Required", style="green")
        
        for issue in blocking:
            table.add_row(
                issue.get("file", ""),
                issue.get("lines", ""),
                issue.get("issue", ""),
                issue.get("action", "")
            )
        
        console.print(table)
        console.print()
    
    # Warnings
    warnings = result.get("warnings", [])
    if warnings:
        table = Table(title="⚠️  Warnings", border_style="yellow", show_header=True)
        table.add_column("File", style="cyan")
        table.add_column("Lines", style="yellow")
        table.add_column("Issue", style="yellow")
        table.add_column("Suggestion", style="green")
        table.add_column("Effort", style="dim")
        
        for warning in warnings:
            table.add_row(
                warning.get("file", ""),
                warning.get("lines", ""),
                warning.get("issue", ""),
                warning.get("suggestion", ""),
                warning.get("effort", "")
            )
        
        console.print(table)
        console.print()
    
    # Suggestions
    suggestions = result.get("suggestions", [])
    if suggestions:
        table = Table(title="💡 Suggestions", border_style="blue", show_header=True)
        table.add_column("File", style="cyan")
        table.add_column("Lines", style="yellow")
        table.add_column("Improvement", style="blue")
        
        for suggestion in suggestions:
            table.add_row(
                suggestion.get("file", ""),
                suggestion.get("lines", ""),
                suggestion.get("improvement", "")
            )
        
        console.print(table)
        console.print()
    
    # Test analysis
    test_analysis = result.get("test_analysis", {})
    if test_analysis:
        prod_lines = test_analysis.get("production_lines", 0)
        test_lines = test_analysis.get("test_lines", 0)
        ratio = test_analysis.get("ratio", 0)
        verdict = test_analysis.get("verdict", "UNKNOWN")
        
        verdict_colors = {
            "EXCELLENT": "green",
            "GOOD": "green",
            "ACCEPTABLE": "yellow",
            "INSUFFICIENT": "red",
            "MISSING": "red"
        }
        verdict_color = verdict_colors.get(verdict, "white")
        
        console.print(
            Panel(
                f"[cyan]Production Lines:[/cyan] {prod_lines}\n"
                f"[cyan]Test Lines:[/cyan] {test_lines}\n"
                f"[cyan]Ratio:[/cyan] {ratio:.2f}\n"
                f"[bold {verdict_color}]Verdict: {verdict}[/bold {verdict_color}]",
                title="[bold]📊 Test Coverage Analysis[/bold]",
                border_style="cyan"
            )
        )
        console.print()


def print_error(message: str):
    """Print an error message."""
    console.print()
    console.print(
        Panel(
            f"[bold red]{message}[/bold red]",
            title="[bold red]❌ Error[/bold red]",
            border_style="red"
        )
    )
    console.print()


def print_json(data: dict):
    """Print JSON data with syntax highlighting."""
    json_str = json.dumps(data, indent=2)
    syntax = Syntax(json_str, "json", theme="monokai", line_numbers=False)
    console.print(syntax)


def ask_approval() -> bool:
    """
    Ask user for approval with styled prompt.
    
    Returns:
        True if approved, False otherwise
    """
    while True:
        response = console.input("\n[bold cyan]Approve execution?[/bold cyan] [dim](y/n)[/dim]: ").strip().lower()
        if response in ["y", "yes"]:
            console.print("[green]✓ Approved[/green]")
            return True
        elif response in ["n", "no"]:
            console.print("[red]✗ Cancelled[/red]")
            return False
        else:
            console.print("[yellow]Please enter 'y' or 'n'[/yellow]")

import json
import os
import platform
import shutil
import sys
import tempfile
import urllib.request
from pathlib import Path
from typing import Optional, Tuple

from uatiari.logger import console, print_error
from uatiari.version import GITHUB_REPO, __version__


def get_system_arch() -> Tuple[str, str]:
    """Get current system OS and architecture."""
    system = platform.system().lower()
    machine = platform.machine().lower()

    if system == "darwin":
        system = "macos"

    # Normalize arch
    if machine == "x86_64":
        machine = "x64"
    elif machine == "aarch64":
        machine = "arm64"

    return system, machine


def check_for_updates() -> Optional[dict]:
    """
    Check GitHub Releases for a newer version.
    Returns release dict if update available, None otherwise.
    """
    url = f"https://api.github.com/repos/{GITHUB_REPO}/releases/latest"
    try:
        req = urllib.request.Request(url)
        req.add_header("User-Agent", "uatiari-cli")

        with urllib.request.urlopen(req, timeout=5) as response:
            data = json.loads(response.read().decode())
            latest_tag = data.get("tag_name", "").lstrip("v")
            current_tag = __version__.lstrip("v")

            # Simple semantic version check
            # For robust comparsion use packaging.version but we want zero deps here if possible
            # Assuming strictly format x.y.z
            if latest_tag != current_tag:
                # TODO: improved version comparison
                return data

    except Exception:
        # Silently fail on connection errors during auto-check
        pass

    return None


def download_asset(download_url: str, target_path: Path):
    """Download asset with progress bar."""
    import rich.progress

    req = urllib.request.Request(download_url)
    req.add_header("User-Agent", "uatiari-cli")
    req.add_header("Accept", "application/octet-stream")

    with urllib.request.urlopen(req) as response:
        total_size = int(response.info().get("Content-Length", 0))

        with rich.progress.Progress(
            rich.progress.SpinnerColumn(),
            rich.progress.TextColumn("[progress.description]{task.description}"),
            rich.progress.BarColumn(),
            rich.progress.DownloadColumn(),
            console=console,
        ) as progress:
            task = progress.add_task("Downloading update...", total=total_size)

            with open(target_path, "wb") as f:
                while True:
                    chunk = response.read(8192)
                    if not chunk:
                        break
                    f.write(chunk)
                    progress.update(task, advance=len(chunk))


def update_cli():
    """Execute the update process."""
    console.print(f"Checking for updates (current: v{__version__})...")
    release = check_for_updates()

    if not release:
        console.print("[green]You are on the latest version.[/green]")
        return

    latest_version = release["tag_name"]
    console.print(f"[bold yellow]New version available: {latest_version}[/bold yellow]")

    if input("Do you want to update? [y/N] ").lower() != "y":
        console.print("Update cancelled.")
        return

    # Determine asset name
    system, arch = get_system_arch()
    asset_name_prefix = f"uatiari-{system}-{arch}"

    asset = None
    for a in release.get("assets", []):
        if a["name"].startswith(asset_name_prefix):
            asset = a
            break

    if not asset:
        console.print(f"[red]No compatible binary found for {system}-{arch}.[/red]")
        return

    # Download and install
    try:
        with tempfile.NamedTemporaryFile(delete=False) as tmp_file:
            download_path = Path(tmp_file.name)

        download_asset(asset["browser_download_url"], download_path)

        # Verify executable
        os.chmod(download_path, 0o755)

        # Replace current binary
        current_binary = Path(sys.executable)

        # Safety check: are we running frozen?
        if not getattr(sys, "frozen", False):
            console.print(
                "[yellow]Warning: Not running as a frozen binary. Update skipped.[/yellow]"
            )
            console.print(f"Downloaded binary to: {download_path}")
            return

        # Move new binary into place
        # On Unix we can overwrite running executable
        shutil.move(str(download_path), str(current_binary))

        console.print(
            f"\n[bold green]Successfully updated to {latest_version}![/bold green]"
        )
        console.print("Please run `uatiari --version` to verify.")

    except Exception as e:
        print_error(f"Update failed: {e}")
        if "download_path" in locals() and download_path.exists():
            os.unlink(download_path)

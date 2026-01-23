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
    asset_name_suffix = f"{system}-{arch}.tar.gz"

    asset = None
    for a in release.get("assets", []):
        if a["name"].endswith(asset_name_suffix):
            asset = a
            break

    if not asset:
        console.print(f"[red]No compatible update package found for {system}-{arch}.[/red]")
        return

    # Download and install
    download_path = None
    try:
        import tarfile

        with tempfile.NamedTemporaryFile(suffix=".tar.gz", delete=False) as tmp_file:
            download_path = Path(tmp_file.name)

        download_asset(asset["browser_download_url"], download_path)

        # Identify paths
        current_binary = Path(sys.executable)
        install_dir = current_binary.parent

        # Safety check: are we running frozen?
        if not getattr(sys, "frozen", False):
            console.print(
                "[yellow]Warning: Not running as a frozen binary. Update skipped.[/yellow]"
            )
            console.print(f"Downloaded update to: {download_path}")
            return

        with tempfile.TemporaryDirectory() as tmp_extract_dir:
            # Extract
            console.print("Extracting update...")
            with tarfile.open(download_path, "r:gz") as tar:
                # Security: use filter='data' if available (Python 3.12+), otherwise be careful
                if hasattr(tarfile.TarFile, 'extraction_filter'):
                    tar.extractall(path=tmp_extract_dir, filter='data')
                else:
                    tar.extractall(path=tmp_extract_dir)

            # The tar contains a folder named 'uatiari'
            extracted_folder = Path(tmp_extract_dir) / "uatiari"
            if not extracted_folder.exists():
                raise Exception("Update package structure is invalid (missing 'uatiari' folder)")

            # Prepare for swap
            backup_dir = install_dir.parent / f"{install_dir.name}.bak"
            if backup_dir.exists():
                shutil.rmtree(backup_dir)

            console.print("Installing update...")
            
            # Atomic-ish swap
            # 1. Rename current to backup
            shutil.move(str(install_dir), str(backup_dir))
            
            try:
                # 2. Move new to current
                shutil.move(str(extracted_folder), str(install_dir))
                
                # Restore execution permissions for the binary
                new_binary = install_dir / "uatiari"
                if new_binary.exists():
                    os.chmod(new_binary, 0o755)
                
                # 3. Cleanup backup
                shutil.rmtree(backup_dir)
                
            except Exception as e:
                # Rollback
                console.print(f"[red]Installation failed, rolling back... Error: {e}[/red]")
                if install_dir.exists():
                    shutil.rmtree(install_dir)
                shutil.move(str(backup_dir), str(install_dir))
                raise e

        console.print(
            f"\n[bold green]Successfully updated to {latest_version}![/bold green]"
        )
        console.print("Please run `uatiari --version` to verify.")

    except Exception as e:
        print_error(f"Update failed: {e}")
    finally:
        if download_path and download_path.exists():
            os.unlink(download_path)

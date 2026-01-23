#!/bin/bash
set -e

REPO="fernandoguedes/uatiari"
BINARY_NAME="uatiari"
INSTALL_ROOT="$HOME/.local/share"
INSTALL_DIR="$INSTALL_ROOT/uatiari"
BIN_DIR="$HOME/.local/bin"

# Detect OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$OS" == "darwin" ]; then
    OS="macos"
elif [ "$OS" == "linux" ]; then
    OS="linux"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

if [ "$ARCH" == "x86_64" ]; then
    ARCH="x64"
elif [ "$ARCH" == "arm64" ] || [ "$ARCH" == "aarch64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"

echo "Detected platform: $OS-$ARCH"
echo "Fetching latest version..."

# Get latest release tag
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")
TAG_NAME=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG_NAME" ]; then
    echo "Error: Could not find latest release."
    exit 1
fi

echo "Latest version: $TAG_NAME"

# Find asset URL
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url" | grep "$ASSET_NAME" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find package for $ASSET_NAME in release $TAG_NAME."
    exit 1
fi

# Download
TMP_FILE="/tmp/${ASSET_NAME}"
echo "Downloading $DOWNLOAD_URL..."
curl -L -o "$TMP_FILE" "$DOWNLOAD_URL"

# Install
echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_ROOT"
mkdir -p "$BIN_DIR"

if [ -d "$INSTALL_DIR" ]; then
    rm -rf "$INSTALL_DIR"
fi

tar -xzf "$TMP_FILE" -C "$INSTALL_ROOT"
rm "$TMP_FILE"

# Symlink
echo "Creating symlink in $BIN_DIR..."
ln -sf "$INSTALL_DIR/$BINARY_NAME" "$BIN_DIR/$BINARY_NAME"

# Check PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo "⚠️  Warning: $BIN_DIR is not in your PATH."
    echo "Add the following line to your shell configuration (.zshrc, .bashrc):"
    echo "export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo "Successfully installed $BINARY_NAME $TAG_NAME!"
echo "Run '$BINARY_NAME --help' to get started."

#!/bin/bash
set -e

REPO="fernandoguedes/uatiari"
BINARY_NAME="uatiari"
INSTALL_DIR="/usr/local/bin"

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

ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"

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
    echo "Error: Could not find binary for $ASSET_NAME in release $TAG_NAME."
    exit 1
fi

echo "Downloading $DOWNLOAD_URL..."
curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL"
chmod +x "$BINARY_NAME"

echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
fi

echo "Successfully installed $BINARY_NAME $TAG_NAME!"
echo "Run '$BINARY_NAME --help' to get started."

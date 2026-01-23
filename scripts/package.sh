#!/bin/bash
set -e

# Get system info
OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS_TYPE" == "darwin" ]; then
    OS_TYPE="macos"
fi

if [ "$ARCH" == "x86_64" ]; then
    ARCH="x64"
elif [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then
    ARCH="arm64"
fi

# Build
echo "Building uatiari..."
pyinstaller build.spec --noconfirm --clean

# Package
DIST_DIR="dist/uatiari"
PACKAGE_NAME="uatiari-${OS_TYPE}-${ARCH}.tar.gz"

echo "Packaging ${PACKAGE_NAME}..."
cd dist
tar -czf "../${PACKAGE_NAME}" uatiari/
cd ..

echo "✓ Package created at ${PACKAGE_NAME}"

#!/bin/bash
set -euo pipefail

OS_TYPE="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$OS_TYPE" = "darwin" ]; then
    OS_TYPE="macos"
elif [ "$OS_TYPE" != "linux" ]; then
    echo "Unsupported OS: $OS_TYPE"
    exit 1
fi

if [ "$ARCH" = "x86_64" ]; then
    ARCH="x64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

PACKAGE_NAME="uatiari-${OS_TYPE}-${ARCH}.tar.gz"

echo "Building uatiari..."
rm -rf dist
mkdir -p dist/uatiari
go build -trimpath -ldflags="-s -w" -o dist/uatiari/uatiari ./cmd/uatiari

echo "Packaging ${PACKAGE_NAME}..."
(
    cd dist
    tar -czf "../${PACKAGE_NAME}" uatiari
)

echo "Package created at ${PACKAGE_NAME}"

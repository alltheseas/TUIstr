#!/bin/bash
set -e

APP_NAME="communities-tui"
BUILD_DIR="build"
GO_MAIN_FILE="main.go"
INSTALL_DIR="/usr/local/bin"

# Build communities-tui
echo "Building communities-tui application..."
mkdir -p "$BUILD_DIR"
go build -o "$BUILD_DIR/$APP_NAME" "$GO_MAIN_FILE"

# Install communities-tui
echo "Installing communities-tui..."
echo "Copying binary to $INSTALL_DIR (may require sudo)..."
sudo install -m 0755 "$BUILD_DIR/$APP_NAME" "$INSTALL_DIR/$APP_NAME"

echo "Installation complete. You can now run $APP_NAME from your terminal."

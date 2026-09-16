#!/usr/bin/env bash
set -e

REPO="GabeCordo/DistributedFunctions"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin)               OS="mac" ;;
  linux)                OS="linux" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64)  ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

EXT=""
[ "$OS" = "windows" ] && EXT=".exe"

ASSET="fs-${OS}-${ARCH}${EXT}"

# Select destination directory based on environment and write permissions
if [ -n "$INSTALL_DIR" ]; then
  DEST_DIR="$INSTALL_DIR"
elif [ -w "/usr/local/bin" ]; then
  DEST_DIR="/usr/local/bin"
else
  DEST_DIR="$HOME/.local/bin"
fi

mkdir -p "$DEST_DIR"
DEST="${DEST_DIR}/fs${EXT}"

echo "Downloading ${ASSET}..."
curl -fsSL "https://github.com/${REPO}/releases/latest/download/${ASSET}" -o "$DEST"
chmod +x "$DEST"

echo "Successfully installed fs to ${DEST}"

if [[ ":$PATH:" != *":${DEST_DIR}:"* ]]; then
  echo "Warning: ${DEST_DIR} is not in your PATH."
fi

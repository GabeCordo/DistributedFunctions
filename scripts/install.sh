#!/usr/bin/env bash
set -e

REPO="GabeCordo/DistributedFunctions"

# Parse debug flag
DEBUG_INSTALL=false
while getopts ":d" opt; do
  case $opt in
    d)
      DEBUG_INSTALL=true
      ;;
  esac
done
shift $((OPTIND - 1))

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
if [ "$DEBUG_INSTALL" = true ]; then
  # Debug install always uses /usr/local/bin
  DEST_DIR="/usr/local/bin"
else
  if [ -n "$INSTALL_DIR" ]; then
    DEST_DIR="$INSTALL_DIR"
  elif [ -w "/usr/local/bin" ]; then
    DEST_DIR="/usr/local/bin"
  else
    DEST_DIR="$HOME/.local/bin"
  fi
fi

mkdir -p "$DEST_DIR"
DEST="${DEST_DIR}/fs${EXT}"

if [ "$DEBUG_INSTALL" = true ]; then
  # Debug install: build from source
  echo "Performing debug install..."
  
  # Get the directory where this script is located
  SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
  REPO_ROOT="$(dirname "$SCRIPT_DIR")"
  
  # Navigate to cmd/dfunc and build
  cd "$REPO_ROOT/cmd/dfunc"
  echo "Building from source in cmd/dfunc..."
  go build -o "fs${EXT}" .
  
  # Move the built binary to the destination
  mv "fs${EXT}" "$DEST"
  chmod +x "$DEST"
  
  echo "Successfully built and installed fs to ${DEST}"
else
  echo "Downloading ${ASSET}..."
  curl -fsSL "https://github.com/${REPO}/releases/latest/download/${ASSET}" -o "$DEST"
  chmod +x "$DEST"
  
  echo "Successfully installed fs to ${DEST}"
fi

if [[ ":$PATH:" != *":${DEST_DIR}:"* ]]; then
  echo "Warning: ${DEST_DIR} is not in your PATH."
fi

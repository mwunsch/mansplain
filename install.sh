#!/bin/sh
set -eu

REPO="mwunsch/mansplain"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) echo "error: unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect arch
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *) echo "error: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest version from GitHub API
if [ -n "${VERSION:-}" ]; then
  TAG="v$VERSION"
else
  TAG="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)"
fi

if [ -z "$TAG" ]; then
  echo "error: could not determine latest version" >&2
  exit 1
fi

VERSION="${TAG#v}"
NAME="mansplain_${VERSION}_${OS}_${ARCH}"
URL="https://github.com/$REPO/releases/download/$TAG/$NAME.tar.gz"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading mansplain $TAG for $OS/$ARCH..."
curl -fsSL "$URL" -o "$TMPDIR/$NAME.tar.gz"
tar xzf "$TMPDIR/$NAME.tar.gz" -C "$TMPDIR"

echo "Installing binary to $INSTALL_DIR..."
install -d "$INSTALL_DIR"
install -m 755 "$TMPDIR/$NAME/mansplain" "$INSTALL_DIR/mansplain"

MAN_BASE="${MAN_BASE:-$HOME/.local/share/man}"

echo "Installing man pages..."
for page in "$TMPDIR/$NAME"/man/*; do
  section="${page##*.}"
  dest="$MAN_BASE/man$section"
  install -d "$dest"
  install -m 644 "$page" "$dest/$(basename "$page")"
  echo "  $(basename "$page") -> $dest/"
done

echo "Done. Run 'mansplain --version' to verify."

# Check if INSTALL_DIR is in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Note: $INSTALL_DIR is not in your PATH. Add it with:" >&2
     echo "  export PATH=\"$INSTALL_DIR:\$PATH\"" >&2 ;;
esac

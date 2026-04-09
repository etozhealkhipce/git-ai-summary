#!/bin/sh
# Install git-ai-summary from GitHub Releases (diddo-hooks style).
# Usage: curl -sSL https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.sh | sh
# Pin version: GIT_AI_SUMMARY_VERSION=0.1.0 curl -sSL ... | sh

set -e

REPO="etozhealkhipce/git-ai-summary"
BASE_URL="https://github.com/${REPO}"
INSTALL_DIR="${GIT_AI_SUMMARY_INSTALL_DIR:-$HOME/.local/bin}"

detect_target() {
  OS=$(uname -s)
  ARCH=$(uname -m)
  case "$OS" in
  Darwin)
    case "$ARCH" in
    arm64) echo "aarch64-apple-darwin" ;;
    x86_64) echo "x86_64-apple-darwin" ;;
    *) echo "unsupported"; return 1 ;;
    esac ;;
  Linux)
    case "$ARCH" in
    x86_64) echo "x86_64-unknown-linux-gnu" ;;
    aarch64 | arm64) echo "aarch64-unknown-linux-gnu" ;;
    *) echo "unsupported"; return 1 ;;
    esac ;;
  *)
    echo "unsupported"
    return 1
    ;;
  esac
}

get_version() {
  if [ -n "$GIT_AI_SUMMARY_VERSION" ]; then
    echo "$GIT_AI_SUMMARY_VERSION"
    return
  fi
  tag=$(curl -sSL -o /dev/null -w '%{url_effective}' "${BASE_URL}/releases/latest" | sed -n 's|.*/tag/||p')
  if [ -z "$tag" ]; then
    echo "Could not determine latest release. Set GIT_AI_SUMMARY_VERSION=0.1.0 or create a release on GitHub." >&2
    return 1
  fi
  echo "${tag#v}"
}

TARGET=$(detect_target) || exit 1
if [ "$TARGET" = "unsupported" ]; then
  echo "Unsupported platform: $(uname -s) $(uname -m). macOS (Apple Silicon or Intel) and Linux (x86_64, aarch64) are supported." >&2
  exit 1
fi

VERSION=$(get_version) || exit 1
TARBALL="git-ai-summary-${VERSION}-${TARGET}.tar.gz"
URL="${BASE_URL}/releases/download/v${VERSION}/${TARBALL}"

mkdir -p "$INSTALL_DIR"
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT
curl -sSL -o "${tmpdir}/${TARBALL}" "$URL"
tar -xzf "${tmpdir}/${TARBALL}" -C "$tmpdir"
mv "$tmpdir/git-ai-summary" "${INSTALL_DIR}/git-ai-summary"
chmod +x "${INSTALL_DIR}/git-ai-summary"

echo "Installed git-ai-summary ${VERSION} to ${INSTALL_DIR}/git-ai-summary"
if ! command -v git-ai-summary >/dev/null 2>&1; then
  echo "Add ${INSTALL_DIR} to your PATH, for example:"
  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.profile"
  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc  # or ~/.bashrc"
fi

if [ -t 1 ]; then
  echo ""
  printf "Configure API keys now? Run: %s setup\n" "${INSTALL_DIR}/git-ai-summary"
fi

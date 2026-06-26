#!/bin/sh
set -e

REPO="Saku0512/specter"
BINARY="specter"
INSTALL_DIR="/usr/local/bin"

if [ -t 1 ]; then
  C_RESET="$(printf '\033[0m')"
  C_BOLD="$(printf '\033[1m')"
  C_CYAN="$(printf '\033[36m')"
  C_GREEN="$(printf '\033[32m')"
  C_YELLOW="$(printf '\033[33m')"
  C_RED="$(printf '\033[31m')"
else
  C_RESET=""
  C_BOLD=""
  C_CYAN=""
  C_GREEN=""
  C_YELLOW=""
  C_RED=""
fi

banner() {
  printf '%s\n' "${C_CYAN}${C_BOLD}        .-."
  printf '%s\n' "      (o o) boo"
  printf '%s\n' "      | O \\"
  printf '%s\n' "       \\   \\"
  printf '%s\n' "        \`~~~'${C_RESET}"
  printf '%s\n' "${C_BOLD}specter installer${C_RESET}"
  printf '\n'
}

info() {
  printf '%s==>%s %s\n' "${C_CYAN}${C_BOLD}" "${C_RESET}" "$1"
}

warn() {
  printf '%s==>%s %s\n' "${C_YELLOW}${C_BOLD}" "${C_RESET}" "$1"
}

success() {
  printf '%s==>%s %s\n' "${C_GREEN}${C_BOLD}" "${C_RESET}" "$1"
}

fail() {
  printf '%sError:%s %s\n' "${C_RED}${C_BOLD}" "${C_RESET}" "$1" >&2
  exit 1
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi

  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
    return
  fi

  fail "sha256sum or shasum is required to verify the download"
}

banner

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    fail "Unsupported OS: $OS"
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    fail "Unsupported architecture: $ARCH"
    ;;
esac

# Get latest release version
info "Fetching latest release metadata"
VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

if [ -z "$VERSION" ]; then
  fail "Failed to fetch latest version"
fi

info "Installing specter ${VERSION} for ${OS}/${ARCH}"

FILENAME="${BINARY}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/SHA256SUMS.txt"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
TMP="${TMP_DIR}/${FILENAME}"
CHECKSUMS="${TMP_DIR}/SHA256SUMS.txt"

info "Downloading ${FILENAME}"
curl -fsSL "$URL" -o "$TMP"

info "Verifying ${FILENAME}"
curl -fsSL "$CHECKSUM_URL" -o "$CHECKSUMS"
EXPECTED="$(grep "  ${FILENAME}$" "$CHECKSUMS" | awk '{print $1}')"
if [ -z "$EXPECTED" ]; then
  fail "Checksum for ${FILENAME} was not found in SHA256SUMS.txt"
fi

ACTUAL="$(sha256_file "$TMP")"
if [ "$ACTUAL" != "$EXPECTED" ]; then
  fail "Checksum mismatch for ${FILENAME}"
fi

chmod +x "$TMP"

# Install (use sudo if needed)
if [ -w "$INSTALL_DIR" ]; then
  info "Installing to ${INSTALL_DIR}/${BINARY}"
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  warn "Installing to ${INSTALL_DIR}/${BINARY} with sudo"
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

success "specter installed to ${INSTALL_DIR}/${BINARY}"
printf '%sRun:%s specter -c config.yml -p 8080\n' "${C_BOLD}" "${C_RESET}"

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

TMP="$(mktemp)"
info "Downloading ${FILENAME}"
curl -fsSL "$URL" -o "$TMP"
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

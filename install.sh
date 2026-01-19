#!/usr/bin/env bash

set -euo pipefail

# allow overriding the version
VERSION=${CODECRAFTERS_CLI_VERSION:-v46}

MUTED='\033[0;2m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

PLATFORM=$(uname -s)
ARCH=$(uname -m)

if [ "$PLATFORM" = "Darwin" ]; then
	OS=darwin
elif [ "${PLATFORM%% *}" = "Linux" ]; then
	OS=linux
else
	echo "This installer is only supported on Linux and MacOS."
	exit 1
fi

case "$ARCH" in
x86_64)
	ARCH=amd64
	;;
armv8* | arm64* | aarch64*)
	ARCH=arm64
	;;
*)
	echo "unsupported arch: $ARCH"
	exit 1
	;;
esac

INSTALL_DIR=${INSTALL_DIR:-/usr/local/bin}
INSTALL_PATH=${INSTALL_PATH:-$INSTALL_DIR/codecrafters}

DOWNLOAD_URL="https://github.com/codecrafters-io/cli/releases/download/${VERSION}/${VERSION}_${OS}_${ARCH}.tar.gz"

echo -e "Downloading ${GREEN}CodeCrafters CLI ${MUTED}(${VERSION})${NC}..."

if [ "$(id -u)" = "0" ]; then
	echo "Warning: this script is currently running as root. This is dangerous. "
	echo "         Instead run it as normal user. We will sudo as needed."
fi

if ! command -v curl >/dev/null; then
	echo "error: you do not have 'curl' installed which is required for this script."
	exit 1
fi

TEMP_FILE=$(mktemp "${TMPDIR:-/tmp}/.codecrafterscli.XXXXXXXX")
TEMP_FOLDER=$(mktemp -d "${TMPDIR:-/tmp}/.codecrafterscli-headers.XXXXXXXX")

cleanup() {
	echo -e "${NC}" # Ensure none of our colors leak
	rm -f "$TEMP_FILE"
	rm -rf "$TEMP_FOLDER"
}

trap cleanup EXIT

echo -e "${MUTED}" # Muted progress bar

HTTP_CODE=$(curl -SL --progress-bar "$DOWNLOAD_URL" --output "$TEMP_FILE" --write-out "%{http_code}")
if [ "$HTTP_CODE" -lt 200 ] || [ "$HTTP_CODE" -gt 299 ]; then
	echo -e "${NC}"
	echo "error: your platform and architecture (${PLATFORM}-${ARCH}) is unsupported."
	exit 1
fi

echo -e "${NC}"

tar xzf "$TEMP_FILE" -C "$TEMP_FOLDER" codecrafters

chmod 0755 "$TEMP_FOLDER/codecrafters"

if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
	echo -e "${MUTED}Note:${NC} You might need to enter your password to install."
	sudo mkdir -p "$INSTALL_DIR"
fi

if ! mv "$TEMP_FOLDER/codecrafters" "$INSTALL_PATH" 2>/dev/null; then
	echo -e "${MUTED}Note:${NC} You might need to enter your password to install."
	sudo mv "$TEMP_FOLDER/codecrafters" "$INSTALL_PATH"
fi

echo ""
echo -e "${GREEN}✔︎${NC} CodeCrafters CLI installed! ${MUTED}Version: $("$INSTALL_PATH" --version)${NC}"

#!/bin/sh

set -eu

# Ensure curl is installed before fetching the version
if ! command -v curl >/dev/null; then
  echo "error: 'curl' is required to fetch the latest version information."
  exit 1
fi

# Allow overriding the version
VERSION=${CODECRAFTERS_CLI_VERSION:-$(curl -s https://api.github.com/repos/codecrafters-io/cli/releases/latest \
    | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')}

# Fail early if we couldn’t extract a version
if [ -z "$VERSION" ]; then
  echo "error: failed to fetch the latest release tag from GitHub."
  exit 1
fi
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

if [ -f "$INSTALL_PATH" ]; then
  ACTION="update"
else
  ACTION="install"
fi

echo "This script will automatically $ACTION codecrafters (${VERSION}) for you."
echo "You will be prompted for your password by sudo if needed."
echo "Installation path: ${INSTALL_PATH}"

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
  rm -f "$TEMP_FILE"
  rm -rf "$TEMP_FOLDER"
}

trap cleanup EXIT

echo Downloading CodeCrafters CLI...

HTTP_CODE=$(curl -SL --progress-bar "$DOWNLOAD_URL" --output "$TEMP_FILE" --write-out "%{http_code}")
if [ "$HTTP_CODE" -lt 200 ] || [ "$HTTP_CODE" -gt 299 ]; then
  echo "error: your platform and architecture (${PLATFORM}-${ARCH}) is unsupported."
  exit 1
fi

tar xzf "$TEMP_FILE" -C "$TEMP_FOLDER" codecrafters

chmod 0755 "$TEMP_FOLDER/codecrafters"

if ! mv "$TEMP_FOLDER/codecrafters" "$INSTALL_PATH" 2>/dev/null; then
  sudo -k mv "$TEMP_FOLDER/codecrafters" "$INSTALL_PATH"
fi

echo "Installed $("$INSTALL_PATH" --version)"

echo 'Done!'

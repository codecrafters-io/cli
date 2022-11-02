#!/usr/bin/env bash
set -eu

# allow overriding the version
VERSION=${SENTRY_CLI_VERSION:-v17}

PLATFORM=`uname -s`
ARCH=`uname -m`

if [ "$PLATFORM" == "Darwin" ]; then
  OS=darwin
elif [ "$(expr substr "$PLATFORM" 1 5)" == "Linux" ]; then
  OS=linux
else
  echo "This installer is only supported on Linux and MacOS."
  exit 1
fi

if [ "$ARCH" == "x86_64" ]; then
  ARCH=amd64
elif [[ $ARCH == armv8* ]] || [[ $ARCH == arm64* ]] || [[ $ARCH == aarch64* ]]; then
  ARCH=arm64
else
  echo "unsupported arch: $ARCH"
  exit 1
fi


# If the install directory is not set, set it to a default
if [ -z ${INSTALL_DIR+x} ]; then
  INSTALL_DIR=/usr/local/bin
fi

if [ -z ${INSTALL_PATH+x} ]; then
  INSTALL_PATH="${INSTALL_DIR}/codecrafters"
fi

DOWNLOAD_URL="https://github.com/codecrafters-io/cli/releases/download/${VERSION}/${VERSION}_${OS}_${ARCH}.tar.gz"
echo $DOWNLOAD_URL

echo "This script will automatically install codecrafters (${VERSION}) for you."
echo "Installation path: ${INSTALL_PATH}"
if [ "x$(id -u)" == "x0" ]; then
  echo "Warning: this script is currently running as root. This is dangerous. "
  echo "         Instead run it as normal user. We will sudo as needed."
fi

if ! hash curl 2> /dev/null; then
  echo "error: you do not have 'curl' installed which is required for this script."
  exit 1
fi

TEMP_FILE=`mktemp "${TMPDIR:-/tmp}/.codecrafterscli.XXXXXXXX"`
TEMP_FOLDER=`mktemp -d "${TMPDIR:-/tmp}/.codecrafterscli-headers.XXXXXXXX"`

cleanup() {
  rm -f "$TEMP_FILE"
  rm -rf "$TEMP_FOLDER"
}

trap cleanup EXIT
HTTP_CODE=$(curl -SL --progress-bar "$DOWNLOAD_URL" --output "$TEMP_FILE" --write-out "%{http_code}")
if [[ ${HTTP_CODE} -lt 200 || ${HTTP_CODE} -gt 299 ]]; then
  echo "error: your platform and architecture (${PLATFORM}-${ARCH}) is unsupported."
  exit 1
fi

tar xzf "$TEMP_FILE" -C "$TEMP_FOLDER"

chmod 0755 "$TEMP_FOLDER/codecrafters"
if ! mv "$TEMP_FOLDER/codecrafters" "$INSTALL_PATH" 2> /dev/null; then
  sudo -k mv "$TEMP_FOLDER/codecrafters" "$INSTALL_PATH"
fi

echo "Installed $("$INSTALL_PATH" --version)"

echo 'Done!'
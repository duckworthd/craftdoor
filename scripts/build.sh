#!/bin/bash
#
# Build binary and prepare release/ directory.
#

set -e

TARGET="${1:-cmd/master/main}"
SRC=$(dirname "$TARGET")
BINARY=$(basename "$TARGET")
ASSETS_ROOT="assets/"
DST="release/"

# cd into project root.
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"/../
cd "${PROJECT_ROOT}"

# Clean the release/ directory.
echo "Cleaning output directory..."
rm -rf "${DST}"
mkdir -p "${DST}"

# Build for ARM.
echo "Building binary..."
env GOOS=linux \
    GOARCH=arm \
    GOARM=5 \
    CGO_ENABLED=1 \
    CC=arm-linux-gnueabi-gcc \
    go build -o "${DST}/${BINARY}" "${SRC}/${BINARY}.go"

# Copy auxiliary files.
echo "Copying auxiliary files..."
cp -r ${ASSETS_ROOT}/* ${DST}/

echo "Finished building '${SRC}'. Copy '${DST}' to your RPi and run it. For example,"
echo "$ rsync -r release/ raspberrypi:/home/pi/craftdoor"
echo "$ ssh raspberrypi 'cd /home/pi/craftdoor && ./${BINARY}'"
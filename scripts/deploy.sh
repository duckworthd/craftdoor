#!/bin/bash
#
# Copy a release to Raspberry Pi and run it. Exits on Ctrl+C.
#

set -eux

RELEASE_DIR="release/"
REMOTE_DIR="/home/pi/craftdoor"
HOSTNAME="raspberrypi"

# Copy contents of RELEASE_DIR to Raspberry Pi.
rsync -r --delete ${RELEASE_DIR} ${HOSTNAME}:${REMOTE_DIR}

# Launch main.go. Ensure that it exits on Ctrl+C.
ssh -t -t ${HOSTNAME} "cd ${REMOTE_DIR} && ./main develop.json"

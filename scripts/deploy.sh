#!/bin/bash
#
# Deploy a release.
#

set -eux

RELEASE_DIR="release/"
REMOTE_DIR="/home/pi/craftdoor"
HOSTNAME="pi@raspberrypi"

rsync -r ${RELEASE_DIR} ${HOSTNAME}:${REMOTE_DIR}
ssh ${HOSTNAME} "cd ${REMOTE_DIR} && ./main develop.json"

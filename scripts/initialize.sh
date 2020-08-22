#!/bin/bash
#
# Populates database with toy data.

set -eux

HOSTNAME="http://192.168.0.9:8080"

http POST ${HOSTNAME}/members name="John Lennon"
http POST ${HOSTNAME}/members name="Ringo Starr"
http POST ${HOSTNAME}/members name="Paul McCartney"
http POST ${HOSTNAME}/members name="George Harrison"

http POST ${HOSTNAME}/keys uuid="35c17053d7" member_id:=1
http POST ${HOSTNAME}/keys uuid="ffffffffff" member_id:=2

echo "Put a new key in front of the card reader now..."
http POST ${HOSTNAME}/keys/new member_id:=3

http GET ${HOSTNAME}/members
http GET ${HOSTNAME}/keys

http DELETE ${HOSTNAME}/keys/1
http GET ${HOSTNAME}/keys

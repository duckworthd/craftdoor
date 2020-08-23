#!/bin/bash
#
# Populates database with toy data.

set -eux

HOSTNAME="${1:-http://localhost:8080}"

http POST ${HOSTNAME}/members name="John Lennon"
http POST ${HOSTNAME}/members name="Ringo Starr"
http POST ${HOSTNAME}/members name="Paul McCartney"
http POST ${HOSTNAME}/members name="George Harrison"

http POST ${HOSTNAME}/keys uuid="35c17053d7" member_id:=4
http POST ${HOSTNAME}/keys uuid="ffffffffff" member_id:=2

http GET ${HOSTNAME}/keys
http PUT ${HOSTNAME}/keys/2 uuid="ffffffffff" member_id:=2
http DELETE ${HOSTNAME}/keys/2

http GET ${HOSTNAME}/members
http PUT ${HOSTNAME}/members/1 name="John Lennon, Jr."
http DELETE ${HOSTNAME}/members/1

echo "Put a new key in front of the card reader now..."
http POST ${HOSTNAME}/keys/new member_id:=1

#!/bin/bash
#
# Populates database with toy data.

set -eux

HOSTNAME="${1:-http://localhost:8080}"

http POST ${HOSTNAME}/api/members name="John Lennon"
http POST ${HOSTNAME}/api/members name="Ringo Starr"
http POST ${HOSTNAME}/api/members name="Paul McCartney"
http POST ${HOSTNAME}/api/members name="George Harrison"

http POST ${HOSTNAME}/api/keys uuid="35c17053d7" member_id:=4
http POST ${HOSTNAME}/api/keys uuid="ffffffffff" member_id:=2

http GET ${HOSTNAME}/api/keys
http PUT ${HOSTNAME}/api/keys/2 uuid="ffffffffff" member_id:=2
http DELETE ${HOSTNAME}/api/keys/2

http GET ${HOSTNAME}/api/members
http PUT ${HOSTNAME}/api/members/1 name="John Lennon, Jr."
http DELETE ${HOSTNAME}/api/members/1

echo "Put a new key in front of the card reader now..."
http POST ${HOSTNAME}/api/keys/new member_id:=1

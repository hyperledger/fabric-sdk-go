#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e
file_path=$1

DOCKER_CMD="${DOCKER_CMD:-docker}"
DOCKER_COMPOSE_CMD="${DOCKER_COMPOSE_CMD:-docker-compose}"

$DOCKER_COMPOSE_CMD $file_path ps -q | xargs $DOCKER_CMD inspect -f '{{ .Name }},{{ .State.ExitCode }}' | \

while read name ; do
if echo "$name" | grep -q "softhsm2"
then
    statusCode="${name: -1}"
    echo $statusCode
    if  [ "$statusCode" != "0" ]
    then
        exit $statusCode
    else
        exit 0
    fi
fi
done


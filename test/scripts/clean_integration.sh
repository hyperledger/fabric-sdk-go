#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

DOCKER_CMD="${DOCKER_CMD:-docker}"
DOCKER_COMPOSE_CMD="${DOCKER_COMPOSE_CMD:-docker-compose}"
FIXTURE_PROJECT_NAME="${FIXTURE_PROJECT_NAME:-fabsdkgo}"
DOCKER_REMOVE_FORCE="${DOCKER_REMOVE_FORCE:-false}"

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Removing docker-compose network created from fixtures ..."
COMPOSE_PROJECT_NAME=FIXTURE_PROJECT_NAME cd ${SCRIPT_PATH}/../fixtures/dockerenv && $DOCKER_COMPOSE_CMD -f docker-compose-chaincoded.yaml -f docker-compose.yaml -f docker-compose-nopkcs11-test.yaml -f docker-compose-pkcs11-test.yaml down

DOCKER_REMOVE_ARGS=
if [ "$DOCKER_REMOVE_FORCE" = "true" ]; then
    DOCKER_REMOVE_ARGS="-f"
fi

CONTAINERS=$($DOCKER_CMD ps -a | grep "${FIXTURE_PROJECT_NAME}-peer.\.org.\.example\.com-" | awk '{print $1}')
IMAGES=$($DOCKER_CMD images | grep "${FIXTURE_PROJECT_NAME}-peer.\.org.\.example\.com-" | awk '{print $1}')

if [ ! -z "$CONTAINERS" ]; then
    if [ "$DOCKER_REMOVE_FORCE" = "true" ]; then
        echo "Stopping chaincode containers created from fixtures ..."
        $DOCKER_CMD stop $CONTAINERS
    fi

    echo "Removing chaincode containers created from fixtures ..."
    $DOCKER_CMD rm $DOCKER_REMOVE_ARGS $CONTAINERS
fi

if [ ! -z "$IMAGES" ]; then
    echo "Removing chaincode images created from fixtures ..."
    $DOCKER_CMD rmi $DOCKER_REMOVE_ARGS $IMAGES
fi
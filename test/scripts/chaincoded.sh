#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

GO_CMD="${GO_CMD:-go}"
GOPATH="${GOPATH:-$HOME/go}"
CHAINCODE_PATH="${CHAINCODE_PATH:-${GOPATH}/src}"
CHAINCODED_PATH="${CHAINCODED_PATH:-${GOPATH}/src}"

# Temporary fix for Fabric base image
unset GOCACHE

echo "Installing chaincodes ..."
cd ${CHAINCODE_PATH}/github.com/example_cc
${GO_CMD} install github.com/example_cc
cd ${CHAINCODE_PATH}/github.com/example_pvt_cc
${GO_CMD} install github.com/example_pvt_cc
cd ${CHAINCODED_PATH}
${GO_CMD} install chaincoded/cmd/chaincoded

PEERS=(
    peer0.org1.example.com:7052
    peer1.org1.example.com:7152
    peer0.org2.example.com:8052
    peer1.org2.example.com:9052
)

# You can set CHAINCODED_VERBOSE environment variable to see additional chaincoded logs.
#export CHAINCODED_VERBOSE=true

echo "Running chaincoded ..."
chaincoded ":9375" ${PEERS[@]}
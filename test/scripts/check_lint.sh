#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script runs Go linting and vetting tools

set -e



GOMETALINT_CMD=gometalinter


PROJECT_PATH=$GOPATH/src/github.com/hyperledger/fabric-sdk-go

declare -a arr=(
"./pkg"
"./test"
)


echo "Running metalinters..."
for i in "${arr[@]}"
do
   echo "Checking $i"
   $GOMETALINT_CMD --config=./gometalinter.json $i/...
done
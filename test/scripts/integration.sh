#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Environment variables that affect this script:
# USE_PREBUILT_IMAGES: Integration tests are run against fabric docker images. 
#   Can be latest tagged (FALSE) or tags specified in .env (default).
# ARCH: Fabric docker images architecture.
#   If not set, ARCH defaults to the value specified in .env.
# GOTESTFLAGS: Flags are added to the go test command.
# LDFLAGS: Flags are added to the go test command (example: -ldflags=-s).

# Packages to include in test run
PKGS=`go list github.com/hyperledger/fabric-sdk-go/test/integration/... 2> /dev/null | \
                                                  grep -v /vendor/`

if [ "$USE_PREBUILT_IMAGES" == false ]
then
  echo "Setting docker integration fixture tags to latest..."
  source ./test/fixtures/latest-env.sh
fi

echo "Starting fabric and fabric-ca docker images..."
cd ./test/fixtures && docker-compose up --force-recreate -d

echo "Running integration tests..."
cd ../../
gocov test $GOTESTFLAGS $LDFLAGS $PKGS -p 1 -timeout=10m | gocov-xml > integration-report.xml

if [ $? -eq 0 ]
then
  echo "Integration tests passed. Cleaning up..."
  cd ./test/fixtures && docker-compose down
  exit 0
else
  echo "Integration tests failed. Cleaning up..."
  cd ./test/fixtures && docker-compose down
  exit 1
fi

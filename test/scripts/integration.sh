#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Packages to include in test run
PKGS=`go list github.com/hyperledger/fabric-sdk-go/test/integration/... 2> /dev/null | \
                                                  grep -v /vendor/`

# Detect Hyperledger CI environment
if [ "$JENKINS_URL" == "https://jenkins.hyperledger.org/" ] && [ "$USE_PREBUILT_IMAGES" == true ]
then
  echo "In Hyperledger CI - Setting docker integration fixture tags to latest and using pre-built images..."
  source ./test/fixtures/latest-env.sh
fi

echo "Starting fabric and fabric-ca docker images..."
cd ./test/fixtures && docker-compose up --force-recreate -d

echo "Running integration tests..."
cd ../../
gocov test $PKGS -p 1 -timeout=10m | gocov-xml > integration-report.xml

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

#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Packages to exclude
PKGS=`go list github.com/hyperledger/fabric-sdk-go/test/... 2> /dev/null | \
                                                  grep -v /vendor/`

# Detect Hyperledger CI environment
if [ "$JENKINS_URL" == "https://jenkins.hyperledger.org/" ]
then
  echo "In Hyperledger CI - Setting docker integration fixture tags to latest..."
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

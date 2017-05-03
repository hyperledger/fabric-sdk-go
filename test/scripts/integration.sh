#!/bin/bash -exu

# Packages to exclude
PKGS=`go list github.com/hyperledger/fabric-sdk-go/test/... 2> /dev/null | \
                                                  grep -v /vendor/`

echo "Starting fabric and fabric-ca docker images..."
cd ./test/fixtures && docker-compose up --force-recreate -d

echo "Running integration tests..."
cd ../../
gocov test -ldflags "$GO_LDFLAGS" $PKGS -p 1 -timeout=10m | gocov-xml > integration-report.xml

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

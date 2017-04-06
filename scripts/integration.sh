#!/bin/bash

# Packages to exclude
PKGS=`go list github.com/hyperledger/fabric-sdk-go/test/... 2> /dev/null | \
                                                  grep -v /vendor/`

echo "Starting fabric and fabric-ca docker images..."
cd ./test/fixtures && docker-compose up --force-recreate -d

sleep 1

echo "Running tests..."
gocov test -ldflags "$GO_LDFLAGS" $PKGS -p 1 -timeout=10m | gocov-xml > report.xml

docker-compose down

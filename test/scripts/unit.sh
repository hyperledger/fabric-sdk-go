#!/bin/bash

set -e

# Packages to exclude
PKGS=`go list github.com/hyperledger/fabric-sdk-go/... 2> /dev/null | \
                                                  grep -v /vendor/ | \
                                                  grep -v /test/`
echo "Running tests..."
gocov test -ldflags "$GO_LDFLAGS" $PKGS -p 1 -timeout=5m | gocov-xml > report.xml

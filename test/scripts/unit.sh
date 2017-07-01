#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# Environment variables that affect this script:
# GOTESTFLAGS: Flags are added to the go test command.
# LDFLAGS: Flags are added to the go test command (example: -ldflags=-s).

set -e

# Packages to exclude
PKGS=`go list github.com/hyperledger/fabric-sdk-go/... 2> /dev/null | \
                                                  grep -v /vendor/ | \
                                                  grep -v /test/`
echo "Running tests..."
gocov test $GOTESTFLAGS $LDFLAGS $PKGS -p 1 -timeout=5m | gocov-xml > report.xml

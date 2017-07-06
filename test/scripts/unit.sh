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

REPO="github.com/hyperledger/fabric-sdk-go"

# Packages to exclude
PKGS=`go list $REPO... 2> /dev/null | \
      grep -v ^$REPO/api/ | \
      grep -v ^$REPO/pkg/fabric-ca-client/mocks | grep -v ^$REPO/pkg/fabric-client/mocks | \
      grep -v ^$REPO/vendor/ | grep -v ^$REPO/test/`
echo "Running tests..."
gocov test $GOTESTFLAGS $LDFLAGS $PKGS -p 1 -timeout=5m | gocov-xml > report.xml

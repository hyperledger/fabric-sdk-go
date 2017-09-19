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
      grep -v ^$REPO/internal/github.com/ | grep -v ^$REPO/third_party/ | \
      grep -v ^$REPO/vendor/ | grep -v ^$REPO/test/`
echo "Running unit tests..."

RACEFLAG=""
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]
then
    RACEFLAG="-race"
fi
go test $RACEFLAG -cover $GOTESTFLAGS $LDFLAGS $PKGS -p 1 -timeout=40m
#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Packages to include in test run
PKGS=`go list github.com/hyperledger/fabric-sdk-go/test/integration/... 2> /dev/null | \
                                                  grep -v /vendor/`

echo "***Running integration tests..."
go test -cover $GOTESTFLAGS $LDFLAGS $PKGS -p 1 -timeout=40m


#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

GO_CMD="${GO_CMD:-go}"

# Packages to include in test run
PKGS=`$GO_CMD list github.com/hyperledger/fabric-sdk-go/test/integration/... 2> /dev/null | \
                                                  grep -v /vendor/`

echo "Running integration tests ..."
RACEFLAG=""
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]
then
    RACEFLAG="-race"
fi

if [ "$FABRIC_SDK_CLIENT_BCCSP_SECURITY_DEFAULT_PROVIDER" == "PKCS11" ]
then
    echo "Testing with PKCS11 ..."
    GO_TAGS="$GO_TAGS testpkcs11"
fi

$GO_CMD test $RACEFLAG -cover -tags "$GO_TAGS" $GO_TESTFLAGS $GO_LDFLAGS $PKGS -p 1 -timeout=40m

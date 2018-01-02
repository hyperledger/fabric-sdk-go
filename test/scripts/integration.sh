#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# Environment variables that affect this script:
# GO_TESTFLAGS: Flags are added to the go test command.
# GO_LDFLAGS: Flags are added to the go test command (example: -s).
# FABRIC_SDKGO_CODELEVEL_TAG: Go tag that represents the fabric code target
# FABRIC_SDKGO_CODELEVEL_VER: Version that represents the fabric code target (primarily for fixture lookup)
# FABRIC_CRYPTOCONFIG_VERSION: Version of cryptoconfig fixture to use

set -e

GO_CMD="${GO_CMD:-go}"
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-stable}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"
# TODO: better default handling for FABRIC_CRYPTOCONFIG_VERSION

# Packages to include in test run
PKGS=`$GO_CMD list github.com/hyperledger/fabric-sdk-go/test/integration/... 2> /dev/null | \
                                                  grep -v /vendor/`

echo "Running integration tests ..."
RACEFLAG=""
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then
    RACEFLAG="-race"
fi

echo "Testing with code level $FABRIC_SDKGO_CODELEVEL_TAG (Fabric ${FABRIC_SDKGO_CODELEVEL_VER}) ..."
GO_TAGS="$GO_TAGS $FABRIC_SDKGO_CODELEVEL_TAG"

if [ "$FABRIC_SDK_CLIENT_BCCSP_SECURITY_DEFAULT_PROVIDER" == "PKCS11" ]; then
    echo "Testing with PKCS11 ..."
    GO_TAGS="$GO_TAGS testpkcs11 pkcs11"
fi

GO_LDFLAGS="$GO_LDFLAGS -X github.com/hyperledger/fabric-sdk-go/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_SDKGO_CODELEVEL_VER}/channel -X github.com/hyperledger/fabric-sdk-go/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
$GO_CMD test $RACEFLAG -cover -tags "$GO_TAGS" $GO_TESTFLAGS -ldflags="$GO_LDFLAGS" $PKGS -p 1 -timeout=40m

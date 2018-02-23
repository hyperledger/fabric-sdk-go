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
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-devstable}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"

REPO="github.com/hyperledger/fabric-sdk-go"

PKGS="$PKGS $REPO/pkg/core/cryptosuite/bccsp/pkcs11 $REPO/pkg/core/cryptosuite/bccsp/multisuite"
echo "Running PKCS11 unit tests (libltdl and softhsm required)..."

RACEFLAG=""
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then
    RACEFLAG="-race"
fi

# detect softhsm
# created using command: softhsm2-util --init-token --slot 0 --label "ForFabric" --so-pin 1234 --pin 98765432
SOFTHSM=`softhsm2-util --show-slots 2> /dev/null | grep ForFabric` || SOFTHSM=""
if [ "$SOFTHSM" == "" ]; then
    echo "SoftHSM with ForFabric token not detected ..."
    exit 1
fi

echo "Testing with code level $FABRIC_SDKGO_CODELEVEL_TAG (Fabric ${FABRIC_SDKGO_CODELEVEL_VER}) ..."
GO_TAGS="$GO_TAGS $FABRIC_SDKGO_CODELEVEL_TAG"

GO_LDFLAGS="$GO_LDFLAGS -X github.com/hyperledger/fabric-sdk-go/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_SDKGO_CODELEVEL_VER}/channel -X github.com/hyperledger/fabric-sdk-go/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
$GO_CMD test $RACEFLAG -cover -tags "testing $GO_TAGS" $GO_TESTFLAGS -ldflags="$GO_LDFLAGS" $PKGS -p 1 -timeout=40m
#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# Environment variables that affect this script:
# GO_TESTFLAGS: Flags are added to the go test command.
# GO_LDFLAGS: Flags are added to the go test command (example: -s).
# TEST_CHANGED_ONLY: Boolean on whether to only run tests on changed packages.
# TEST_RACE_CONDITIONS: Boolean on whether to test for race conditions.
# TEST_WITH_LINTER: Boolean on whether to run linter prior to unit tests.
# FABRIC_SDKGO_CODELEVEL_TAG: Go tag that represents the fabric code target
# FABRIC_SDKGO_CODELEVEL_VER: Version that represents the fabric code target (primarily for fixture lookup)
# FABRIC_SDKGO_TESTRUN_ID: An identifier for the current run of tests.
# FABRIC_CRYPTOCONFIG_VERSION: Version of cryptoconfig fixture to use

set -e

GO_CMD="${GO_CMD:-go}"
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-devstable}"
FABRIC_SDKGO_TESTRUN_ID="${FABRIC_SDKGO_TESTRUN_ID:-${RANDOM}}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"
TEST_CHANGED_ONLY="${TEST_CHANGED_ONLY:-false}"
TEST_RACE_CONDITIONS="${TEST_RACE_CONDITIONS:-true}"
TEST_WITH_LINTER="${TEST_WITH_LINTER:-false}"
SCRIPT_DIR="$(dirname "$0")"

REPO="github.com/hyperledger/fabric-sdk-go"

source ${SCRIPT_DIR}/lib/find_packages.sh
source ${SCRIPT_DIR}/lib/linter.sh

echo "Running" $(basename "$0")

declare -a PKGS=(
    "${REPO}/pkg/core/cryptosuite/bccsp/pkcs11"
    "${REPO}/pkg/core/cryptosuite/bccsp/multisuite"
    "${REPO}/pkg/core/cryptosuite/common/pkcs11"
)

# Reduce unit tests to changed packages.
if [ "$TEST_CHANGED_ONLY" = true ]; then
    # findChangedFiles assumes that the working directory contains the repo; so change to the repo directory.
    PWD_ORIG=$(pwd)
    cd "${GOPATH}/src/${REPO}"
    findChangedFiles
    cd ${PWD_ORIG}

    if [[ "${CHANGED_FILES[@]}" =~ ( |^)(test/fixtures/|test/metadata/|test/scripts/|Makefile( |$)|Gopkg.lock( |$)|gometalinter.json( |$)|ci.properties( |$)) ]]; then
        echo "Test scripts, fixtures or metadata changed - running all tests"
    else
        findChangedPackages
        filterExcludedPackages
        appendDepPackages
        PKGS=(${DEP_PKGS[@]})
    fi
fi

RACEFLAG=""
if [ "${TEST_RACE_CONDITIONS}" = true ]; then
    ARCH=$(uname -m)

    if [ "${ARCH}" = "x86_64" ]; then
        echo "Enabling data race detection"
        RACEFLAG="-race"
    else
        echo "Data race detection not supported on ${ARCH}"
    fi
fi

if [ ${#PKGS[@]} -eq 0 ]; then
    echo "Skipping tests since no packages were changed"
    exit 0
fi

if [ "${TEST_WITH_LINTER}" = true ]; then
    runLinter
fi

echo "Code level ${FABRIC_SDKGO_CODELEVEL_TAG} (Fabric ${FABRIC_SDKGO_CODELEVEL_VER})"
echo "Running PKCS11 unit tests (libltdl and softhsm required)..."

# detect softhsm
# created using command: softhsm2-util --init-token --slot 0 --label "ForFabric" --so-pin 1234 --pin 98765432
SOFTHSM=`softhsm2-util --show-slots 2> /dev/null | grep ForFabric` || SOFTHSM=""
if [ "${SOFTHSM}" == "" ]; then
    echo "SoftHSM with ForFabric token not detected ..."
    exit 1
fi

echo "creating new slot and label..."
softhsm2-util --init-token --slot 1 --label "ForFabric1" --pin 98765432 --so-pin 987654
softhsm2-util --init-token --slot 2 --label "ForFabric2" --pin 22334455 --so-pin 987654



GO_TAGS="${GO_TAGS} ${FABRIC_SDKGO_CODELEVEL_TAG}"

GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_SDKGO_CODELEVEL_VER}/channel"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.TestRunID=${FABRIC_SDKGO_TESTRUN_ID}"
$GO_CMD test ${RACEFLAG} -cover -tags "testing ${GO_TAGS}" ${GO_TESTFLAGS} -ldflags="${GO_LDFLAGS}" ${PKGS[@]} -p 1 -timeout=40m
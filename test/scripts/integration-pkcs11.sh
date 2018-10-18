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
# FABRIC_SDKGO_CODELEVEL_TAG: Go tag that represents the fabric code target
# FABRIC_SDKGO_CODELEVEL_VER: Version that represents the fabric code target
# FABRIC_SDKGO_TESTRUN_ID: An identifier for the current run of tests.
# FABRIC_FIXTURE_VERSION: Version of fabric fixtures
# FABRIC_CRYPTOCONFIG_VERSION: Version of cryptoconfig fixture to use
# CONFIG_FILE: config file to use

set -e

GO_CMD="${GO_CMD:-go}"
GOPATH="${GOPATH:-$HOME/go}"
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-stable}"
FABRIC_SDKGO_TESTRUN_ID="${FABRIC_SDKGO_TESTRUN_ID:-${RANDOM}}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"
FABRIC_FIXTURE_VERSION="${FABRIC_FIXTURE_VERSION:-v1.2}"
CONFIG_FILE="${CONFIG_FILE:-config_test.yaml}"
TEST_LOCAL="${TEST_LOCAL:-false}"
TEST_CHANGED_ONLY="${TEST_CHANGED_ONLY:-false}"
TEST_RACE_CONDITIONS="${TEST_RACE_CONDITIONS:-true}"
SCRIPT_DIR="$(dirname "$0")"
# TODO: better default handling for FABRIC_CRYPTOCONFIG_VERSION

REPO="github.com/hyperledger/fabric-sdk-go"

source ${SCRIPT_DIR}/lib/find_packages.sh
source ${SCRIPT_DIR}/lib/docker.sh

echo "Running" $(basename "$0")

# Packages to include in test run
PKGS=($(${GO_CMD} list ${REPO}/test/integration/... 2> /dev/null | \
      grep ^${REPO}/test/integration/e2e/pkcs11 | \
      tr '\n' ' '))

# Reduce tests to changed packages.
if [ "${TEST_CHANGED_ONLY}" = true ]; then
    # findChangedFiles assumes that the working directory contains the repo; so change to the repo directory.
    PWD_ORIG=$(pwd)
    cd "${GOPATH}/src/${REPO}"
    findChangedFiles
    cd ${PWD_ORIG}

    if [[ "${CHANGED_FILES[@]}" =~ ( |^)(test/fixtures/|test/metadata/|test/scripts/|Makefile( |$)|Gopkg.lock( |$)|ci.properties( |$)) ]]; then
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
    echo "Skipping integration tests since no packages were changed"
    exit 0
fi

workingDir=${GOPATH}/src/${REPO}
declare -a PRIVATE_KEYS=($(find ${workingDir}/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION} | grep '/users/' | grep '_sk$' | tr '\n' ' '))
PRIVATE_KEYS+=($(find ${workingDir}/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION} | grep 'User1@tls.example.com' | grep '.key$' | tr '\n' ' '))

for i in "${PRIVATE_KEYS[@]}"
do
    echo "Importing key : ${i}"
    openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in ${i} -out private.p8
    pkcs11helper -action import -keyFile private.p8
    rm -rf private.p8
done

waitForCoreVMUp

echo "Code level ${FABRIC_SDKGO_CODELEVEL_TAG} (Fabric ${FABRIC_FIXTURE_VERSION})"
echo "Running integration tests ..."

GO_TAGS="$GO_TAGS ${FABRIC_SDKGO_CODELEVEL_TAG}"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_FIXTURE_VERSION}/channel"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.TestRunID=${FABRIC_SDKGO_TESTRUN_ID}"
$GO_CMD test ${RACEFLAG} -tags "${GO_TAGS}" ${GO_TESTFLAGS} -ldflags="${GO_LDFLAGS}" ${PKGS[@]} -p 1 -timeout=40m configFile=${CONFIG_FILE} testLocal=${TEST_LOCAL}

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

echo "Running" $(basename "$0")

# Packages to include in test run
PKGS=($(${GO_CMD} list ${REPO}/test/integration/... 2> /dev/null | \
      grep ^${REPO}/test/integration/e2e/pkcs11 | \
      tr '\n' ' '))

# Reduce tests to changed packages.
if [ "${TEST_CHANGED_ONLY}" = true ]; then
    # findChangedFiles assumes that the working directory contains the repo; so change to the repo directory.
    PWD=$(pwd)
    cd "${GOPATH}/src/${REPO}"
    findChangedFiles
    cd ${PWD}

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

#Add entry here below for your key to be imported into softhsm
declare -a PRIVATE_KEYS=(
	"github.com/hyperledger/fabric-sdk-go/test/fixtures/config/mutual_tls/client_sdk_go-key.pem"
	"github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/ordererOrganizations/example.com/users/Admin@example.com/msp/keystore/f4aa194b12d13d7c2b7b275a7115af5e6f728e11710716f2c754df4587891511_sk"
	"github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/ce142124e13093a3e13bc4708b0f2b26e1d4d2ea4d4cc59942790bfc0f3bcc6d_sk"
	"github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/abbe8ee0f86c227b1917d208921497603d2ff28f4ba8e902d703744c4a6fa7b7_sk"
	"github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp/keystore/371ea01078b18f3b92c1fc8233dfa8d209d882ae40aeff4defd118ba9d572a15_sk"
	"github.com/hyperledger/fabric-sdk-go/test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config/peerOrganizations/org2.example.com/users/User1@org2.example.com/msp/keystore/7777a174c9fe40ab5abe33199a4fe82f1e0a7c45715e395e73a78cc3480d0021_sk"
)

GO_SRC=/opt/gopath/src
for i in "${PRIVATE_KEYS[@]}"
do
    echo "Importing key : ${GO_SRC}/${i}"
    openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in ${GO_SRC}/${i} -out private.p8
    pkcs11helper -action import -keyFile private.p8
    rm -rf private.p8
done

echo "Code level ${FABRIC_SDKGO_CODELEVEL_TAG} (Fabric ${FABRIC_FIXTURE_VERSION})"
echo "Running integration tests ..."

GO_TAGS="$GO_TAGS ${FABRIC_SDKGO_CODELEVEL_TAG}"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_FIXTURE_VERSION}/channel"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
GO_LDFLAGS="${GO_LDFLAGS} -X github.com/hyperledger/fabric-sdk-go/test/metadata.TestRunID=${FABRIC_SDKGO_TESTRUN_ID}"
$GO_CMD test ${RACEFLAG} -tags "${GO_TAGS}" ${GO_TESTFLAGS} -ldflags="${GO_LDFLAGS}" ${PKGS[@]} -p 1 -timeout=40m configFile=${CONFIG_FILE} testLocal=${TEST_LOCAL}

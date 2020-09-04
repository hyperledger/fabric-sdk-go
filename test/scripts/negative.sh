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
# FABRIC_SDKGO_CODELEVEL_VER: Version that represents the fabric code target (primarily for fixture lookup)
# FABRIC_SDKGO_TESTRUN_ID: An identifier for the current run of tests.
# FABRIC_CRYPTOCONFIG_VERSION: Version of cryptoconfig fixture to use
# CONFIG_FILE: config file to use

set -e

GO_CMD="${GO_CMD:-go}"
FABRIC_SDKGO_CODELEVEL_TAG="${FABRIC_SDKGO_CODELEVEL_TAG:-stable}"
FABRIC_SDKGO_TESTRUN_ID="${FABRIC_SDKGO_TESTRUN_ID:-${RANDOM}}"
FABRIC_CRYPTOCONFIG_VERSION="${FABRIC_CRYPTOCONFIG_VERSION:-v1}"
CONFIG_FILE="${CONFIG_FILE:-config_test.yaml}"
TEST_CHANGED_ONLY="${TEST_CHANGED_ONLY:-false}"
TEST_RACE_CONDITIONS="${TEST_RACE_CONDITIONS:-true}"
SCRIPT_DIR="$(dirname "$0")"
CC_MODE="${CC_MODE:-lifecycle}"
# TODO: better default handling for FABRIC_CRYPTOCONFIG_VERSION

GOMOD_PATH=$(cd ${SCRIPT_DIR} && ${GO_CMD} env GOMOD)
PROJECT_MODULE=$(awk -F' ' '$1 == "module" {print $2}' ${GOMOD_PATH})
PROJECT_DIR=$(dirname ${GOMOD_PATH})

MODULE="${PROJECT_MODULE}/test/integration"
MODULE_PATH="${PROJECT_DIR}/${MODULE#${PROJECT_MODULE}}" && MODULE_PATH=${MODULE_PATH%/}

# Temporary fix for Fabric base image
unset GOCACHE

source ${SCRIPT_DIR}/lib/find_packages.sh
source ${SCRIPT_DIR}/lib/docker.sh

echo "Running" $(basename "$0")

# Packages to exclude from test run
PWD_ORIG=$(pwd)
cd "${MODULE_PATH}"
PKGS=($(${GO_CMD} list ${PROJECT_MODULE}/test/integration/negative/... 2> /dev/null | \
      tr '\n' ' '))

# Reduce tests to changed packages.
if [ "${TEST_CHANGED_ONLY}" = true ]; then
    # Find changed files across the project as these may be dependencies of the module.
    PWD_ORIG_FIND=$(pwd)
    cd "${PROJECT_DIR}"
    findChangedFiles
    cd "${PWD_ORIG_FIND}"

    if [[ "${CHANGED_FILES[@]}" =~ ( |^)(test/fixtures/|test/metadata/|test/scripts/|Makefile( |$)|go.mod( |$)|ci.properties( |$)) ]]; then
        echo "Test scripts, fixtures or metadata changed - running all tests"
    else
        findChangedPackages
        filterExcludedPackages
        appendDepPackages
        PKGS=(${DEP_PKGS[@]})
    fi
fi

RACEFLAG=""
if [ "$TEST_RACE_CONDITIONS" = true ]; then
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

waitForCoreVMUp

echo "Code level ${FABRIC_SDKGO_CODELEVEL_TAG} (Fabric ${FABRIC_FIXTURE_VERSION})"
echo "Running integration tests for expired orderer certificates ..."

GO_TAGS="${GO_TAGS} ${FABRIC_SDKGO_CODELEVEL_TAG}"
GO_LDFLAGS="${GO_LDFLAGS} -X ${PROJECT_MODULE}/test/metadata.ProjectPath=${PROJECT_DIR}"
GO_LDFLAGS="${GO_LDFLAGS} -X ${PROJECT_MODULE}/test/metadata.ChannelConfigPath=test/fixtures/fabric/${FABRIC_SDKGO_CODELEVEL_VER}/channel"
GO_LDFLAGS="${GO_LDFLAGS} -X ${PROJECT_MODULE}/test/metadata.CryptoConfigPath=test/fixtures/fabric/${FABRIC_CRYPTOCONFIG_VERSION}/crypto-config"
GO_LDFLAGS="${GO_LDFLAGS} -X ${PROJECT_MODULE}/test/metadata.TestRunID=${FABRIC_SDKGO_TESTRUN_ID}"
GO_LDFLAGS="${GO_LDFLAGS} -X ${PROJECT_MODULE}/test/metadata.CCMode=${CC_MODE}"

${GO_CMD} test ${RACEFLAG} -tags "${GO_TAGS}" ${GO_TESTFLAGS} -ldflags="${GO_LDFLAGS}" ${PKGS[@]} -p 1 -timeout=40m configFile=${CONFIG_FILE}
cd ${PWD_ORIG}

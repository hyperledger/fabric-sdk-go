#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script fetches code used in the SDK originating from other Hyperledger Fabric projects
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

set -e

UPSTREAM_PROJECT="github.com/hyperledger/fabric-ca"
UPSTREAM_BRANCH="${UPSTREAM_BRANCH:-release}"
SCRIPTS_PATH="scripts/third_party_pins/fabric-ca"
PATCHES_PATH="scripts/third_party_pins/fabric-ca/patches"

THIRDPARTY_INTERNAL_FABRIC_CA_PATH='internal/github.com/hyperledger/fabric-ca'

####
# Clone and patch packages into repo

# Clone original project into temporary directory
echo "Fetching upstream project ($UPSTREAM_PROJECT:$UPSTREAM_COMMIT) ..."
CWD=`pwd`
TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`

TMP_PROJECT_PATH=$TMP/src/$UPSTREAM_PROJECT
mkdir -p $TMP_PROJECT_PATH
cd ${TMP_PROJECT_PATH}/..

git clone https://${UPSTREAM_PROJECT}.git
cd $TMP_PROJECT_PATH
git checkout $UPSTREAM_BRANCH
git reset --hard $UPSTREAM_COMMIT

echo "Patching upstream project ..."
git am ${CWD}/${PATCHES_PATH}/*

cd $CWD

echo 'Removing current upstream project from working directory ...'
rm -Rf "${THIRDPARTY_INTERNAL_FABRIC_CA_PATH}"
mkdir -p "${THIRDPARTY_INTERNAL_FABRIC_CA_PATH}"

# fabric-ca client utils
echo "Pinning and patching fabric-ca client utils..."
declare -a CLIENT_UTILS_IMPORT_SUBSTS=(
    's/\"github.com\/hyperledger\/fabric-ca/\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric-ca/g'
    's/\"github.com\/hyperledger\/fabric\/bccsp\/factory/factory\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric-ca\/sdkpatch\/cryptosuitebridge/g'
    's/\"github.com\/hyperledger\/fabric\/bccsp/\"github.com\/hyperledger\/fabric-sdk-go\/pkg\/cryptosuite\/bccsp/g'
    '/clog.\"github.com\/cloudflare\/cfssl\/log/!s/\"github.com\/cloudflare\/cfssl\/log/log\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric-ca\/sdkpatch\/logbridge/g'
    's/\"github.com\/hyperledger\/fabric\//\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\//g'
)
INTERNAL_PATH=$THIRDPARTY_INTERNAL_FABRIC_CA_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH IMPORT_SUBSTS="${CLIENT_UTILS_IMPORT_SUBSTS[*]}" $SCRIPTS_PATH/apply_fabric_ca_client_utils.sh

# Cleanup temporary files from patch application
echo "Removing temporary files ..."
rm -Rf $TMP
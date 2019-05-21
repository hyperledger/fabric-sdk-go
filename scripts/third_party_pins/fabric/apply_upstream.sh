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

UPSTREAM_PROJECT="github.com/hyperledger/fabric"
UPSTREAM_BRANCH="${UPSTREAM_BRANCH:-release}"
SCRIPTS_PATH="scripts/third_party_pins/fabric"
PATCHES_PATH="${SCRIPTS_PATH}/patches"

THIRDPARTY_FABRIC_PATH='third_party/github.com/hyperledger/fabric'
THIRDPARTY_INTERNAL_FABRIC_PATH='internal/github.com/hyperledger/fabric'

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
rm -Rf "${THIRDPARTY_FABRIC_PATH}" "${THIRDPARTY_INTERNAL_FABRIC_PATH}"
mkdir -p "${THIRDPARTY_FABRIC_PATH}" "${THIRDPARTY_INTERNAL_FABRIC_PATH}"

# restore module definitions
git checkout "${THIRDPARTY_FABRIC_PATH}/go.mod" "${THIRDPARTY_FABRIC_PATH}/go.sum"

# Create internal utility structure
mkdir -p ${TMP_PROJECT_PATH}/internal/protoutil
cp -R ${TMP_PROJECT_PATH}/protoutil ${TMP_PROJECT_PATH}/internal/

# copy required files that are under internal into non-internal structure.
mkdir -p ${TMP_PROJECT_PATH}/sdkinternal
cp -R ${TMP_PROJECT_PATH}/internal/* ${TMP_PROJECT_PATH}/sdkinternal/

# fabric client utils
echo "Pinning and patching fabric client utils..."
declare -a CLIENT_UTILS_IMPORT_SUBSTS=(
    's/\"github.com\/hyperledger\/fabric\/internal/\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkinternal/g'
    's/[[:space:]]logging[[:space:]]\"github.com/\"github.com/g'
    's/\"github.com\/hyperledger\/fabric\/common\/flogging\/httpadmin/\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/logbridge\/httpadmin/g'
    's/\"github.com\/hyperledger\/fabric\/common\/flogging/flogging\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/logbridge/g'
    's/\"github.com\/op\/go-logging/logging\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/logbridge/g'
    's/\"github.com\/hyperledger\/fabric\/protos/\"github.com\/hyperledger\/fabric-sdk-go\/third_party\/github.com\/hyperledger\/fabric\/protos/g'
    's/\"github.com\/hyperledger\/fabric\//\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\//g'
)

INTERNAL_PATH=$THIRDPARTY_INTERNAL_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH IMPORT_SUBSTS="${CLIENT_UTILS_IMPORT_SUBSTS[*]}" $SCRIPTS_PATH/apply_fabric_client_utils.sh
INTERNAL_PATH=$THIRDPARTY_INTERNAL_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH IMPORT_SUBSTS="${CLIENT_UTILS_IMPORT_SUBSTS[*]}" $SCRIPTS_PATH/apply_fabric_common_utils.sh

# external utils
echo "Pinning and patching fabric external utils ..."
declare -a EXTERNAL_UTILS_IMPORT_SUBSTS=(
    's/\"github.com\/hyperledger\/fabric\/protoutil/\"github.com\/hyperledger\/fabric-sdk-go\/third_party\/github.com\/hyperledger\/fabric\/internal\/protoutil/g'
    's/\"github.com\/hyperledger\/fabric\//\"github.com\/hyperledger\/fabric-sdk-go\/third_party\/github.com\/hyperledger\/fabric\//g'
)
INTERNAL_PATH=$THIRDPARTY_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH IMPORT_SUBSTS="${EXTERNAL_UTILS_IMPORT_SUBSTS[*]}" $SCRIPTS_PATH/apply_fabric_external_utils.sh
INTERNAL_PATH=$THIRDPARTY_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH IMPORT_SUBSTS="${EXTERNAL_UTILS_IMPORT_SUBSTS[*]}" $SCRIPTS_PATH/apply_fabric_common_utils.sh

# protos
echo "Pinning and patching protos (fabric common)..."
declare -a PROTOS_IMPORT_SUBSTS=(
    's/\"github.com\/hyperledger\/fabric\/protoutil/\"github.com\/hyperledger\/fabric-sdk-go\/third_party\/github.com\/hyperledger\/fabric\/internal\/protoutil/g'
    's/\"github.com\/hyperledger\/fabric\//\"github.com\/hyperledger\/fabric-sdk-go\/third_party\/github.com\/hyperledger\/fabric\//g'
)
INTERNAL_PATH=$THIRDPARTY_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH IMPORT_SUBSTS="${PROTOS_IMPORT_SUBSTS[*]}" $SCRIPTS_PATH/apply_fabric_protos_common.sh

echo "Pinning and patching protos (fabric network) ..."
INTERNAL_PATH=$THIRDPARTY_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH IMPORT_SUBSTS="${PROTOS_IMPORT_SUBSTS[*]}" $SCRIPTS_PATH/apply_fabric_protos_network.sh

# Cleanup temporary files from patch application
echo "Removing temporary files ..."
rm -Rf $TMP
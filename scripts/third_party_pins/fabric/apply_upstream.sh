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

echo "UPSTREAM_BRANCH=$UPSTREAM_BRANCH"
echo "UPSTREAM_COMMIT=$UPSTREAM_COMMIT"

UPSTREAM_PROJECT="github.com/hyperledger/fabric"
UPSTREAM_BRANCH="${UPSTREAM_BRANCH:-release}"
SCRIPTS_PATH="scripts/third_party_pins/fabric"

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

cd $CWD

echo 'Removing current upstream project from working directory ...'
rm -Rf "${THIRDPARTY_FABRIC_PATH}" "${THIRDPARTY_INTERNAL_FABRIC_PATH}"
mkdir -p "${THIRDPARTY_FABRIC_PATH}" "${THIRDPARTY_INTERNAL_FABRIC_PATH}"

# Create internal utility structure
mkdir -p ${TMP_PROJECT_PATH}/internal/protoutil
cp -R ${TMP_PROJECT_PATH}/protoutil ${TMP_PROJECT_PATH}/internal/

# copy required files that are under internal into non-internal structure.
mkdir -p ${TMP_PROJECT_PATH}/sdkinternal
cp -R ${TMP_PROJECT_PATH}/internal/* ${TMP_PROJECT_PATH}/sdkinternal/

# fabric client utils
echo "Pinning and patching fabric client utils..."
INTERNAL_PATH=$THIRDPARTY_INTERNAL_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH $SCRIPTS_PATH/apply_fabric_client_utils.sh
INTERNAL_PATH=$THIRDPARTY_INTERNAL_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH $SCRIPTS_PATH/apply_fabric_common_utils.sh

# external utils
echo "Pinning and patching fabric external utils ..."
INTERNAL_PATH=$THIRDPARTY_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH $SCRIPTS_PATH/apply_fabric_external_utils.sh
INTERNAL_PATH=$THIRDPARTY_FABRIC_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH $SCRIPTS_PATH/apply_fabric_common_utils.sh

# Cleanup temporary files from patch application
echo "Removing temporary files ..."
rm -Rf $TMP
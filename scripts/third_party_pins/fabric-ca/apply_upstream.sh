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

UPSTREAM_PROJECT="github.com/hyperledger/fabric-ca"
UPSTREAM_BRANCH="${UPSTREAM_BRANCH:-release}"
SCRIPTS_PATH="scripts/third_party_pins/fabric-ca"

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

cd $CWD

echo 'Removing current upstream project from working directory ...'
rm -Rf "${THIRDPARTY_INTERNAL_FABRIC_CA_PATH}"
mkdir -p "${THIRDPARTY_INTERNAL_FABRIC_CA_PATH}"

# fabric-ca client utils
echo "Pinning and patching fabric-ca client utils..."
INTERNAL_PATH=$THIRDPARTY_INTERNAL_FABRIC_CA_PATH TMP_PROJECT_PATH=$TMP_PROJECT_PATH $SCRIPTS_PATH/apply_fabric_ca_client_utils.sh

# Cleanup temporary files from patch application
echo "Removing temporary files ..."
rm -Rf $TMP
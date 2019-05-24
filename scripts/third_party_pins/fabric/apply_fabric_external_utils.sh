#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins from Hyperledger Fabric into the SDK
# Note: This script must be adjusted as upstream makes adjustments

set -e

IMPORT_SUBSTS=($IMPORT_SUBSTS)

GOIMPORTS_CMD=goimports
GOFILTER_CMD="go run scripts/_go/src/gofilter/cmd/gofilter/gofilter.go"

# Create and populate patching directory.
declare TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
declare PATCH_PROJECT_PATH=$TMP/src/$UPSTREAM_PROJECT
cp -R ${TMP_PROJECT_PATH} ${PATCH_PROJECT_PATH}
declare TMP_PROJECT_PATH=${PATCH_PROJECT_PATH}

declare -a PKGS=(
        "internal/protoutil"
)

declare -a FILES=(
        "internal/protoutil/commonutils.go"
)

# Create directory structure for packages
for i in "${PKGS[@]}"
do
    mkdir -p $INTERNAL_PATH/${i}
done

# Apply fine-grained patching
gofilter() {
    echo "Filtering: ${FILTER_FILENAME}"
    cp ${TMP_PROJECT_PATH}/${FILTER_FILENAME} ${TMP_PROJECT_PATH}/${FILTER_FILENAME}.bak
    $GOFILTER_CMD -filename "${TMP_PROJECT_PATH}/${FILTER_FILENAME}.bak" \
        -filters "$FILTERS_ENABLED" -fn "$FILTER_FN" -gen "$FILTER_GEN" -type "$FILTER_TYPE" \
        > "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
}

echo "Filtering Go sources for allowed functions ..."
FILTERS_ENABLED="fn"

FILTER_FILENAME="internal/protoutil/commonutils.go"
FILTER_FN="MarshalOrPanic"
gofilter

echo "Filtering Go sources for allowed declarations ..."
FILTERS_ENABLED="gen,type"
FILTER_TYPE="IMPORT,CONST"
# Allow no declarations
FILTER_GEN=

FILTER_FILENAME="common/util/utils.go"
gofilter

FILTER_FILENAME="common/channelconfig/bundle.go"
FILTER_GEN="logger"
gofilter

# Apply patching
echo "Patching import paths on upstream project ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" IMPORT_SUBSTS="${IMPORT_SUBSTS[@]}" scripts/third_party_pins/common/apply_import_patching.sh

echo "Inserting modification notice ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" scripts/third_party_pins/common/apply_header_notice.sh

# Copy patched project into internal paths
echo "Copying patched upstream project into working directory ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
done

rm -Rf ${TMP_PROJECT_PATH}
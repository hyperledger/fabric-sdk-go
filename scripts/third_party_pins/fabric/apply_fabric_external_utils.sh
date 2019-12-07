#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins from Hyperledger Fabric into the SDK
# Note: This script must be adjusted as upstream makes adjustments

set -e

# Create and populate patching directory.
declare TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
declare PATCH_PROJECT_PATH=$TMP/src/$UPSTREAM_PROJECT
cp -R ${TMP_PROJECT_PATH} ${PATCH_PROJECT_PATH}
declare TMP_PROJECT_PATH=${PATCH_PROJECT_PATH}

declare -a FILES=(
        "internal/protoutil/commonutils.go"
)

# Copy patched project into internal paths and insert modification notice
echo "Copying patched upstream project into working directory and inserting modification notice ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    TARGET_BASENAME=`basename $INTERNAL_PATH/${i}`
    mkdir -p $TARGET_PATH && cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
    scripts/third_party_pins/common/apply_header_notice.sh $TARGET_PATH/$TARGET_BASENAME
done

rm -Rf ${TMP_PROJECT_PATH}
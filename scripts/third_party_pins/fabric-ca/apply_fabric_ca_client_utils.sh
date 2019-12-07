#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the BCCSP package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

set -e

# Create and populate patching directory.
declare TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
declare PATCH_PROJECT_PATH=$TMP/src/$UPSTREAM_PROJECT
cp -R ${TMP_PROJECT_PATH} ${PATCH_PROJECT_PATH}
declare TMP_PROJECT_PATH=${PATCH_PROJECT_PATH}

declare -a FILES=(
    "api/client.go"
    "api/net.go"

    "lib/attrmgr/attrmgr.go"
    "lib/client.go"
    "lib/identity.go"
    "lib/clientconfig.go"
    "lib/util.go"
    "lib/serverrevoke.go"

    "lib/streamer/jsonstreamer.go"

    "lib/tls/tls.go"

    "lib/client/credential/credential.go"
    "lib/client/credential/x509/credential.go"
    "lib/client/credential/x509/signer.go"

    "lib/common/serverresponses.go"

    "util/util.go"
    "util/csp.go"
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
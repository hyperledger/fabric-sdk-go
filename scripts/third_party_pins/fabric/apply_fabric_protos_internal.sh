#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the proto utils package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

IMPORT_SUBSTS=($IMPORT_SUBSTS)

declare -a PKGS=(
    "protos/utils"

    "protos/ledger/rwset"
    "protos/ledger/rwset/kvrwset"
    "protos/orderer"
)

declare -a FILES=(
    "protos/ledger/rwset/rwset.pb.go"
    "protos/ledger/rwset/kvrwset/kv_rwset.pb.go"

    "protos/orderer/ab.pb.go"
    "protos/orderer/configuration.pb.go"

    "protos/utils/commonutils.go"
    "protos/utils/proputils.go"
    "protos/utils/txutils.go"
)

#echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}/protos"
mkdir -p "${INTERNAL_PATH}/protos"

# Create directory structure for packages
for i in "${PKGS[@]}"
do
    mkdir -p $INTERNAL_PATH/${i}
done

# Apply global import patching
echo "Patching import paths on upstream project ..."
for i in "${FILES[@]}"
do
    for subst in "${IMPORT_SUBSTS[@]}"
    do
        sed -i '' -e $subst $TMP_PROJECT_PATH/${i}
    done
    goimports -w $TMP_PROJECT_PATH/${i}
done

# Copy patched project into internal paths
echo "Copying patched upstream project into working directory ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
done

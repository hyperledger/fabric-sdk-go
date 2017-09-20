#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins client and common package families from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

IMPORT_SUBSTS=($IMPORT_SUBSTS)

declare -a PKGS=(
    "common/crypto"
    "common/errors"
    "common/util"
    "common/metadata"
    "common/channelconfig"
    "common/cauthdsl"
    "common/ledger/util"

    "core/comm"
    "core/config"
    "core/ledger/kvledger/txmgmt/rwsetutil"
    "core/ledger/kvledger/txmgmt/version"
    "core/ledger/util"

    "events/consumer"

    "msp"
    "msp/cache"
    "msp/mgmt"
)

declare -a FILES=(
    "common/crypto/random.go"
    "common/crypto/signer.go"

    "common/util/utils.go"
    "common/metadata/metadata.go"
    "common/channelconfig/keys.go"
    "common/cauthdsl/cauthdsl_builder.go"

    "core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"
    "core/ledger/kvledger/txmgmt/version/version.go"
    "core/ledger/util/txvalidationflags.go"
    "core/ledger/util/util.go"

    "events/consumer/adapter.go"
    "events/consumer/consumer.go"

    "msp/cert.go"
    "msp/configbuilder.go"
    "msp/identities.go"
    "msp/msp.go"
    "msp/mspimpl.go"
    "msp/mspmgrimpl.go"
    "msp/cache/cache.go"
    "msp/mgmt/mgmt.go"

    "core/comm/config.go"
    "core/comm/connection.go"

    "core/config/config.go"

    "core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"

    "common/ledger/util/util.go"
)

echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}"
mkdir -p "${INTERNAL_PATH}"

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

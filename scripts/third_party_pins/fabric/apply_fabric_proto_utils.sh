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
    "protos/common"
    "protos/ledger/rwset"
    "protos/ledger/rwset/kvrwset"
    "protos/msp"
    "protos/orderer"
    "protos/peer"
)

declare -a FILES=(
    "protos/utils/blockutils.go"
    "protos/utils/commonutils.go"
    "protos/utils/proputils.go"
    "protos/utils/txutils.go"

    "protos/common/block.go"
    "protos/common/common.go"
    "protos/common/configtx.go"
    "protos/common/configuration.go"
    "protos/common/policies.go"
    "protos/common/signed_data.go"

    "protos/ledger/rwset/kvrwset/helper.go"

    "protos/msp/msp_config.go"
    "protos/msp/msp_principal.go"

    "protos/orderer/configuration.go"

    "protos/peer/admin.pb.go"
    "protos/peer/chaincodeunmarshall.go"
    "protos/peer/configuration.go"
    "protos/peer/configuration.pb.go"
    "protos/peer/init.go"
    "protos/peer/proposal.go"
    "protos/peer/proposal_response.go"
    "protos/peer/resources.go"
    "protos/peer/transaction.go"
)

#echo 'Removing current upstream project from working directory ...'
#rm -Rf "${INTERNAL_PATH}/protos"
#mkdir -p "${INTERNAL_PATH}/protos"

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

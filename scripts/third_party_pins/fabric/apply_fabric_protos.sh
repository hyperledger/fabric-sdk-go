#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the proto package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

IMPORT_SUBSTS=($IMPORT_SUBSTS)

declare -a PKGS=(
    "protos/common"
    "protos/ledger/rwset"
    "protos/ledger/rwset/kvrwset"
    "protos/msp"
    "protos/orderer"
    "protos/peer"
)

# TODO: selective removal of files
declare -a FILES=(
    "protos/common/common.pb.go"
    "protos/common/configtx.pb.go"
    "protos/common/configuration.pb.go"
    "protos/common/ledger.pb.go"
    "protos/common/policies.pb.go"

    "protos/ledger/rwset/rwset.pb.go"
    "protos/ledger/rwset/kvrwset/kv_rwset.pb.go"

    "protos/msp/identities.pb.go"
    "protos/msp/msp_config.pb.go"
    "protos/msp/msp_principal.pb.go"

    "protos/orderer/ab.pb.go"
    "protos/orderer/configuration.pb.go"
    "protos/orderer/kafka.pb.go"

    "protos/peer/admin.pb.go"
    "protos/peer/chaincode.pb.go"
    "protos/peer/chaincode_event.pb.go"
    "protos/peer/chaincode_shim.pb.go"
    "protos/peer/configuration.pb.go"
    "protos/peer/events.pb.go"
    "protos/peer/peer.pb.go"
    "protos/peer/proposal.pb.go"
    "protos/peer/proposal_response.pb.go"
    "protos/peer/query.pb.go"
    "protos/peer/resources.pb.go"
    "protos/peer/signed_cc_dep_spec.pb.go"
    "protos/peer/transaction.pb.go"
)

echo 'Removing current upstream project from working directory ...'
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

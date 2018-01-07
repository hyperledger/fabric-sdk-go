#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the proto package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

set -e

IMPORT_SUBSTS=($IMPORT_SUBSTS)

GOIMPORTS_CMD=goimports
NAMESPACE_PREFIX="sdk."

declare -a PKGS=(
    "protos/common"
    "protos/peer"

    "protos/msp"

    "protos/ledger/rwset"
    "protos/ledger/rwset/kvrwset"
    "protos/orderer"
)

declare -a FILES=(
    "protos/common/common.pb.go"
    "protos/common/configtx.pb.go"
    "protos/common/configuration.pb.go"
    "protos/common/ledger.pb.go"
    "protos/common/policies.pb.go"
    "protos/common/collection.pb.go"

    "protos/peer/chaincode.pb.go"
    "protos/peer/chaincode_event.pb.go"
    "protos/peer/configuration.pb.go"
    "protos/peer/events.pb.go"
    "protos/peer/peer.pb.go"
    "protos/peer/proposal.pb.go"
    "protos/peer/proposal_response.pb.go"
    "protos/peer/query.pb.go"
    "protos/peer/transaction.pb.go"
    "protos/peer/signed_cc_dep_spec.pb.go"

    "protos/msp/identities.pb.go"
    "protos/msp/msp_config.pb.go"
    "protos/msp/msp_principal.pb.go"

    "protos/ledger/rwset/rwset.pb.go"
    "protos/ledger/rwset/kvrwset/kv_rwset.pb.go"

    "protos/orderer/configuration.pb.go"
)

# Create directory structure for packages
for i in "${PKGS[@]}"
do
    mkdir -p $INTERNAL_PATH/${i}
done

# Apply patching
echo "Patching import paths on upstream project ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" IMPORT_SUBSTS="${IMPORT_SUBSTS[@]}" scripts/third_party_pins/common/apply_import_patching.sh

echo "Inserting modification notice ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" ALLOW_NONE_LICENSE_ID="true" scripts/third_party_pins/common/apply_header_notice.sh

echo "Changing proto registration paths to be unique"
for i in "${FILES[@]}"
do
  if [[ ${i} == "protos/common"* ]]; then
    sed -i'' -e "/proto.RegisterType/s/common/${NAMESPACE_PREFIX}common/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/common/${NAMESPACE_PREFIX}common/g" "${TMP_PROJECT_PATH}/${i}"
  fi
  if [[ ${i} == "protos/ledger/rwset/rwset.pb.go" ]]; then
    sed -i'' -e "/proto.RegisterType/s/rwset/${NAMESPACE_PREFIX}rwset/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/rwset/${NAMESPACE_PREFIX}rwset/g" "${TMP_PROJECT_PATH}/${i}"
  fi
  if [[ ${i} == "protos/ledger/rwset/kvrwset/kv_rwset.pb.go" ]]; then
    sed -i'' -e "/proto.RegisterType/s/kvrwset/${NAMESPACE_PREFIX}kvrwset/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/kvrwset/${NAMESPACE_PREFIX}kvrwset/g" "${TMP_PROJECT_PATH}/${i}"
  fi
  if [[ ${i} == "protos/msp"* ]]; then
    sed -i'' -e "/proto.RegisterType/s/msp/${NAMESPACE_PREFIX}msp/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/msp/${NAMESPACE_PREFIX}msp/g" "${TMP_PROJECT_PATH}/${i}"
  fi
  if [[ ${i} == "protos/msp/msp_principal.pb.go" ]]; then
    sed -i'' -e "/proto.RegisterType/s/common/${NAMESPACE_PREFIX}common/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/common/${NAMESPACE_PREFIX}common/g" "${TMP_PROJECT_PATH}/${i}"
  fi
  if [[ ${i} == "protos/orderer"* ]]; then
    sed -i'' -e "/proto.RegisterType/s/orderer/${NAMESPACE_PREFIX}orderer/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/orderer/${NAMESPACE_PREFIX}orderer/g" "${TMP_PROJECT_PATH}/${i}"
  fi
  if [[ ${i} == "protos/peer"* ]]; then
    sed -i'' -e "/proto.RegisterType/s/protos/${NAMESPACE_PREFIX}protos/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/protos/${NAMESPACE_PREFIX}protos/g" "${TMP_PROJECT_PATH}/${i}"
  fi
done

# Copy patched project into internal paths
echo "Copying patched upstream project into working directory ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
done

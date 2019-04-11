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
        "common/cauthdsl"
        "core/common/ccprovider"
        "core/ledger/kvledger/txmgmt/rwsetutil"
        "core/ledger/util"
)

declare -a FILES=(
        "common/cauthdsl/cauthdsl_builder.go"
        "common/cauthdsl/policyparser.go"
        "core/common/ccprovider/ccprovider.go"
        "core/common/ccprovider/cdspackage.go"
        "core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"
        "core/ledger/util/txvalidationflags.go"
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

FILTER_FILENAME="core/common/ccprovider/ccprovider.go"
FILTER_FN=Reset,String,ProtoMessage
gofilter
sed -i'' -e 's/var ccInfoCache = NewCCInfoCache(ccInfoFSProvider)//g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="core/common/ccprovider/cdspackage.go"
FILTER_FN=Reset,String,ProtoMessage
gofilter
sed -i'' -e 's/var ccInfoCache = NewCCInfoCache(ccInfoFSProvider)//g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"
FILTER_FN="NewHeight,ToProtoBytes,FromProtoBytes,toProtoMsg,TxRwSetFromProtoMsg,TxPvtRwSetFromProtoMsg,nsRwSetFromProtoMsg,nsPvtRwSetFromProtoMsg"
FILTER_FN+=",collHashedRwSetFromProtoMsg,collPvtRwSetFromProtoMsg"
gofilter

FILTER_FILENAME="core/ledger/util/txvalidationflags.go"
FILTER_FN="IsValid,IsInvalid,Flag,IsSetTo,NewTxValidationFlags,newTxValidationFlagsSetValue"
gofilter

echo "Filtering Go sources for allowed declarations ..."
FILTER_FILENAME="core/common/ccprovider/ccprovider.go"
FILTERS_ENABLED="gen,type"
FILTER_TYPE="IMPORT,CONST"
FILTER_GEN="CCPackage,ChaincodeData"
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
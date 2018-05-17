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

IMPORT_SUBSTS=($IMPORT_SUBSTS)

GOIMPORTS_CMD=goimports
GOFILTER_CMD="go run scripts/_go/src/gofilter/cmd/gofilter/gofilter.go"

declare -a PKGS=(
        "common/cauthdsl"
        "protos/utils"
        "core/common/ccprovider"
        "core/ledger/kvledger/txmgmt/rwsetutil"
        "core/ledger/util"
)

declare -a FILES=(
        "common/cauthdsl/cauthdsl_builder.go"
        "common/cauthdsl/policyparser.go"
        "protos/utils/commonutils.go"
        "protos/utils/proputils.go"
        "protos/utils/txutils.go"
        "core/common/ccprovider/ccprovider.go"
        "core/common/ccprovider/cdspackage.go"
        "core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"
        "core/ledger/util/txvalidationflags.go"
)

echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}/common"
mkdir -p "${INTERNAL_PATH}/common"

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
        -filters fn -fn "$FILTER_FN" \
        > "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
}

echo "Filtering Go sources for allowed functions ..."
FILTER_FILENAME="protos/utils/commonutils.go"
FILTER_FN="UnmarshalChannelHeader,MarshalOrPanic,UnmarshalChannelHeader,MakeChannelHeader,MakePayloadHeader,ExtractPayload"
FILTER_FN+=",Marshal,ExtractEnvelope,ExtractEnvelopeOrPanic,ExtractPayloadOrPanic"
gofilter

FILTER_FILENAME="protos/utils/proputils.go"
FILTER_FN="GetHeader,GetChaincodeProposalPayload,GetSignatureHeader,GetChaincodeHeaderExtension,GetBytesChaincodeActionPayload"
FILTER_FN+=",GetBytesTransaction,GetBytesPayload,GetHeader,GetBytesProposalResponsePayload,GetBytesProposal"
FILTER_FN+=",CreateChaincodeProposalWithTxIDNonceAndTransient"
FILTER_FN+=",GetTransaction,GetPayload,GetBytesChaincodeProposalPayload"
FILTER_FN+=",GetChaincodeActionPayload,GetProposalResponsePayload,GetChaincodeAction,GetChaincodeEvents,GetBytesChaincodeEvent,GetBytesEnvelope"
gofilter
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp\/factory"/factory "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.SHA256Opts{}/factory.GetSHA256Opts()/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="protos/utils/txutils.go"
FILTER_FN="GetBytesProposalPayloadForTx,GetEnvelopeFromBlock"
gofilter

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

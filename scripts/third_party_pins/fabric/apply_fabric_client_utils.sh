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

GOIMPORTS_CMD=goimports
GOFILTER_CMD="go run scripts/_go/cmd/gofilter/gofilter.go"

declare -a PKGS=(
    "common/crypto"
    "common/errors"
    "common/util"
    "common/channelconfig"
    "common/attrmgr"

    "sdkpatch/logbridge"

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
    "common/attrmgr/attrmgr.go"

    "common/channelconfig/applicationorg.go"
    "common/channelconfig/channel.go"
    "common/channelconfig/util.go"
    "common/channelconfig/orderer.go"
    "common/channelconfig/organization.go"
    
    "sdkpatch/logbridge/logbridge.go"

    "core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"
    "core/ledger/kvledger/txmgmt/version/version.go"
    "core/ledger/util/txvalidationflags.go"

    "events/consumer/adapter.go"
    "events/consumer/consumer.go"

    "msp/cert.go"
    "msp/configbuilder.go"
    "msp/identities.go"
    "msp/msp.go"
    "msp/mspimpl.go"
    "msp/mspmgrimpl.go"
    "msp/mspimplsetup.go"
    "msp/mspimplvalidate.go"
    "msp/cache/cache.go"
    "msp/mgmt/mgmt.go"
)

echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}"
mkdir -p "${INTERNAL_PATH}"

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

FILTER_FILENAME="common/crypto/random.go"
FILTER_FN="GetRandomNonce,GetRandomBytes"
gofilter

FILTER_FILENAME="common/crypto/signer.go"
FILTER_FN=
gofilter

FILTER_FILENAME="common/util/utils.go"
FILTER_FN="GenerateIDfromTxSHAHash,ComputeSHA256,CreateUtcTimestamp,ConcatenateBytes"
gofilter

FILTER_FILENAME="common/attrmgr/attrmgr.go"
FILTER_FN=
gofilter

FILTER_FILENAME="common/channelconfig/applicationorg.go"
FILTER_FN=
gofilter

FILTER_FILENAME="common/channelconfig/channel.go"
FILTER_FN=
gofilter

FILTER_FILENAME="common/channelconfig/util.go"
FILTER_FN=
gofilter

FILTER_FILENAME="common/channelconfig/orderer.go"
FILTER_FN=
gofilter

FILTER_FILENAME="common/channelconfig/organization.go"
FILTER_FN=
gofilter

FILTER_FILENAME="core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"
FILTER_FN="NewHeight,ToProtoBytes,toProtoMsg"
gofilter

FILTER_FILENAME="core/ledger/kvledger/txmgmt/version/version.go"
FILTER_FN=
gofilter

FILTER_FILENAME="core/ledger/util/txvalidationflags.go"
FILTER_FN="IsValid,IsInvalid,Flag,IsSetTo"
gofilter

FILTER_FILENAME="events/consumer/adapter.go"
FILTER_FN=
gofilter

FILTER_FILENAME="events/consumer/consumer.go"
FILTER_FN=
gofilter

FILTER_FILENAME="msp/cert.go"
FILTER_FN="certToPEM,isECDSASignedCert,sanitizeECDSASignedCert,certFromX509Cert,String"
gofilter

FILTER_FILENAME="msp/configbuilder.go"
FILTER_FN=
gofilter

FILTER_FILENAME="msp/identities.go"
FILTER_FN="newIdentity,newSigningIdentity,ExpiresAt,GetIdentifier,GetMSPIdentifier"
FILTER_FN+=",GetOrganizationalUnits,SatisfiesPrincipal,Serialize,Validate,Verify"
FILTER_FN+=",getHashOpt,GetPublicVersion,Sign"
gofilter

FILTER_FILENAME="msp/msp.go"
FILTER_FN=
gofilter

FILTER_FILENAME="msp/mspimpl.go"
FILTER_FN="sanitizeCert,SatisfiesPrincipal,Validate,getCertificationChainIdentifier,DeserializeIdentity,deserializeIdentityInternal"
FILTER_FN+=",getCertificationChain,getCertificationChainIdentifierFromChain,getUniqueValidationChain"
FILTER_FN+=",getUniqueValidationChain,GetDefaultSigningIdentity"
FILTER_FN+=",getCertificationChainForBCCSPIdentity,validateIdentityAgainstChain,GetIdentifier"
FILTER_FN+=",getValidationChain,GetSigningIdentity"
FILTER_FN+=",GetTLSIntermediateCerts,GetTLSRootCerts,GetType,Setup"
FILTER_FN+=",getCertFromPem,getIdentityFromConf,getSigningIdentityFromConf"
FILTER_FN+=",newBccspMsp,IsWellFormed"
gofilter
# TODO - adapt to msp/factory.go rather than changing newBccspMsp
sed -i'' -e 's/newBccspMsp/NewBccspMsp/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="msp/mspimplsetup.go"
FILTER_FN="setupCrypto,setupCAs,setupAdmins,setupCRLs,finalizeSetupCAs,setupSigningIdentity"
FILTER_FN+=",setupOUs,setupTLSCAs"
gofilter

FILTER_FILENAME="msp/mspimplvalidate.go"
FILTER_FN="validateTLSCAIdentity,validateCAIdentity,validateIdentity,validateIdentityAgainstChain"
FILTER_FN+=",validateCertAgainstChain,validateIdentityOUs,getValidityOptsForCert,isCACert"
FILTER_FN+=",getSubjectKeyIdentifierFromCert,getAuthorityKeyIdentifierFromCrl"
gofilter

FILTER_FILENAME="msp/mspmgrimpl.go"
FILTER_FN="NewMSPManager,DeserializeIdentity,GetMSPs,Setup,IsWellFormed"
gofilter

FILTER_FILENAME="msp/cache/cache.go"
FILTER_FN="New"
gofilter

FILTER_FILENAME="msp/mgmt/mgmt.go"
FILTER_FN="GetLocalMSP"
gofilter

echo "Filtering Go sources for allowed declarations ..."
FILTERS_ENABLED="gen,type"
FILTER_TYPE="IMPORT,CONST"
# Allow no declarations
FILTER_GEN=

FILTER_FILENAME="common/channelconfig/applicationorg.go"
gofilter

FILTER_FILENAME="common/channelconfig/channel.go"
gofilter

FILTER_FILENAME="common/channelconfig/util.go"
gofilter

FILTER_FILENAME="common/channelconfig/orderer.go"
gofilter

FILTER_FILENAME="common/channelconfig/organization.go"
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

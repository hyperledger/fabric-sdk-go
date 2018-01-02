#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins client and common package families from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

set -e

IMPORT_SUBSTS=($IMPORT_SUBSTS)

GOIMPORTS_CMD=goimports
GOFILTER_CMD="go run scripts/_go/cmd/gofilter/gofilter.go"

declare -a PKGS=(

    "bccsp"
    "bccsp/factory"
    "bccsp/pkcs11"
    "bccsp/signer"
    "bccsp/sw"
    "bccsp/utils"

    "common/crypto"
    "common/errors"
    "common/util"
    "common/channelconfig"
    "common/attrmgr"
    "common/ledger"

    "sdkpatch/logbridge"
    "sdkpatch/cryptosuitebridge"

    "core/common/ccprovider"

    "core/ledger/kvledger/txmgmt/rwsetutil"
    "core/ledger/kvledger/txmgmt/version"
    "core/ledger/util"

    "events/consumer"

    "msp"
    "msp/cache"
    "msp/mgmt"
)

declare -a FILES=(

    "bccsp/aesopts.go"
    "bccsp/bccsp.go"
    "bccsp/ecdsaopts.go"
    "bccsp/hashopts.go"
    "bccsp/keystore.go"
    "bccsp/opts.go"
    "bccsp/rsaopts.go"

    "bccsp/factory/factory.go"
    "bccsp/factory/nopkcs11.go"
    "bccsp/factory/opts.go"
    "bccsp/factory/pkcs11.go"
    "bccsp/factory/pkcs11factory.go"
    "bccsp/factory/swfactory.go"
    "bccsp/factory/pluginfactory.go"
    "bccsp/factory/sdkpatch_pluginfactory_noplugin.go"

    "bccsp/pkcs11/conf.go"
    "bccsp/pkcs11/ecdsa.go"
    "bccsp/pkcs11/ecdsakey.go"
    "bccsp/pkcs11/impl.go"
    "bccsp/pkcs11/pkcs11.go"

    "bccsp/signer/signer.go"

    "bccsp/sw/aes.go"
    "bccsp/sw/aeskey.go"
    "bccsp/sw/conf.go"
    "bccsp/sw/dummyks.go"
    "bccsp/sw/ecdsa.go"
    "bccsp/sw/ecdsakey.go"
    "bccsp/sw/fileks.go"
    "bccsp/sw/hash.go"
    "bccsp/sw/impl.go"
    "bccsp/sw/internals.go"
    "bccsp/sw/keyderiv.go"
    "bccsp/sw/keygen.go"
    "bccsp/sw/keyimport.go"
    "bccsp/sw/rsa.go"
    "bccsp/sw/rsakey.go"

    "bccsp/utils/errs.go"
    "bccsp/utils/io.go"
    "bccsp/utils/keys.go"
    "bccsp/utils/slice.go"
    "bccsp/utils/x509.go"
    "bccsp/utils/ecdsa.go"
    "common/crypto/random.go"
    "common/crypto/signer.go"

    "common/util/utils.go"
    "common/attrmgr/attrmgr.go"

    "common/channelconfig/applicationorg.go"
    "common/channelconfig/channel.go"
    "common/channelconfig/util.go"
    "common/channelconfig/orderer.go"

    "common/channelconfig/organization.go"

    "common/ledger/ledger_interface.go"
    "core/common/ccprovider/ccprovider.go"
    
    "sdkpatch/logbridge/logbridge.go"
    "sdkpatch/cryptosuitebridge/cryptosuitebridge.go"

    "core/ledger/ledger_interface.go"
    "core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"
    "core/ledger/kvledger/txmgmt/version/version.go"
    "core/ledger/util/txvalidationflags.go"


    "events/consumer/adapter.go"
    "events/consumer/consumer.go"

    "msp/factory.go"
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

FILTER_FILENAME="bccsp/signer/signer.go"
FILTER_FN=New,Public,Sign
gofilter
sed -i'' -e '/"crypto"/ a \
"github.com\/hyperledger\/fabric-sdk-go\/api\/apicryptosuite"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/apicryptosuite.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/apicryptosuite.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="common/crypto/random.go"
FILTER_FN="GetRandomNonce,GetRandomBytes"
gofilter

FILTER_FILENAME="common/crypto/signer.go"
FILTER_FN=
gofilter

FILTER_FILENAME="common/util/utils.go"
FILTER_FN="GenerateIDfromTxSHAHash,ComputeSHA256,CreateUtcTimestamp,ConcatenateBytes"
gofilter
sed -i'' -e 's/&bccsp.SHA256Opts{}/factory.GetSHA256Opts()/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp\/factory"/factory "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

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
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp"/bccsp "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

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
FILTER_FN="IsValid,IsInvalid,Flag,IsSetTo,NewTxValidationFlags"
gofilter


FILTER_FILENAME="core/common/ccprovider/ccprovider.go"
FILTER_FN=Reset,String,ProtoMessage
gofilter
sed -i'' -e 's/var ccInfoCache = NewCCInfoCache(ccInfoFSProvider)//g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="events/consumer/adapter.go"
FILTER_FN=
gofilter

FILTER_FILENAME="events/consumer/consumer.go"
FILTER_FN=
gofilter

FILTER_FILENAME="msp/factory.go"
FILTER_FN=
gofilter

FILTER_FILENAME="msp/cert.go"
FILTER_FN="certToPEM,isECDSASignedCert,sanitizeECDSASignedCert,certFromX509Cert,String"
gofilter
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp\/utils"/utils "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="msp/configbuilder.go"
FILTER_FN=
gofilter

FILTER_FILENAME="msp/identities.go"
FILTER_FN="newIdentity,newSigningIdentity,ExpiresAt,GetIdentifier,GetMSPIdentifier"
FILTER_FN+=",GetOrganizationalUnits,SatisfiesPrincipal,Serialize,Validate,Verify"
FILTER_FN+=",getHashOpt,GetPublicVersion,Sign"
gofilter
sed -i'' -e '/"encoding\/hex/ a\
"github.com\/hyperledger\/fabric-sdk-go\/api\/apicryptosuite"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp"/bccsp "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/apicryptosuite.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.HashOpts/apicryptosuite.HashOpts/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

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
FILTER_FN+=",newBccspMsp,IsWellFormed,GetVersion"
FILTER_FN+=",hasOURole,hasOURoleInternal"
gofilter
# TODO - adapt to msp/factory.go rather than changing newBccspMsp
sed -i'' -e 's/newBccspMsp/NewBccspMsp/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/NewBccspMsp(version MSPVersion)/NewBccspMsp(version MSPVersion, cryptoSuite apicryptosuite.CryptoSuite)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp := factory.GetDefault()//g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/theMsp.bccsp = bccsp/theMsp.bccsp = cryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp\/factory"/factory "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/apicryptosuite.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key,/apicryptosuite.Key,/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.GetHashOpt/factory.GetHashOpt/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/signer.New(/factory.NewCspSigner(/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.ECDSAPrivateKeyImportOpts{Temporary: true}/factory.GetECDSAPrivateKeyImportOpts(true)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/&bccsp.X509PublicKeyImportOpts{Temporary: true}/factory.GetX509PublicKeyImportOpts(true)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

#sed -i'' -e 's/signer.New(msp.bccsp, privKey)/signer.New(cryptosuite.GetSuite(msp.bccsp), cryptosuite.GetKey(privKey))/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="msp/mspimplsetup.go"
FILTER_FN="setupCrypto,setupCAs,setupAdmins,setupCRLs,finalizeSetupCAs,setupSigningIdentity"
FILTER_FN+=",setupOUs,setupTLSCAs,setupV1,setupV11,getCertifiersIdentifier"
FILTER_FN+=",preSetupV1,postSetupV1,setupNodeOUs,postSetupV11"
gofilter
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp"/bccsp "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="msp/mspimplvalidate.go"
FILTER_FN="validateTLSCAIdentity,validateCAIdentity,validateIdentity,validateIdentityAgainstChain"
FILTER_FN+=",validateCertAgainstChain,validateIdentityOUs,getValidityOptsForCert,isCACert"
FILTER_FN+=",getSubjectKeyIdentifierFromCert,getAuthorityKeyIdentifierFromCrl"
FILTER_FN+=",validateIdentityOUsV1,validateIdentityOUsV11"
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

# adjust bccsp pkcs11 build tags
FILTER_FILENAME="bccsp/factory/pkcs11factory.go"
sed -i'' -e 's/\+build !nopkcs11/\+build pkcs11/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="bccsp/factory/pkcs11.go"
sed -i'' -e 's/\+build !nopkcs11/\+build pkcs11/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="bccsp/factory/nopkcs11.go"
sed -i'' -e 's/\+build nopkcs11/\+build !pkcs11/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

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

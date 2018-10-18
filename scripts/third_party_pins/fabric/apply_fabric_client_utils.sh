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
GOFILTER_CMD="go run scripts/_go/src/gofilter/cmd/gofilter/gofilter.go"

declare -a PKGS=(

    "bccsp"
    "bccsp/factory/sw"
    "bccsp/factory/pkcs11"
    "bccsp/factory/plugin"
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
    "common/metrics"

    "sdkpatch/logbridge"
    "sdkpatch/cryptosuitebridge"
    "sdkpatch/cachebridge"

    "core/ledger/kvledger/txmgmt/version"
    "core/ledger/util"

    "msp"
    "msp/cache"

    "discovery/client"
    "gossip/util"
)

declare -a FILES=(

    "bccsp/aesopts.go"
    "bccsp/bccsp.go"
    "bccsp/ecdsaopts.go"
    "bccsp/hashopts.go"
    "bccsp/keystore.go"
    "bccsp/opts.go"
    "bccsp/rsaopts.go"

    "bccsp/factory/pkcs11/pkcs11factory.go"
    "bccsp/factory/sw/swfactory.go"
    "bccsp/factory/plugin/pluginfactory.go"

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
    "bccsp/sw/new.go"
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

    "common/metrics/server.go"
    "common/metrics/tally_provider.go"
    "common/metrics/types.go"

    "sdkpatch/logbridge/logbridge.go"
    "sdkpatch/cryptosuitebridge/cryptosuitebridge.go"
    "sdkpatch/cachebridge/cache.go"

    "core/ledger/ledger_interface.go"
    "core/ledger/kvledger/txmgmt/version/version.go"

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

    "discovery/client/api.go"
    "discovery/client/client.go"
    "discovery/client/selection.go"
    "discovery/client/signer.go"

    "gossip/util/misc.go"
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

echo "Modifying go source files"
FILTER_FILENAME="bccsp/pkcs11/impl.go"
sed -i'' -e '/"math\/big"/a sdkp11 "github.com\/hyperledger\/fabric-sdk-go\/pkg\/core\/cryptosuite\/common\/pkcs11"' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "lib := opts.Library" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..12}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i "$START_LINE i \/\/Load PKCS11 context handle" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
let "START_LINE+=1"
sed -i "$START_LINE i pkcs11Ctx, err := sdkp11.LoadContextAndLogin(opts.Library, opts.Pin, opts.Label)" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
let "START_LINE+=1"
sed -i "$START_LINE i if err != nil {return nil, errors.Wrapf(err, \"Failed initializing PKCS11 context\")}" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
let "START_LINE+=1"
sed -i "$START_LINE i csp := &impl{BCCSP: swCSP, conf: conf, ks: keyStore, softVerify: opts.SoftVerify, pkcs11Ctx: pkcs11Ctx}" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

START_LINE=`grep -n "type impl struct {" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
let "START_LINE+=6"
for i in {1..5}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done

sed -i "$START_LINE i pkcs11Ctx *sdkp11.ContextHandle" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="bccsp/pkcs11/pkcs11.go"
sed -i'' -e '/"github.com\/hyperledger"/a "time"/' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/"math\/big"/a "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cachebridge"' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/"math\/big"/a sdkp11 "github.com\/hyperledger\/fabric-sdk-go\/pkg\/core\/cryptosuite\/common\/pkcs11"' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/session = s/a cachebridge.ClearSession(fmt.Sprintf("%d", session))' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= findKeyPairFromSKI(p11lib,/= csp.pkcs11Ctx.FindKeyPairFromSKI(/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/func findKeyPairFromSKI(mod/func (csp \*impl) findKeyPairFromSKI(mod/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
START_LINE=`grep -n "func (csp \*impl) findKeyPairFromSKI(mod" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
let "START_LINE+=1"
for i in {1..27}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done
sed -i'' -e '/func (csp \*impl) findKeyPairFromSKI(mod/a return cachebridge.GetKeyPairFromSessionSKI(&cachebridge.KeyPairCacheKey{Mod: mod, Session: session, SKI: ski, KeyType: keyType})' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/func (csp \*impl) findKeyPairFromSKI(mod/i \
func timeTrack(start time.Time, msg string) {\
	elapsed := time.Since(start)\
	logger.Debugf("%s took %s", msg, elapsed)\
}\

' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/\"go.uber.org\/zap\/zapcore/logging\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/logbridge/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/zapcore.DebugLevel/logging.DEBUG/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

START_LINE=`grep -n "func loadLib(lib, pin, label string)" "${TMP_PROJECT_PATH}/${FILTER_FILENAME}" | head -n 1 | awk -F':' '{print $1}'`
for i in {1..97}
do
    sed -i'' -e ${START_LINE}'d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
done

sed -i'' -e 's/p11lib := csp.ctx//g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/session := csp.getSession()/session := csp.pkcs11Ctx.GetSession()/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/defer csp.returnSession(session)/defer csp.pkcs11Ctx.ReturnSession(session)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= ecPoint(p11lib/= ecPoint(csp.pkcs11Ctx/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.GenerateKeyPair(/= csp.pkcs11Ctx.GenerateKeyPair(/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.SetAttributeValue(/= csp.pkcs11Ctx.SetAttributeValue(/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.GetAttributeValue(session/= csp.pkcs11Ctx.GetAttributeValue(session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.CreateObject(/= csp.pkcs11Ctx.CreateObject(/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.VerifyInit(session/= csp.pkcs11Ctx.VerifyInit(session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.Verify(session/= csp.pkcs11Ctx.Verify(session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.SignInit(session/= csp.pkcs11Ctx.SignInit(session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.Sign(session/= csp.pkcs11Ctx.Sign(session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.CopyObject(session/= csp.pkcs11Ctx.CopyObject(session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/= p11lib.DestroyObject(session/= csp.pkcs11Ctx.DestroyObject(session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/listAttrs(p11lib, session/listAttrs(csp.pkcs11Ctx, session/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/func listAttrs(p11lib \*pkcs11.Ctx,/func listAttrs(p11lib \*sdkp11.ContextHandle,/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/func ecPoint(p11lib \*pkcs11.Ctx,/func ecPoint(p11lib \*sdkp11.ContextHandle,/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/attr, err := csp.pkcs11Ctx.GetAttributeValue(session, key, template)/attr, err := p11lib.GetAttributeValue(session, key, template)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/attr, err := csp.pkcs11Ctx.GetAttributeValue(session, obj, template)/attr, err := p11lib.GetAttributeValue(session, obj, template)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/privateKey, err := csp.pkcs11Ctx.FindKeyPairFromSKI/a defer timeTrack(time.Now(), fmt.Sprintf("signing [session: %d]", session))' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

echo "Filtering Go sources for allowed functions ..."
FILTERS_ENABLED="fn"

FILTER_FILENAME="bccsp/signer/signer.go"
FILTER_FN=New,Public,Sign
gofilter
sed -i'' -e '/"crypto"/ a \
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

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

FILTER_FILENAME="core/ledger/kvledger/txmgmt/version/version.go"
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
FILTER_FN+=",getHashOpt,GetPublicVersion,Sign,Anonymous"
gofilter
sed -i'' -e '/"encoding\/hex/ a\
"github.com\/hyperledger\/fabric-sdk-go\/pkg\/common\/providers\/core"\
' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp"/bccsp "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key/core.Key/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.HashOpts/core.HashOpts/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/\"go.uber.org\/zap\/zapcore/logging\"github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/logbridge/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/zapcore.DebugLevel/logging.DEBUG/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

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
FILTER_FN+=",hasOURole,hasOURoleInternal,collectPrincipals,satisfiesPrincipalInternalV13,satisfiesPrincipalInternalPreV13"
gofilter
# TODO - adapt to msp/factory.go rather than changing newBccspMsp
sed -i'' -e 's/newBccspMsp/NewBccspMsp/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/NewBccspMsp(version MSPVersion)/NewBccspMsp(version MSPVersion, cryptoSuite core.CryptoSuite)/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp := factory.GetDefault()//g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/theMsp.bccsp = bccsp/theMsp.bccsp = cryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/"github.com\/hyperledger\/fabric\/bccsp\/factory"/factory "github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric\/sdkpatch\/cryptosuitebridge"/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.BCCSP/core.CryptoSuite/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/bccsp.Key,/core.Key,/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
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

FILTER_FILENAME="gossip/util/misc.go"
FILTER_FN="GetRandomIndices,RandomInt,IndexInSlice,numbericEqual,RandomUInt64"
gofilter

# Split BCCSP factory into subpackages
mkdir ${TMP_PROJECT_PATH}/bccsp/factory/sw
mkdir ${TMP_PROJECT_PATH}/bccsp/factory/pkcs11
mkdir ${TMP_PROJECT_PATH}/bccsp/factory/plugin
mv ${TMP_PROJECT_PATH}/bccsp/factory/swfactory.go ${TMP_PROJECT_PATH}/bccsp/factory/sw/swfactory.go
mv ${TMP_PROJECT_PATH}/bccsp/factory/pkcs11factory.go ${TMP_PROJECT_PATH}/bccsp/factory/pkcs11/pkcs11factory.go
mv ${TMP_PROJECT_PATH}/bccsp/factory/pluginfactory.go ${TMP_PROJECT_PATH}/bccsp/factory/plugin/pluginfactory.go

FILTER_FILENAME="bccsp/factory/pkcs11/pkcs11factory.go"
sed -i'' -e '/\+build pkcs11/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/package factory/package pkcs11/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/config \*FactoryOpts/p11Opts \*pkcs11.PKCS11Opts/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/if config == nil || config.Pkcs11Opts == nil/if p11Opts == nil/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/p11Opts := config.Pkcs11Opts/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="bccsp/factory/sw/swfactory.go"
sed -i'' -e 's/package factory/package sw/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/config \*FactoryOpts/swOpts \*SwOpts/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/if config == nil || config.SwOpts == nil/if swOpts == nil/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e '/swOpts := config.SwOpts/d' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

FILTER_FILENAME="bccsp/factory/plugin/pluginfactory.go"
sed -i'' -e 's/package factory/package plugin/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/config \*FactoryOpts/pluginOpts \*PluginOpts/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/if config == nil || config.PluginOpts == nil/if pluginOpts == nil/g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"
sed -i'' -e 's/config.PluginOpts./pluginOpts./g' "${TMP_PROJECT_PATH}/${FILTER_FILENAME}"

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

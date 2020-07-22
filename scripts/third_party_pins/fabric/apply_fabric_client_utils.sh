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

# Create and populate patching directory.
declare TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`
declare PATCH_PROJECT_PATH=$TMP/src/$UPSTREAM_PROJECT
cp -R ${TMP_PROJECT_PATH} ${PATCH_PROJECT_PATH}
declare TMP_PROJECT_PATH=${PATCH_PROJECT_PATH}

# Split BCCSP factory into subpackages
mkdir ${TMP_PROJECT_PATH}/bccsp/factory/sw
mkdir ${TMP_PROJECT_PATH}/bccsp/factory/pkcs11
mv ${TMP_PROJECT_PATH}/bccsp/factory/swfactory.go ${TMP_PROJECT_PATH}/bccsp/factory/sw/swfactory.go
mv ${TMP_PROJECT_PATH}/bccsp/factory/pkcs11factory.go ${TMP_PROJECT_PATH}/bccsp/factory/pkcs11/pkcs11factory.go

declare -a FILES=(

    "bccsp/aesopts.go"
    "bccsp/bccsp.go"
    "bccsp/ecdsaopts.go"
    "bccsp/hashopts.go"
    "bccsp/keystore.go"
    "bccsp/opts.go"

    "bccsp/factory/pkcs11/pkcs11factory.go"
    "bccsp/factory/sw/swfactory.go"

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
    "bccsp/sw/inmemoryks.go"
    "bccsp/sw/internals.go"
    "bccsp/sw/keys.go"
    "bccsp/sw/keyderiv.go"
    "bccsp/sw/keygen.go"
    "bccsp/sw/keyimport.go"
    "bccsp/sw/new.go"

    "bccsp/utils/ecdsa.go"

    "common/configtx/configtx.go"


    "common/capabilities/application.go"
    "common/capabilities/capabilities.go"
    "common/capabilities/channel.go"
    "common/capabilities/orderer.go"

    "common/genesis/genesis.go"

    "common/crypto/random.go"

    "common/channelconfig/application.go"
    "common/channelconfig/consortium.go"
    "common/channelconfig/consortiums.go"
    "common/channelconfig/applicationorg.go"
    "common/channelconfig/channel.go"
    "common/channelconfig/util.go"
    "common/channelconfig/orderer.go"
    "common/channelconfig/organization.go"
    "common/channelconfig/msp.go"
    "common/channelconfig/api.go"
    "common/channelconfig/standardvalues.go"
    "common/channelconfig/acls.go"
    "common/channelconfig/bundle.go"

    "common/policies/policy.go"
    "common/policies/util.go"
    "common/policies/implicitmetaparser.go"

    "common/ledger/ledger_interface.go"

    "common/metrics/disabled/provider.go"
    "common/metrics/internal/namer/namer.go"
    "common/metrics/prometheus/provider.go"
    "common/metrics/provider.go"
    "common/metrics/statsd/goruntime/collector.go"
    "common/metrics/statsd/goruntime/metrics.go"
    "common/metrics/statsd/provider.go"

    "core/middleware/chain.go"
    "core/middleware/request_id.go"
    "core/middleware/require_cert.go"
    "core/operations/metrics.go"
    "core/operations/system.go"
    "core/operations/tls.go"

    "common/util/utils.go"

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
    "msp/cache/second_chance.go"

    "protoutil/blockutils.go"
    "protoutil/commonutils.go"
    "protoutil/configtxutils.go"
    "protoutil/proputils.go"
    "protoutil/signeddata.go"
    "protoutil/txutils.go"
    "protoutil/configtxutils.go"
    "protoutil/unmarshalers.go"

    "discovery/client/api.go"
    "discovery/client/client.go"
    "discovery/client/selection.go"
    "discovery/client/signer.go"
    "discovery/protoext/response.go"
    "discovery/protoext/querytype.go"

    "gossip/protoext/signing.go"
    "gossip/protoext/message.go"
    "gossip/protoext/stringers.go"
    "gossip/util/misc.go"

    "sdkinternal/configtxgen/encoder/encoder.go"
    "sdkinternal/configtxgen/genesisconfig/config.go"
    "sdkinternal/configtxlator/update/update.go"
    "sdkinternal/configtxlator/update/update.go"

    "sdkinternal/pkg/identity/identity.go"

    "sdkinternal/pkg/comm/config.go"
    "sdkinternal/pkg/txflags/validation_flags.go"

    "core/chaincode/platforms/golang/list.go"
    "core/chaincode/platforms/golang/platform.go"
    "core/chaincode/platforms/java/platform.go"
    "core/chaincode/platforms/node/platform.go"
    "core/chaincode/platforms/util/writer.go"
    "core/chaincode/persistence/persistence.go"
    "core/chaincode/platforms/platforms.go"
    "core/chaincode/persistence/chaincode_package.go"
    "sdkinternal/ccmetadata/validators.go"
    "sdkinternal/peer/packaging/platforms.go"
)

# Copy patched project into internal paths and insert modification notice
echo "Copying patched upstream project into working directory and inserting modification notice ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    TARGET_BASENAME=`basename $INTERNAL_PATH/${i}`
    mkdir -p $TARGET_PATH && cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
    scripts/third_party_pins/common/apply_header_notice.sh $TARGET_PATH/$TARGET_BASENAME
done

rm -Rf ${TMP_PROJECT_PATH}
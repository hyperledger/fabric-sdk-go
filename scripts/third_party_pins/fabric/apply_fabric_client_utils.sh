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
    "common/flogging"
    "common/util"
    "common/metadata"
    "common/channelconfig"
    "common/cauthdsl"
    "common/configtx"
    "common/configtx/api"
    "common/policies"
    "common/ledger"
    "common/ledger/util"

    "core/comm"
    "core/config"
    "core/ledger"
    "core/ledger/kvledger/txmgmt/rwsetutil"
    "core/ledger/kvledger/txmgmt/version"
    "core/ledger/util"

    "events/consumer"

    "msp"
    "msp/cache"
    "msp/mgmt"
)

# TODO: selective removal of files
declare -a FILES=(
    "common/crypto/random.go"
    "common/crypto/signer.go"

    "common/errors/codes.go"
    "common/errors/errors.go"

    "common/flogging/grpclogger.go"
    "common/flogging/logging.go"

    "common/util/utils.go"

    "common/metadata/metadata.go"

    "common/channelconfig/api.go"
    "common/channelconfig/application.go"
    "common/channelconfig/application_util.go"
    "common/channelconfig/applicationorg.go"
    "common/channelconfig/bundle.go"
    "common/channelconfig/bundlesource.go"
    "common/channelconfig/channel.go"
    "common/channelconfig/channel_util.go"
    "common/channelconfig/consortium.go"
    "common/channelconfig/consortiums.go"
    "common/channelconfig/logsanitychecks.go"
    "common/channelconfig/msp.go"
    "common/channelconfig/msp_util.go"
    "common/channelconfig/orderer.go"
    "common/channelconfig/orderer_util.go"
    "common/channelconfig/organization.go"
    "common/channelconfig/standardvalues.go"
    "common/channelconfig/template.go"

    "common/cauthdsl/cauthdsl.go"
    "common/cauthdsl/cauthdsl_builder.go"
    "common/cauthdsl/policy.go"
    "common/cauthdsl/policy_util.go"
    "common/cauthdsl/policyparser.go"

    "common/configtx/compare.go"
    "common/configtx/configmap.go"
    "common/configtx/manager.go"
    "common/configtx/template.go"
    "common/configtx/update.go"
    "common/configtx/util.go"
    "common/configtx/api/api.go"

    "common/policies/implicitmeta.go"
    "common/policies/implicitmeta_util.go"
    "common/policies/policy.go"

    "core/ledger/kvledger/txmgmt/rwsetutil/query_results_helper.go"
    "core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go"
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
    "msp/mgmt/deserializer.go"
    "msp/mgmt/mgmt.go"
    "msp/mgmt/principal.go"

    "core/comm/config.go"
    "core/comm/connection.go"
    "core/comm/creds.go"
    "core/comm/producer.go"
    "core/comm/server.go"

    "core/config/config.go"

    "core/ledger/ledger_interface.go"

    "core/ledger/kvledger/txmgmt/rwsetutil/query_results_helper.go"
    "core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go"
    "core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go"

    "common/ledger/ledger_interface.go"

    "common/ledger/util/ioutil.go"
    "common/ledger/util/protobuf_util.go"
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

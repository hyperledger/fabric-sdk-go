#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the BCCSP package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

IMPORT_SUBSTS=($IMPORT_SUBSTS)

declare -a PKGS=(
    "api"
    "lib"
    "lib/tls"
    "lib/tcert"
    "lib/spi"
    "util"
)

declare -a FILES=(
    "api/client.go"
    "api/net.go"

    "lib/client.go"
    "lib/identity.go"
    "lib/signer.go"
    "lib/clientconfig.go"
    "lib/util.go"
    "lib/serverstruct.go"

    "lib/tls/tls.go"

    "lib/tcert/api.go"
    "lib/tcert/util.go"
    "lib/tcert/tcert.go"
    "lib/tcert/keytree.go"

    "lib/spi/affiliation.go"
    "lib/spi/userregistry.go"

    "util/util.go"
    "util/args.go"
    "util/csp.go"
    "util/struct.go"
    "util/flag.go"
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

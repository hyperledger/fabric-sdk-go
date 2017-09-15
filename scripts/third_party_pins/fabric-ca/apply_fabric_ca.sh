#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script fetches code used in the SDK originating from other Hyperledger Fabric projects
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

UPSTREAM_PROJECT="github.com/hyperledger/fabric-ca"
INTERNAL_PATH="internal/${UPSTREAM_PROJECT}"
UPSTREAM_BRANCH="release"
PATCHES_PATH="scripts/third_party_pins/fabric-ca/patches"

# TODO - in a future CS, fabric imports need to have imports rewritten.
#IMPORT_FABRIC_SUBST='s/github.com\/hyperledger\/fabric/github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric/g'
IMPORT_FABRICCA_SUBST='s/github.com\/hyperledger\/fabric-ca/github.com\/hyperledger\/fabric-sdk-go\/internal\/github.com\/hyperledger\/fabric-ca/g'

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

####
# Clone and patch packages into repo

# Cleanup existing internal packages
echo 'Removing current upstream project from working directory ...'
rm -Rf $INTERNAL_PATH
mkdir -p $INTERNAL_PATH

# Create directory structure for packages
for i in "${PKGS[@]}"
do
    mkdir -p $INTERNAL_PATH/${i}
done

# Clone original project into temporary directory
echo "Fetching upstream project ($UPSTREAM_PROJECT:$UPSTREAM_COMMIT) ..."
CWD=`pwd`
TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir'`

TMP_PROJECT_PATH=$TMP/src/$UPSTREAM_PROJECT
mkdir -p $TMP_PROJECT_PATH
cd ${TMP_PROJECT_PATH}/..

git clone https://${UPSTREAM_PROJECT}.git
cd $TMP_PROJECT_PATH
git checkout $UPSTREAM_BRANCH
git reset --hard $UPSTREAM_COMMIT

echo "Patching upstream project ..."
git am ${CWD}/${PATCHES_PATH}/*

cd $CWD

# Apply global import patching
echo "Patching import paths on upstream project ..."
for i in "${FILES[@]}"
do
    # TODO Patch fabric paths (in upcoming change set)
    #sed -i '' -e $IMPORT_FABRIC_SUBST $INTERNAL_PATH/${i}
    sed -i '' -e $IMPORT_FABRICCA_SUBST $TMP_PROJECT_PATH/${i}
    goimports -w $TMP_PROJECT_PATH/${i}
done

# Copy patched project into internal paths
echo "Copying patched upstream project into working directory ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
done

# Cleanup temporary files from patch application
echo "Removing temporary files ..."
rm -Rf $TMP
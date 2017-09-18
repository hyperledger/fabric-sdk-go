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
    "bccsp"
    "bccsp/factory"
    "bccsp/pkcs11"
    "bccsp/signer"
    "bccsp/sw"
    "bccsp/utils"
)

declare -a FILES=(
    "bccsp/aesopts.go"
    "bccsp/bccsp.go"
    "bccsp/ecdsaopts.go"
    "bccsp/hashopts.go"
    "bccsp/keystore.go"
    "bccsp/opts.go"
    "bccsp/rsaopts.go"
    "bccsp/rsaopts.go"
    "bccsp/rsaopts.go"

    "bccsp/factory/factory.go"
    "bccsp/factory/nopkcs11.go"
    "bccsp/factory/opts.go"
    "bccsp/factory/pkcs11.go"
    "bccsp/factory/pkcs11factory.go"
    "bccsp/factory/swfactory.go"

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
)

echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}/bccsp"
mkdir -p "${INTERNAL_PATH}/bccsp"

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

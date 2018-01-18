#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script pins the proto utils package family from Hyperledger Fabric into the SDK
# These files are checked into internal paths.
# Note: This script must be adjusted as upstream makes adjustments

set -e

IMPORT_SUBSTS=($IMPORT_SUBSTS)

GOIMPORTS_CMD=goimports
GOFILTER_CMD="go run scripts/_go/src/gofilter/cmd/gofilter/gofilter.go"
NAMESPACE_PREFIX="sdk."

declare -a PKGS=(
    "protos/utils"
    "protos/orderer"
)

declare -a FILES=(
    "protos/orderer/ab.pb.go"
)

declare -a NPBFILES=(
)

declare -a PBFILES=(
    "protos/orderer/ab.pb.go"
)

#echo 'Removing current upstream project from working directory ...'
rm -Rf "${INTERNAL_PATH}/protos"
mkdir -p "${INTERNAL_PATH}/protos"

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

# Apply patching
echo "Patching import paths on upstream project ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${FILES[@]}" IMPORT_SUBSTS="${IMPORT_SUBSTS[@]}" scripts/third_party_pins/common/apply_import_patching.sh

echo "Inserting modification notice ..."
WORKING_DIR=$TMP_PROJECT_PATH FILES="${NPBFILES[@]}" scripts/third_party_pins/common/apply_header_notice.sh
WORKING_DIR=$TMP_PROJECT_PATH FILES="${PBFILES[@]}" ALLOW_NONE_LICENSE_ID="true" scripts/third_party_pins/common/apply_header_notice.sh

echo "Changing proto registration paths to be unique"
for i in "${FILES[@]}"
do
  if [[ ${i} == "protos/orderer"* ]]; then
    sed -i'' -e "/proto.RegisterType/s/orderer/${NAMESPACE_PREFIX}orderer/g" "${TMP_PROJECT_PATH}/${i}"
    sed -i'' -e "/proto.RegisterEnum/s/orderer/${NAMESPACE_PREFIX}orderer/g" "${TMP_PROJECT_PATH}/${i}"
  fi
done

# Copy patched project into internal paths
echo "Copying patched upstream project into working directory ..."
for i in "${FILES[@]}"
do
    TARGET_PATH=`dirname $INTERNAL_PATH/${i}`
    cp $TMP_PROJECT_PATH/${i} $TARGET_PATH
done

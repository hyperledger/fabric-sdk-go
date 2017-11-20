#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

FILES=($FILES)
NOTICE=$'Notice: This file has been modified for Hyperledger Fabric SDK Go usage.\nPlease review third_party pinning scripts and patches for more details.'
SPDX_LICENSE_ID="SPDX-License-Identifier: Apache-2.0"
OLD_APACHE_LICENSE_ID="http://www.apache.org/licenses/LICENSE-2.0"
NONE_LICENSE_ID="::NONE::"
ALLOW_NONE_LICENSE_ID="${ALLOW_NONE_LICENSE_ID:-false}"

if [ -z $WORKING_DIR ]; then
    WORKING_DIR=`pwd`
fi

for i in "${FILES[@]}"
do
    if APPLIED=`scripts/third_party_pins/common/insert_header_notice.sh ${WORKING_DIR}/${i} "$NOTICE" "$SPDX_LICENSE_ID"`; then
        echo "$APPLIED" > ${WORKING_DIR}/${i}
    elif APPLIED=`scripts/third_party_pins/common/insert_header_notice.sh ${WORKING_DIR}/${i} "$NOTICE" "$OLD_APACHE_LICENSE_ID"`; then
        echo "$APPLIED" > ${WORKING_DIR}/${i}
    elif [ "$ALLOW_NONE_LICENSE_ID" == "true" ] && APPLIED=`scripts/third_party_pins/common/insert_header_notice.sh ${WORKING_DIR}/${i} "$NOTICE" "$NONE_LICENSE_ID"`; then
        echo "$APPLIED" > ${WORKING_DIR}/${i}
    else
        echo "Failed to apply notice to ${WORKING_DIR}/${i}"
        exit 1    
    fi
done
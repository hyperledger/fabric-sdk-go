#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

FILES=($FILES)
IMPORT_SUBSTS=($IMPORT_SUBSTS)
GOIMPORTS_CMD=goimports

if [ -z $WORKING_DIR ]; then
    WORKING_DIR=`pwd`
fi

for i in "${FILES[@]}"
do
    for subst in "${IMPORT_SUBSTS[@]}"
    do
        sed -i'' -e $subst $WORKING_DIR/${i}
    done
    $GOIMPORTS_CMD -w $WORKING_DIR/${i}
done
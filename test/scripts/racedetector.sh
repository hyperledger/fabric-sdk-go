#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

SCRIPT_PATH=$(dirname "$0")

export GOTESTFLAGS="$GOTESTFLAGS -race -v"
export TEST_MASSIVE_ORDERER_COUNT=2000
$SCRIPT_PATH/unit.sh
$SCRIPT_PATH/integration.sh
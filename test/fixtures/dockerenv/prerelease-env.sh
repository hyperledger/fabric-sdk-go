#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This file contains environment overrides to enable testing
# against the latest pre-release target.
export FABRIC_FIXTURE_VERSION="v1.4"
export FABRIC_CRYPTOCONFIG_VERSION="v1"

export FABRIC_CA_FIXTURE_TAG="1.4.0-rc2"
export FABRIC_ORDERER_FIXTURE_TAG="1.4.0-rc2"
export FABRIC_PEER_FIXTURE_TAG="1.4.0-rc2"
export FABRIC_BUILDER_FIXTURE_TAG="1.4.0-rc2"

#!/usr/bin/env bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This variable can be altered to test against prebuilt fabric and fabric-ca images
# If set to false, CI will build images from scratch for the commit levels specified below
export USE_PREBUILT_IMAGES=true

# versions of fabric to build (if USE_PREBUILT_IMAGES=false)
export FABRIC_COMMIT=
export FABRIC_CA_COMMIT=

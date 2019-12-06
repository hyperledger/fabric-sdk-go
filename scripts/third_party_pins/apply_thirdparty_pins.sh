#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script fetches code originating from other upstream projects
# These files are checked into internal paths.

set -e

if output=$(git status --porcelain) && [ -z "$output" ]
then
  echo "Working directory clean, proceeding with upstream patching"
else
  echo "ERROR: git status must be clean before applying upstream patches"
  exit 1
fi

export UPSTREAM_COMMIT="${THIRDPARTY_FABRIC_COMMIT}"
export UPSTREAM_BRANCH="${THIRDPARTY_FABRIC_BRANCH}"
scripts/third_party_pins/fabric/apply_upstream.sh

export UPSTREAM_COMMIT="${THIRDPARTY_FABRIC_CA_COMMIT}"
export UPSTREAM_BRANCH="${THIRDPARTY_FABRIC_CA_BRANCH}"
scripts/third_party_pins/fabric-ca/apply_upstream.sh

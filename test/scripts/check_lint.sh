#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script runs Go linting and vetting tools

set -e
LINT_CHANGED_ONLY="${LINT_CHANGED_ONLY:-false}"
GO_CMD="${GO_CMD:-go}"
SCRIPT_DIR="$(dirname "$0")"

REPO="github.com/hyperledger/fabric-sdk-go"

echo "Running" $(basename "$0")

source ${SCRIPT_DIR}/lib/find_packages.sh
source ${SCRIPT_DIR}/lib/linter.sh

# Find all packages that should be linted.
declare -a PKG_SRC=(
    "./pkg"
    "./test"
)
declare PKG_EXCLUDE=""
findPackages

# Reduce Linter checks to changed packages.
if [ "$LINT_CHANGED_ONLY" = true ]; then
    findChangedLinterPkgs
fi

if [ ${#PKGS[@]} -eq 0 ]; then
    echo "Skipping tests since no packages were changed"
    exit 0
fi

runLinter

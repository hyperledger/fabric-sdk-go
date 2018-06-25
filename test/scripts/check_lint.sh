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
GOMETALINT_CMD="gometalinter"
SCRIPT_DIR="$(dirname "$0")"

REPO="github.com/hyperledger/fabric-sdk-go"

echo "Running" $(basename "$0")

source ${SCRIPT_DIR}/lib/find_packages.sh

# Find all packages that should be linted.
declare -a PKG_SRC=(
"./pkg"
"./test"
)
findPackages

# Reduce Linter checks to changed packages.
if [ "$LINT_CHANGED_ONLY" = true ]; then
    findChangedFiles
    findChangedPackages
    filterExcludedPackages
    appendDepPackages
    PKGS=(${DEP_PKGS[@]})
fi

packagesToDirs

if [ ${#DIRS[@]} -eq 0 ]; then
    echo "Skipping linter since no packages were changed"
    exit 0
fi

if [ "$LINT_CHANGED_ONLY" = true ]; then
    echo "Changed directories to lint: ${DIRS[@]}"
fi

echo "Running metalinters..."
$GOMETALINT_CMD --config=./gometalinter.json "${DIRS[@]}"
echo "Metalinters finished successfully"

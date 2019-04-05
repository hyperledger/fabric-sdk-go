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
GOPATH="${GOPATH:-$HOME/go}"
SCRIPT_DIR="$(dirname "$0")"
CONFIG_DIR=$(pwd)
PKG_ROOT="${PKG_ROOT:-./}"

echo "Running" $(basename "$0") "(${MODULE} ${PKG_ROOT})"

source ${SCRIPT_DIR}/lib/find_packages.sh
source ${SCRIPT_DIR}/lib/linter.sh

# Find all packages that should be linted.
PWD_ORIG=$(pwd)
cd "${GOPATH}/src/${MODULE}"
declare -a PKG_SRC=(
    "${PKG_ROOT}"
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
cd ${PWD_ORIG}

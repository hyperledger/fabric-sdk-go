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
PKG_ROOT="${PKG_ROOT:-./}"

PROJECT_MODULE=$(awk -F' ' '$1 == "module" {print $2}' < $(${GO_CMD} env GOMOD))
PROJECT_DIR=$(dirname $(${GO_CMD} env GOMOD))

CONFIG_DIR=${PROJECT_DIR}

echo "Running" $(basename "$0") "(${MODULE} ${PKG_ROOT})"

source ${SCRIPT_DIR}/lib/find_packages.sh
source ${SCRIPT_DIR}/lib/linter.sh

# Find all packages that should be linted.
PWD_ORIG=$(pwd)
cd "${PROJECT_DIR}/${MODULE#${PROJECT_MODULE}}"
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

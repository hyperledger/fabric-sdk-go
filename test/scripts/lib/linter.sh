#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

GOLANGCI_LINT_PKG="github.com/golangci/golangci-lint/cmd/golangci-lint"
GOLANGCI_LINT_VER="${GOLANGCI_LINT_VER:-}"

if [ "${GOLANGCI_LINT_VER}" != "" ]; then
    GOLANGCI_LINT_AT_VER="@${GOLANGCI_LINT_VER}"
fi

GOLANGCI_LINT_CMD="${GOLANGCI_LINT_CMD:-gobin -run ${GOLANGCI_LINT_PKG}${GOLANGCI_LINT_AT_VER}}"

function runLinter {
    packagesToDirs

    echo "Directories to lint: ${DIRS[@]}"

    echo "Running golangci-lint${GOLANGCI_LINT_AT_VER} ..."
    ${GOLANGCI_LINT_CMD} run -c "${CONFIG_DIR}/golangci.yml" "${DIRS[@]}"
    echo "golangci-lint finished successfully"
}

function findChangedLinterPkgs {
    findChangedFiles
    declare matcher='( |^)(test/fixtures/|test/metadata/|test/scripts/|Makefile( |$)|go.mod( |$))'
    if [[ "${CHANGED_FILES[@]}" =~ ${matcher} ]]; then
        echo "Test scripts, fixtures or metadata changed - running all tests"
    else
        findChangedPackages
        filterExcludedPackages
        appendDepPackages
        PKGS=(${DEP_PKGS[@]})
    fi
}
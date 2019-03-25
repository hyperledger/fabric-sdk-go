#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

GOLANGCI_LINT_CMD="${GOLANGCI_LINT_CMD:-golangci-lint}"

function runLinter {
    packagesToDirs

    echo "Directories to lint: ${DIRS[@]}"

    echo "Running golangci-lint..."
    ${GOLANGCI_LINT_CMD} run -c "./golangci.yml" "${DIRS[@]}"
    echo "golangci-lint finished successfully"
}

function findChangedLinterPkgs {
    findChangedFiles
    declare matcher='( |^)(test/fixtures/|test/metadata/|test/scripts/|Makefile( |$)|Gopkg.lock( |$))'
    if [[ "${CHANGED_FILES[@]}" =~ ${matcher} ]]; then
        echo "Test scripts, fixtures or metadata changed - running all tests"
    else
        findChangedPackages
        filterExcludedPackages
        appendDepPackages
        PKGS=(${DEP_PKGS[@]})
    fi
}
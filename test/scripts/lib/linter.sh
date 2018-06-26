#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

GOMETALINT_CMD="gometalinter"

function runLinter {
    packagesToDirs

    echo "Directories to lint: ${DIRS[@]}"

    echo "Running metalinters..."
    ${GOMETALINT_CMD} --config=./gometalinter.json "${DIRS[@]}"
    echo "Metalinters finished successfully"
}

function findChangedLinterPkgs {
    findChangedFiles

    if [[ "${CHANGED_FILES[@]}" =~ ( |^)(test/fixtures/|test/metadata/|test/scripts/|Makefile( |$)) ]]; then
        echo "Test scripts, fixtures or metadata changed - running all tests"
    else
        findChangedPackages
        filterExcludedPackages
        appendDepPackages
        PKGS=(${DEP_PKGS[@]})
    fi
}
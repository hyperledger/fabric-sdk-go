#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

GO_METALINTER_CMD="${GO_METALINTER_CMD:-gometalinter}"

function runLinter {
    packagesToDirs

    echo "Directories to lint: ${DIRS[@]}"

    echo "Running metalinters..."
    ${GO_METALINTER_CMD} --config=./gometalinter.json "${DIRS[@]}"
    echo "Metalinters finished successfully"
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
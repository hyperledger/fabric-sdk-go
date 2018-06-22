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
REPO="github.com/hyperledger/fabric-sdk-go"
GOMETALINT_CMD=gometalinter

function findPackages {
    # Find all packages that should be linted.
    declare -a src=(
    "./pkg"
    "./test"
    )

    PKGS=()
    for i in "${src[@]}"
    do
       PKG_LIST=`$GO_CMD list $i/... 2> /dev/null`
       while read -r line; do
          PKGS+=("$line")
       done <<< "$PKG_LIST"
    done
}

function findChangedPackages {
    # Determine which directories have changes.
    CHANGED=$(git diff --name-only --diff-filter=ACMRTUXB HEAD)

    if [[ "$CHANGED" != "" ]]; then
        CHANGED+=$'\n'
    fi

    LAST_COMMITS=($(git log -2 --pretty=format:"%h"))
    CHANGED+=$(git diff-tree --no-commit-id --name-only --diff-filter=ACMRTUXB -r ${LAST_COMMITS[1]} ${LAST_COMMITS[0]})

    CHANGED_PKGS=()
    while read -r line; do
        if [ "$line" != "" ]; then
            DIR=`dirname $line`
            if [ "$DIR" = "." ]; then
                CHANGED_PKGS+=("$REPO")
            else
                CHANGED_PKGS+=("$REPO/$DIR")
            fi
        fi
    done <<< "$CHANGED"
    CHANGED_PKGS=($(printf "%s\n" "${CHANGED_PKGS[@]}" | sort -u | tr '\n' ' '))
}

function filterExcludedPackages {
    FILTERED_PKGS=()

    for pkg in "${PKGS[@]}"
    do
        for i in "${CHANGED_PKGS[@]}"
        do
            if [ "$pkg" = "$i" ]; then
              FILTERED_PKGS+=("$pkg")
            fi
        done
    done

    PKGS=("${FILTERED_PKGS[@]}")
}

# packagesToDirs convert packages to directories
function packagesToDirs {
    DIRS=()
    for i in "${PKGS[@]}"
    do
        lintdir=${i#$REPO/}
        DIRS+=($lintdir)
    done
}

findPackages

# Reduce Linter checks to changed packages.
if [ "$LINT_CHANGED_ONLY" = true ]; then
    findChangedPackages
    filterExcludedPackages
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

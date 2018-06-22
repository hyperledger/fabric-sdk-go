#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

function findPackages {
    PKGS=()
    for i in "${PKG_SRC[@]}"
    do
       PKG_LIST=`${GO_CMD} list $i/... 2> /dev/null`
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

function appendDepPackages {
    DEP_PKGS=()
    for pkg in "${PKGS[@]}"
    do
        DEP_PKGS+=(${pkg})
        DEP_PKGS+=($(${GO_CMD} list -f '{{.Deps}}' ${pkg} | tr ' ' '\n' | grep "^${REPO}" | \
            grep -v "^${REPO}/vendor/" | \
            grep -v "^${REPO}/internal/github.com/" | \
            grep -v "^${REPO}/third_party/github.com/" | \
            tr '\n' ' '))
    done

    PKGS=($(printf "%s\n" "${DEP_PKGS[@]}" | sort -u | tr '\n' ' '))
}

# packagesToDirs convert packages to directories
function packagesToDirs {
    DIRS=()
    for i in "${PKGS[@]}"
    do
        pkgdir=${i#$REPO/}
        DIRS+=(${pkgdir})
    done
}
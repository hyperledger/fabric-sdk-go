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
       PKGS+=($(${GO_CMD} list ${i}/... 2> /dev/null | tr '\n' ' '))
    done
}

function findChangedFiles {
    CHANGED_FILES=($(git diff --name-only --diff-filter=ACMRTUXB HEAD | tr '\n' ' '))
    declare REMOTE_REF=$(git log -1 --pretty=format:"%d" | grep '[(].*\/' | wc -l)

    # If CHANGED_FILES is empty then there is no working directory changes: fallback to last two commits.
    # Else if REMOTE_REF=0 then working copy commits are even with remote: only use the working copy changes.
    # Otherwise assume that the change is amending the previous commit: use both last two commit and working copy changes.
    if [[ ${#CHANGED_FILES[@]} -eq 0 ]] || [[ "${REMOTE_REF}" -eq 0 ]]; then
        if [[ ! -z "${CHANGED_FILES}" ]]; then
            echo "Examining last commit and working directory changes"
        else
            echo "Examining last commit changes"
        fi

        declare -a LAST_COMMITS=($(git log -2 --pretty=format:"%h"))
        CHANGED_FILES+=($(git diff-tree --no-commit-id --name-only --diff-filter=ACMRTUXB -r ${LAST_COMMITS[1]} ${LAST_COMMITS[0]} | tr '\n' ' '))
    else
        echo "Examining working directory changes"
    fi
}

function findChangedPackages {
    CHANGED_PKGS=()
    for file in "${CHANGED_FILES[@]}"
    do
        # TODO filter out non GO/YAML/JSON files
        # TODO handle vendor

        if [ "$file" != "" ]; then
            DIR=`dirname $file`
            if [ "$DIR" = "." ]; then
                CHANGED_PKG+=("$REPO")
            else
                CHANGED_PKGS+=("$REPO/$DIR")
            fi
        fi
    done

    # Make result unique and filter out non-Go "packages".
    CHANGED_PKGS=($(printf "%s\n" "${CHANGED_PKGS[@]}" | sort -u | xargs ${GO_CMD} list 2> /dev/null | tr '\n' ' '))
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

    FILTERED_PKGS=("${FILTERED_PKGS[@]}")
}

function calcDepPackages {
    echo "Calculating package dependencies ..."

    for pkg in "${PKGS[@]}"
    do
        declare testImports=$(${GO_CMD} list -f '{{.TestImports}}' ${pkg} | tr -d '[]' | tr ' ' '\n' | \
            grep "^${REPO}" | \
            grep -v "^${REPO}/vendor/" | \
            grep -v "^${REPO}/internal/github.com/" | \
            grep -v "^${REPO}/third_party/github.com/" | \
            tr '\n' ' ')

        declare pkgDeps=$(${GO_CMD} list -f '{{.Deps}}' ${pkg} ${testImports} | tr -d '[]')

        declare val=$(${GO_CMD} list ${testImports} ${pkgDeps} | tr '\n' ' ')

        export PKGDEPS__${pkg//[-\.\/]/_}="${val}"
    done
}

function appendDepPackages {
    calcDepPackages

    DEP_PKGS=("${FILTERED_PKGS[@]}")

    # For each changed package, see if a candidate package uses that changed package as a dependency.
    # If so, include that candidate package.
    for cpkg in "${CHANGED_PKGS[@]}"
    do
        for pkg in "${PKGS[@]}"
        do
            declare key="PKGDEPS__${pkg//[-\.\/]/_}"
            declare -a pkgDeps=(${!key})

            for i in "${pkgDeps[@]}"
            do
                if [ "${cpkg}" = "${i}" ]; then
                  DEP_PKGS+=("${pkg}")
                fi
            done
        done
    done

    DEP_PKGS=($(printf "%s\n" "${DEP_PKGS[@]}" | sort -u | tr '\n' ' '))
}

# packagesToDirs convert packages to directories
function packagesToDirs {
    DIRS=()
    for i in "${PKGS[@]}"
    do
        declare -a pkgDir=${i#$REPO/}
        DIRS+=(${pkgDir})
    done
}
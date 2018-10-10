#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script populates the vendor folder.

set -e

GO_DEP_CMD="${GO_DEP_CMD:-dep}"
LASTRUN_INFO_FILENAME="populate-vendor.txt"
VENDOR_TREE_FILENAME="vendor-tree.txt"
SCRIPT_REVISION=$(git log -1 --pretty=format:"%h" test/scripts/populate-vendor.sh)
LOCK_REVISION=$(git log -1 --pretty=format:"%h" Gopkg.lock)
DATE=$(date +"%m-%d-%Y")

CACHE_PATH=""
function setCachePath {
    declare envOS=$(uname -s)
    declare pkgDir="fabric-sdk-go"

    if [ ${envOS} = 'Darwin' ]; then
        CACHE_PATH="${HOME}/Library/Caches/${pkgDir}"
    else
        CACHE_PATH="${HOME}/.cache/${pkgDir}"
    fi
}

# recordCacheResult writes the date and revision of successful script runs, to preempt unnecessary installs.
function recordCacheResult {
    declare VENDOR_TREE=$(ls -R vendor)

    mkdir -p ${CACHE_PATH}
    echo ${SCRIPT_REVISION} ${LOCK_REVISION} ${DATE} > "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"
    echo "${VENDOR_TREE}" > "${CACHE_PATH}/${VENDOR_TREE_FILENAME}"
}

function isScriptCurrent {
    declare filesModified=$(git diff --name-only --diff-filter=ACMRTUXBD HEAD | tr '\n' ' ' | xargs)
    declare matcher='( |^)(test/scripts/populate-vendor.sh|Gopkg.lock)( |$)'
    if [[ "${filesModified}" =~ ${matcher} ]]; then
        echo "Vendor script or Gopkg.lock modified - will need to repopulate vendor"
        return 1
    fi
}

function isVendorCurrent {
    if [ ! -d "vendor" ]; then
        echo "Vendor directory does not exist - will need to populate vendor"
        return 1
    fi

    if [ ! -f "${CACHE_PATH}/${VENDOR_TREE_FILENAME}" ]; then
        echo "Vendor cache doesn't exist - populating vendor"
        return 1
    fi

    declare VENDOR_TREE=$(ls -R vendor)
    declare cachedVendorTree=$(< "${CACHE_PATH}/${VENDOR_TREE_FILENAME}")
    if [ "${VENDOR_TREE}" != "${cachedVendorTree}" ]; then
        echo "Vendor directory modified - will need to repopulate vendor"
        return 1
    fi
}

function isLastPopulateCurrent {

    if [ -f "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}" ]; then
        declare -a lastScriptUsage=($(< "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"))
        echo "Last populated vendor on ${lastScriptUsage[2]} using revision ${lastScriptUsage[0]} with Gopkg.lock revision ${lastScriptUsage[1]}"

        if [ "${lastScriptUsage[0]}" = "${SCRIPT_REVISION}" ] && [ "${lastScriptUsage[1]}" = "${LOCK_REVISION}" ]; then
            return 0
        fi
    fi

    return 1
}

function isPopulateCurrent {
    if ! isScriptCurrent || ! isVendorCurrent || ! isLastPopulateCurrent; then
        return 1
    fi
}


function isForceMode {
    if [ "${BASH_ARGV[0]}" != "-f" ]; then
        return 1
    fi
}

function populateVendor {
    echo "Populating vendor ..."
	${GO_DEP_CMD} ensure -vendor-only

    echo "Populating dockerd vendor ..."
    declare chaincodedPath="scripts/_go/src/chaincoded"
    rm -Rf ${chaincodedPath}/vendor/
    mkdir -p ${chaincodedPath}/vendor/github.com/hyperledger/fabric
    git clone --branch release-1.2 --depth=1 https://github.com/hyperledger/fabric.git ${chaincodedPath}/vendor/github.com/hyperledger/fabric

}

setCachePath

if ! isPopulateCurrent || isForceMode; then
    populateVendor
    recordCacheResult
else
    echo "No need to populate vendor"
fi
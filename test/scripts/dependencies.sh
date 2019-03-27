#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# This script installs dependencies for testing tools
# Environment variables that affect this script:
# GO_DEP_COMMIT: Tag or commit level of the go dep tool to install

set -e

GO_CMD="${GO_CMD:-go}"
GO_DEP_CMD="${GO_DEP_CMD:-dep}"
GO_DEP_REPO="github.com/golang/dep"
GOLANGCI_LINT_CMD="${GOLANGCI_LINT_CMD:-golangci-lint}"
GOPATH="${GOPATH:-${HOME}/go}"

DEPEND_SCRIPT_REVISION=$(git log -1 --pretty=format:"%h" test/scripts/dependencies.sh)
DATE=$(date +"%m-%d-%Y")

LASTRUN_INFO_FILENAME="dependencies.txt"

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
    mkdir -p ${CACHE_PATH}
    echo ${DEPEND_SCRIPT_REVISION} ${DATE} > "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"
}

function installGoDep {
    declare repo=$1
    declare revision=$2

    installGoPkg "${repo}" "${revision}" "/cmd/dep" "dep"
}

function installGolangCiLint {
    declare repo="github.com/golangci/golangci-lint/cmd/golangci-lint"
    declare revision="v1.15.0"

    declare pkg="github.com/golangci/golangci-lint/cmd/golangci-lint"

    installGoPkg "${repo}" "${revision}" "" "golangci-lint"
    cp -f ${BUILD_TMP}/bin/* ${GOPATH}/bin/
    rm -Rf ${GOPATH}/src/${pkg}
    mkdir -p ${GOPATH}/src/${pkg}
    cp -Rf ${BUILD_TMP}/src/${repo}/* ${GOPATH}/src/${pkg}/
}

function installGoPkg {
    declare repo=$1
    declare revision=$2
    declare pkgPath=$3
    shift 3
    declare -a cmds=$@

    echo "Installing ${repo}@${revision} to $GOPATH/bin ..."

    GOPATH=${BUILD_TMP} go get -d ${repo}
    tag=$(cd ${BUILD_TMP}/src/${repo} && git tag -l --sort=-version:refname | head -n 1 | grep "${revision}" || true)
    if [ ! -z "${tag}" ]; then
        revision=${tag}
        echo "  using tag ${revision}"
    fi
    (cd ${BUILD_TMP}/src/${repo} && git reset --hard ${revision})
    GOPATH=${BUILD_TMP} GOBIN=${BUILD_TMP}/bin go install -i ${repo}/${pkgPath}

    mkdir -p ${GOPATH}/bin
    for cmd in ${cmds[@]}
    do
        echo "Copying ${cmd} to ${GOPATH}/bin"
        cp -f ${BUILD_TMP}/bin/${cmd} ${GOPATH}/bin/
    done
}

function isScriptCurrent {
    declare filesModified=$(git diff --name-only --diff-filter=ACMRTUXBD HEAD | tr '\n' ' ' | xargs)
    declare matcher='( |^)(test/scripts/dependencies.sh)( |$)'
    if [[ "${filesModified}" =~ ${matcher} ]]; then
        echo "Dependencies script modified - will need to install dependencies"
        return 1
    fi
}

function isLastInstallCurrent {
    if [ -f "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}" ]; then
        declare -a lastScriptUsage=($(< "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"))
        echo "Last installed dependencies on ${lastScriptUsage[1]} with revision ${lastScriptUsage[0]}"

        if [ "${lastScriptUsage[0]}" = "${DEPEND_SCRIPT_REVISION}" ] && [ "${lastScriptUsage[1]}" = "${DATE}" ]; then
            return 0
        fi
    fi

    return 1
}

function isDependencyCurrent {
    if ! isScriptCurrent || ! isLastInstallCurrent; then
        return 1
    fi
}

# areImagesInstalled checks that the docker images are installed.
function areImagesInstalled {
    declare imgCount=$(docker images | grep fabsdkgo-softhsm2 | wc -l)

    if [ ${imgCount} -eq 0 ]; then
        echo "fabsdkgo-softhsm2 docker image does not exist"
        return 1
    fi
}

# isDependenciesInstalled checks that Go tools are installed and help the user if they are missing
function isDependenciesInstalled {
    declare printMsgs=$1
    declare -a msgs=()

    # Check that Go tools are installed and help the user if they are missing
    type gocov >/dev/null 2>&1 || msgs+=("gocov is not installed (go get -u github.com/axw/gocov/...)")
    type gocov-xml >/dev/null 2>&1 || msgs+=("gocov-xml is not installed (go get -u github.com/AlekSi/gocov-xml)")
    type mockgen >/dev/null 2>&1 || msgs+=("mockgen is not installed (go get -u github.com/golang/mock/mockgen)")
    type ${GO_DEP_CMD} >/dev/null 2>&1 || msgs+=("dep is not installed (go get -u github.com/golang/dep/cmd/dep)")
    type ${GOLANGCI_LINT_CMD} >/dev/null 2>&1 || msgs+=("golangci-lint is not installed (go get -u ${GOLANGCI_LINT_CMD})")

    if [ ${#msgs[@]} -gt 0 ]; then
        if [ ${printMsgs} = true ]; then
            echo >& 2 $(echo ${msgs[@]} | tr ' ' '\n')
        fi

        return 1
    fi
}

function installDependencies {
    echo "Installing dependencies ..."
    rm -f "${CACHE_PATH}/${LASTRUN_INFO_FILENAME}"

    BUILD_TMP=`mktemp -d 2>/dev/null || mktemp -d -t 'fabricsdkgo'`
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/axw/gocov/...
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/AlekSi/gocov-xml
    GOPATH=${BUILD_TMP} ${GO_CMD} get -u github.com/golang/mock/mockgen

    installGolangCiLint

    # Install specific version of go dep (particularly for CI)
    if [ -n "${GO_DEP_COMMIT}" ]; then
        installGoDep ${GO_DEP_REPO} ${GO_DEP_COMMIT}
    fi

    rm -Rf ${BUILD_TMP}
}

function buildDockerImages {
    echo "Creating docker images used by tests ..."
    make build-softhsm2-image

    # chaincoded is currently able to intercept the docker calls without need for forwarding.
    # (as long as this remains true, socat is not needed).
    #make build-socat-image
}

function isForceMode {
    if [ "${BASH_ARGV[0]}" != "-f" ]; then
        return 1
    fi
}

function isCheckOnlyMode {
    if [ "${BASH_ARGV[0]}" != "-c" ]; then
        return 1
    fi
}

setCachePath

if isCheckOnlyMode; then
    if ! isDependenciesInstalled true; then
        echo "Missing tool dependency. You can fix by running make depend or installing the tool listed above."
        exit 1
    fi

    if ! areImagesInstalled; then
        echo "Missing docker image dependency. You can fix by running make depend or make build-softhsm2-image."
        exit 1
    fi
    exit 0
fi

if ! isDependencyCurrent || ! isDependenciesInstalled false || ! areImagesInstalled || isForceMode; then
    installDependencies
    buildDockerImages
    recordCacheResult
else
    echo "No need to install dependencies"
fi

